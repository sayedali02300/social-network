package ws

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"log"
	"net/http"
	"socialnetwork/backend/internal/models"
	"socialnetwork/backend/internal/repository"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		// Keep permissive in local development; tighten this for production.
		return true
	},
}

const sessionCookieName = "session_id"

const TypePostSend = "post_send"
const TypePostEvent = "post_event"
const TypePostDelete = "delete_post"

const TypeCommentSend = "comment_send"
const TypeCommentEvent = "comment_event"
const TypeCommentDelete = "delete_comment"

const (
	maxInboundMessageBytes = 64 * 1024
	rateLimitWindow        = 10 * time.Second
	rateLimitMaxMessages   = 30
)

// NotifyFunc is a callback used to create a persistent notification for a user.
// userID is the recipient, actorID is the one who triggered it.
type NotifyFunc func(userID, actorID, notifType, payload string)

// NewHandler upgrades authenticated HTTP connections and registers them with the hub.
func NewHandler(hub *Hub, sessions *SessionStore, chats *ChatStore, PostRepo *repository.PostRepository, CommentRepo *repository.CommentRepository, notify NotifyFunc) http.HandlerFunc {
	limiter := newFixedWindowRateLimiter(rateLimitWindow, rateLimitMaxMessages)

	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie == nil || cookie.Value == "" {
			http.Error(w, "unauthorized: missing session cookie", http.StatusUnauthorized)
			return
		}

		userID, err := sessions.ValidateSession(cookie.Value, r)
		if err != nil {
			if err == sql.ErrNoRows || strings.Contains(err.Error(), "session revoked") {
				http.Error(w, "unauthorized: invalid or expired session", http.StatusUnauthorized)
				return
			}
			log.Printf("ws session lookup failed: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade failed: %v", err)
			return
		}
		conn.SetReadLimit(maxInboundMessageBytes)

		client := &Client{
			UserID: userID,
			Conn:   conn,
		}
		hub.Register(client)
		defer func() {
			hub.Unregister(client)
			_ = conn.Close()
		}()

		for {
			msgType, payload, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("ws closed user=%s err=%v", userID, err)
					return
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Printf("ws read unexpected close user=%s err=%v", userID, err)
					return
				}
				log.Printf("ws read failed user=%s err=%v", userID, err)
				return
			}

			if msgType != websocket.TextMessage {
				if err := writeProtocolAndClose(client, "UNSUPPORTED_MESSAGE_TYPE", "only text websocket messages are supported", websocket.CloseUnsupportedData); err != nil {
					log.Printf("ws write failed user=%s err=%v", userID, err)
				}
				return
			}

			if !limiter.Allow(userID, time.Now().UTC()) {
				if err := writeProtocolAndClose(client, "RATE_LIMIT_EXCEEDED", "too many websocket messages, retry later", websocket.ClosePolicyViolation); err != nil {
					log.Printf("ws write failed user=%s err=%v", userID, err)
				}
				return
			}

			if err := routeIncoming(client, payload, hub, chats, PostRepo, CommentRepo, notify); err != nil {
				var wsErr *wsMessageError
				if errors.As(err, &wsErr) {
					if writeErr := writeProtocolError(client, wsErr.Code, wsErr.Message); writeErr != nil {
						log.Printf("ws write failed user=%s err=%v", userID, writeErr)
						return
					}
					continue
				}

				log.Printf("ws route failed user=%s err=%v", userID, err)
				if writeErr := writeProtocolError(client, "INTERNAL_ERROR", "internal websocket handler error"); writeErr != nil {
					log.Printf("ws write failed user=%s err=%v", userID, writeErr)
					return
				}
			}
		}
	}
}

type wsMessageError struct {
	Code    string
	Message string
}

func (e *wsMessageError) Error() string {
	return e.Code + ": " + e.Message
}

func routeIncoming(client *Client, raw []byte, hub *Hub, chats *ChatStore, PostRepo *repository.PostRepository, CommentRepo *repository.CommentRepository, notify NotifyFunc) error {
	env, err := decodeIncoming(raw)
	if err != nil {
		return &wsMessageError{
			Code:    "INVALID_ENVELOPE",
			Message: err.Error(),
		}
	}

	switch env.Type {
	case TypePing:
		return writeAck(client, TypePing)
	case TypePrivateMessageSend:
		p, err := parsePrivateMessagePayload(env.Payload)
		if err != nil {
			return &wsMessageError{
				Code:    "INVALID_PRIVATE_PAYLOAD",
				Message: err.Error(),
			}
		}
		allowed, err := chats.CanSendPrivateMessage(client.UserID, p.ToUserID)
		if err != nil {
			return err
		}
		if !allowed {
			return &wsMessageError{
				Code:    "PRIVATE_MESSAGE_NOT_ALLOWED",
				Message: "private message requires follow relationship",
			}
		}

		messageEvent, err := chats.SavePrivateMessage(client.UserID, p.ToUserID, p.Content)
		if err != nil {
			return err
		}

		hub.EmitToUser(p.ToUserID, OutgoingEnvelope{
			Type:    TypePrivateMessage,
			Payload: messageEvent,
		})
		hub.EmitToUser(client.UserID, OutgoingEnvelope{
			Type:    TypePrivateMessage,
			Payload: messageEvent,
		})
		if notify != nil {
			muted, _ := chats.IsMuted(p.ToUserID, "private", client.UserID)
			if !muted {
				notify(p.ToUserID, client.UserID, "new_private_message", client.UserID)
			}
		}
		return writeAck(client, TypePrivateMessageSend)
	case TypeGroupMessageSend:
		p, err := parseGroupMessagePayload(env.Payload)
		if err != nil {
			return &wsMessageError{
				Code:    "INVALID_GROUP_PAYLOAD",
				Message: err.Error(),
			}
		}

		allowed, err := chats.CanSendGroupMessage(client.UserID, p.GroupID)
		if err != nil {
			return err
		}
		if !allowed {
			return &wsMessageError{
				Code:    "GROUP_MESSAGE_NOT_ALLOWED",
				Message: "group message requires group membership",
			}
		}

		messageEvent, err := chats.SaveGroupMessage(client.UserID, p.GroupID, p.Content)
		if err != nil {
			return err
		}

		memberIDs, err := chats.GroupMemberIDs(p.GroupID)
		if err != nil {
			return err
		}

		hub.EmitToUsers(memberIDs, OutgoingEnvelope{
			Type:    TypeGroupMessage,
			Payload: messageEvent,
		})

		if notify != nil {
			for _, memberID := range memberIDs {
				if memberID == client.UserID {
					continue
				}
				muted, _ := chats.IsMuted(memberID, "group", p.GroupID)
				if !muted {
					notify(memberID, client.UserID, "new_group_message", p.GroupID)
				}
			}
		}

		return writeAck(client, TypeGroupMessageSend)
	case TypePostSend:
		var payload struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Privacy string `json:"privacy"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return &wsMessageError{Code: "INVALID_POST_PAYLOAD", Message: err.Error()}
		}

		payload.Title = strings.TrimSpace(payload.Title)
		payload.Content = strings.TrimSpace(payload.Content)
		payload.Privacy = strings.TrimSpace(payload.Privacy)

		//checkers
		if payload.Title == "" || payload.Content == "" {
			return &wsMessageError{Code: "VALIDATION_ERROR", Message: "Title and Content are required"}
		}
		if len(payload.Title) < 3 || len(payload.Title) > 60 {
			return &wsMessageError{Code: "VALIDATION_ERROR", Message: "Title should be between 3 and 60 chars"}
		}
		if len(payload.Content) < 3 || len(payload.Content) > 5000 {
			return &wsMessageError{Code: "VALIDATION_ERROR", Message: "Content should be between 3 and 5000 chars"}
		}

		allowedPrivacy := map[string]bool{"public": true, "private": true, "almost_private": true}
		if !allowedPrivacy[payload.Privacy] {
			payload.Privacy = "public"
		}
		// fixing client-provided imagePath accepted without validation: WebSocket posts should not accept image paths from client
		newPost := models.Post{
			ID:        uuid.New().String(),
			UserID:    client.UserID,
			Title:     payload.Title,
			Content:   payload.Content,
			Privacy:   payload.Privacy,
			CreatedAt: time.Now().UTC(),
		}
		if err := PostRepo.CreatePost(&newPost); err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}
		postWithAuthor, err := PostRepo.GetSinglePost(newPost.ID, client.UserID)
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}
		var userIDs []string
		switch newPost.Privacy {
		case "public":
			userIDs, err = PostRepo.GetFeedUserIDs(newPost.UserID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			userIDs = append(userIDs, newPost.UserID) // poster sees their own post
		case "almost_private":
			userIDs, err = PostRepo.GetFollowersIDs(newPost.UserID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			userIDs = append(userIDs, newPost.UserID)
		case "private":
			userIDs, err = PostRepo.GetAllowedUsers(newPost.ID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			userIDs = append(userIDs, postWithAuthor.UserID)
		}
		hub.EmitToUsers(userIDs, OutgoingEnvelope{
			Type:    TypePostEvent,
			Payload: postWithAuthor,
		})
		return writeAck(client, TypePostSend)
	case TypePostDelete:
		var payload struct {
			PostID string `json:"postId"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return &wsMessageError{Code: "INVALID_PAYLOAD", Message: err.Error()}
		}
		ownerID, err := PostRepo.GetPostOwnerID(payload.PostID)
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}
		if ownerID != client.UserID {
			return &wsMessageError{Code: "NOT_ALLOWED", Message: "User is not allowed to delete this post"}
		}
		// fixing post deleted before fetching privacy for WS broadcast: get privacy info before deletion
		postBeforeDelete, err := PostRepo.GetSinglePost(payload.PostID, client.UserID)
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}
		var postPrivacy string
		if postBeforeDelete != nil {
			postPrivacy = postBeforeDelete.Privacy
		}

		// Resolve recipients BEFORE deletion — post_allowed_users has ON DELETE CASCADE,
		// so GetAllowedUsers would return empty after DeletePost.
		var userIDs []string
		switch postPrivacy {
		case "almost_private":
			userIDs, err = PostRepo.GetFollowersIDs(ownerID)
		case "private":
			userIDs, err = PostRepo.GetAllowedUsers(payload.PostID)
		default:
			userIDs, err = PostRepo.GetFeedUserIDs(ownerID)
		}
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}

		if err := PostRepo.DeletePost(payload.PostID); err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}

		userIDs = append(userIDs, ownerID)

		hub.EmitToUsers(userIDs, OutgoingEnvelope{
			Type:    TypePostDelete,
			Payload: map[string]string{"postId": payload.PostID},
		})
		return writeAck(client, TypePostDelete)
	case TypeCommentSend:
		var payload struct {
			PostID  string `json:"postId"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal(env.Payload, &payload); err != nil {
			return &wsMessageError{Code: "INVALID_COMMENT_PAYLOAD", Message: err.Error()}
		}

		payload.Content = strings.TrimSpace(payload.Content)

		// fixing missing validation for empty comment: must have content
		if payload.Content == "" {
			return &wsMessageError{Code: "VALIDATION_ERROR", Message: "Comment content is required"}
		}

		if len(payload.Content) < 3 || len(payload.Content) > 2500 {
			return &wsMessageError{
				Code:    "VALIDATION_ERROR",
				Message: "Comment should be between 3 to 2500 characters",
			}
		}

		// fixing missing post access validation before creating comment via WebSocket
		postDetails, err := PostRepo.GetSinglePost(payload.PostID, client.UserID)
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}
		if postDetails == nil {
			return &wsMessageError{Code: "NOT_ALLOWED", Message: "Post not found or access denied"}
		}

		// fixing client-provided imagePath accepted without validation: WebSocket comments should not accept image paths
		newComment := models.Comment{
			ID:        uuid.New().String(),
			PostID:    payload.PostID,
			UserID:    client.UserID,
			Content:   payload.Content,
			CreatedAt: time.Now().UTC(),
		}
		if err := CommentRepo.CreateComment(&newComment); err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}

		commentWithAuthor, err := CommentRepo.GetCommentByID(newComment.ID)
		if err != nil {
			return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
		}

		var userIDs []string
		switch postDetails.Privacy {
		case "public":
			userIDs, err = PostRepo.GetFeedUserIDs(postDetails.UserID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			userIDs = append(userIDs, postDetails.UserID)
		case "almost_private":
			userIDs, err = PostRepo.GetFollowersIDs(postDetails.UserID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			userIDs = append(userIDs, postDetails.UserID)
		case "private":
			userIDs, err = PostRepo.GetAllowedUsers(postDetails.ID)
			if err != nil {
				return &wsMessageError{Code: "DB_ERROR", Message: err.Error()}
			}
			// fixing private post comment adding comment author instead of post owner to recipients
			userIDs = append(userIDs, postDetails.UserID)
		}

		hub.EmitToUsers(userIDs, OutgoingEnvelope{
			Type:    TypeCommentEvent,
			Payload: commentWithAuthor,
		})

		return writeAck(client, TypeCommentSend)

	case TypePrivateTyping:
		var p PrivateTypingPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil // silently drop malformed typing signals
		}
		p.ToUserID = strings.TrimSpace(p.ToUserID)
		if p.ToUserID == "" || p.ToUserID == client.UserID {
			return nil
		}
		hub.EmitToUser(p.ToUserID, OutgoingEnvelope{
			Type: TypeTypingEvent,
			Payload: TypingEventPayload{
				SenderID:    client.UserID,
				ContextType: "private",
				ContextID:   client.UserID,
				IsTyping:    p.IsTyping,
			},
		})
		return nil

	case TypeGroupTyping:
		var p GroupTypingPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return nil
		}
		p.GroupID = strings.TrimSpace(p.GroupID)
		if p.GroupID == "" {
			return nil
		}
		allowed, err := chats.CanSendGroupMessage(client.UserID, p.GroupID)
		if err != nil || !allowed {
			return nil
		}
		memberIDs, err := chats.GroupMemberIDs(p.GroupID)
		if err != nil {
			return nil
		}
		for _, memberID := range memberIDs {
			if memberID == client.UserID {
				continue
			}
			hub.EmitToUser(memberID, OutgoingEnvelope{
				Type: TypeTypingEvent,
				Payload: TypingEventPayload{
					SenderID:    client.UserID,
					ContextType: "group",
					ContextID:   p.GroupID,
					IsTyping:    p.IsTyping,
				},
			})
		}
		return nil

	default:
		return &wsMessageError{
			Code:    "UNSUPPORTED_TYPE",
			Message: "unsupported message type: " + env.Type,
		}
	}
}

func writeAck(client *Client, ackFor string) error {
	return client.WriteJSON(OutgoingEnvelope{
		Type: TypeAck,
		Payload: AckPayload{
			For: ackFor,
		},
	})
}

func writeProtocolError(client *Client, code, message string) error {
	return client.WriteJSON(OutgoingEnvelope{
		Type: TypeError,
		Payload: ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func writeProtocolAndClose(client *Client, code, message string, closeCode int) error {
	if err := writeProtocolError(client, code, message); err != nil {
		return err
	}
	closeWebsocket(client, closeCode, message)
	return nil
}

func closeWebsocket(client *Client, closeCode int, closeText string) {
	if client == nil || client.Conn == nil {
		return
	}
	wsConn, ok := client.Conn.(*websocket.Conn)
	if !ok {
		_ = client.Conn.Close()
		return
	}

	_ = wsConn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, closeText),
		time.Now().Add(time.Second),
	)
	_ = wsConn.Close()
}
