package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"socialnetwork/backend/internal/models"
	wshandler "socialnetwork/backend/internal/ws"
	"strconv"
	"strings"
	"time"
)

type notificationActor struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
}

type notificationResponse struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Payload   string              `json:"payload"`
	IsRead    bool                `json:"isRead"`
	CreatedAt string              `json:"createdAt"`
	Actor     *notificationActor  `json:"actor,omitempty"`
	Target    *notificationTarget `json:"target,omitempty"`
	Message   string              `json:"message"`
}

type notificationTarget struct {
	GroupID    string `json:"groupId,omitempty"`
	GroupTitle string `json:"groupTitle,omitempty"`
	EventID    string `json:"eventId,omitempty"`
	EventTitle string `json:"eventTitle,omitempty"`
	PostID     string `json:"postId,omitempty"`
	CommentID  string `json:"commentId,omitempty"`
}

func (h *HandlerStruct) GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	limit := 20
	limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitRaw != "" {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil || parsed <= 0 {
			sendJSONError(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	notifications, err := h.UserRepo.ListNotificationsByUserID(userID, limit)
	if err != nil {
		log.Printf("Error loading notifications for user %s: %v", userID, err)
		sendJSONError(w, "Could not load notifications", http.StatusInternalServerError)
		return
	}

	response := make([]notificationResponse, 0, len(notifications))
	for _, item := range notifications {
		response = append(response, h.buildNotificationResponse(item, userID))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *HandlerStruct) PatchNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	notificationID := strings.TrimSpace(r.PathValue("notificationId"))
	if notificationID == "" {
		sendJSONError(w, "notificationId is required", http.StatusBadRequest)
		return
	}

	updated, err := h.UserRepo.MarkNotificationRead(userID, notificationID)
	if err != nil {
		log.Printf("Error marking notification read: %v", err)
		sendJSONError(w, "Could not mark notification as read", http.StatusInternalServerError)
		return
	}
	if !updated {
		sendJSONError(w, "Notification not found or already read", http.StatusNotFound)
		return
	}

	h.emitNotificationUnreadCountRealtime(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"id":     notificationID,
		"status": "read",
	})
}

func (h *HandlerStruct) PatchNotificationsReadAllHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	updatedCount, err := h.UserRepo.MarkAllNotificationsRead(userID)
	if err != nil {
		log.Printf("Error marking all notifications read for user %s: %v", userID, err)
		sendJSONError(w, "Could not mark all notifications as read", http.StatusInternalServerError)
		return
	}

	h.emitNotificationUnreadCountRealtime(userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]int64{
		"updated": updatedCount,
	})
}

func (h *HandlerStruct) GetNotificationsUnreadCountHandler(w http.ResponseWriter, r *http.Request) {
	if h.UserRepo == nil {
		sendJSONError(w, "User repository is not configured", http.StatusInternalServerError)
		return
	}

	userID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	count, err := h.UserRepo.CountUnreadNotificationsByUserID(userID)
	if err != nil {
		log.Printf("Error counting unread notifications for user %s: %v", userID, err)
		sendJSONError(w, "Could not load unread notification count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]int{
		"unread": count,
	})
}

func (h *HandlerStruct) emitNotificationRealtime(item *models.Notification) {
	if h == nil || h.Notifier == nil || item == nil {
		return
	}

	out := h.buildNotificationResponse(*item, item.UserID)
	h.Notifier.EmitToUser(item.UserID, wshandler.OutgoingEnvelope{
		Type:    "notification",
		Payload: out,
	})
}

func (h *HandlerStruct) buildNotificationResponse(item models.Notification, viewerID string) notificationResponse {
	out := notificationResponse{
		ID:        item.ID,
		Type:      item.Type,
		Payload:   item.Payload,
		IsRead:    item.IsRead,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}

	out.Actor = h.loadNotificationActor(item.ActorID, item.ID)
	out.Target = h.loadNotificationTarget(item.Type, item.Payload, viewerID, item.ID)
	out.Message = buildNotificationMessage(out.Type, out.Actor, out.Target, viewerID)

	return out
}

func (h *HandlerStruct) loadNotificationActor(actorID, notificationID string) *notificationActor {
	trimmedActorID := strings.TrimSpace(actorID)
	if trimmedActorID == "" || h == nil || h.UserRepo == nil {
		return nil
	}

	actor, err := h.UserRepo.GetUserByID(trimmedActorID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error loading actor user %s for notification %s: %v", trimmedActorID, notificationID, err)
		}
		return nil
	}

	return &notificationActor{
		ID:        actor.ID,
		FirstName: actor.FirstName,
		LastName:  actor.LastName,
		Nickname:  actor.Nickname,
		Avatar:    actor.Avatar,
	}
}

func (h *HandlerStruct) loadNotificationTarget(notificationType, payload, viewerID, notificationID string) *notificationTarget {
	if h == nil || h.GroupRepo == nil {
		return nil
	}

	trimmedPayload := strings.TrimSpace(payload)
	if trimmedPayload == "" {
		return nil
	}

	switch notificationType {
	case "group_invitation_received":
		invite, err := h.GroupRepo.GetInviteByID(trimmedPayload)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading invite %s for notification %s: %v", trimmedPayload, notificationID, err)
			}
			return nil
		}
		if viewerID != "" && invite.ReceiverID != viewerID {
			return nil
		}
		group, err := h.GroupRepo.GetGroup(invite.GroupID, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading group %s for notification %s: %v", invite.GroupID, notificationID, err)
			}
			return &notificationTarget{GroupID: invite.GroupID}
		}
		return &notificationTarget{
			GroupID:    invite.GroupID,
			GroupTitle: group.Title,
		}
	case "group_join_request_received":
		request, err := h.GroupRepo.GetJoinRequestByID(trimmedPayload)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading join request %s for notification %s: %v", trimmedPayload, notificationID, err)
			}
			return nil
		}
		group, err := h.GroupRepo.GetGroup(request.GroupID, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading group %s for notification %s: %v", request.GroupID, notificationID, err)
			}
			return &notificationTarget{GroupID: request.GroupID}
		}
		if viewerID != "" && group.CreatorID != viewerID {
			return nil
		}
		return &notificationTarget{
			GroupID:    request.GroupID,
			GroupTitle: group.Title,
		}
	case "group_event_created", "group_event_updated", "group_event_due":
		event, err := h.GroupRepo.GetEvent(trimmedPayload, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading event %s for notification %s: %v", trimmedPayload, notificationID, err)
			}
			return nil
		}
		target := &notificationTarget{
			GroupID:    event.GroupID,
			EventID:    event.ID,
			EventTitle: event.Title,
		}
		group, err := h.GroupRepo.GetGroup(event.GroupID, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading group %s for notification %s: %v", event.GroupID, notificationID, err)
			}
			return target
		}
		target.GroupTitle = group.Title
		return target
	case "new_comment", "new_comment_reply":
		var dp struct {
			PostID    string `json:"postId"`
			CommentID string `json:"commentId"`
		}
		if err := json.Unmarshal([]byte(trimmedPayload), &dp); err != nil {
			log.Printf("Error parsing comment payload for notification %s: %v", notificationID, err)
			return nil
		}
		return &notificationTarget{
			PostID:    dp.PostID,
			CommentID: dp.CommentID,
		}
	case "new_private_message":
		return nil
	case "new_group_message":
		group, err := h.GroupRepo.GetGroup(trimmedPayload, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading group %s for notification %s: %v", trimmedPayload, notificationID, err)
			}
			return &notificationTarget{GroupID: trimmedPayload}
		}
		return &notificationTarget{GroupID: group.ID, GroupTitle: group.Title}
	case "group_invitation_accepted", "group_invitation_declined",
		"group_join_request_accepted", "group_join_request_declined":
		group, err := h.GroupRepo.GetGroup(trimmedPayload, viewerID)
		if err != nil {
			if err != sql.ErrNoRows {
				log.Printf("Error loading group %s for notification %s: %v", trimmedPayload, notificationID, err)
			}
			return &notificationTarget{GroupID: trimmedPayload}
		}
		return &notificationTarget{
			GroupID:    group.ID,
			GroupTitle: group.Title,
		}
	case "group_event_deleted":
		var dp struct {
			GroupID    string `json:"groupId"`
			EventTitle string `json:"eventTitle"`
		}
		if err := json.Unmarshal([]byte(trimmedPayload), &dp); err != nil {
			log.Printf("Error parsing deleted event payload for notification %s: %v", notificationID, err)
			return nil
		}
		target := &notificationTarget{
			GroupID:    dp.GroupID,
			EventTitle: dp.EventTitle,
		}
		group, err := h.GroupRepo.GetGroup(dp.GroupID, viewerID)
		if err == nil {
			target.GroupTitle = group.Title
		}
		return target
	default:
		return nil
	}
}

func (h *HandlerStruct) emitNotificationUnreadCountRealtime(userID string) {
	if h == nil || h.Notifier == nil || h.UserRepo == nil {
		return
	}

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return
	}

	count, err := h.UserRepo.CountUnreadNotificationsByUserID(trimmedUserID)
	if err != nil {
		log.Printf("Error counting unread notifications for realtime emit user %s: %v", trimmedUserID, err)
		return
	}

	h.Notifier.EmitToUser(trimmedUserID, wshandler.OutgoingEnvelope{
		Type: "notification_unread_count",
		Payload: map[string]int{
			"unread": count,
		},
	})
}

func buildNotificationMessage(notificationType string, actor *notificationActor, target *notificationTarget, viewerID string) string {
	name := "Someone"
	if actor != nil {
		if actor.Nickname != "" {
			name = "@" + actor.Nickname
		} else {
			fullName := strings.TrimSpace(actor.FirstName + " " + actor.LastName)
			if fullName != "" {
				name = fullName
			}
		}
	}
	isSelfActor := actor != nil && strings.TrimSpace(actor.ID) != "" && strings.TrimSpace(actor.ID) == strings.TrimSpace(viewerID)

	switch notificationType {
	case "follow_request_received":
		return name + " sent you a follow request."
	case "follow_request_accepted":
		return name + " accepted your follow request."
	case "new_follower":
		return name + " started following you."
	case "group_invitation_received":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` invited you to join "` + target.GroupTitle + `".`
		}
		return name + " invited you to a group."
	case "group_join_request_received":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` requested to join "` + target.GroupTitle + `".`
		}
		return name + " requested to join your group."
	case "group_event_created":
		if isSelfActor {
			if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
				return `You created the event "` + target.EventTitle + `" in "` + target.GroupTitle + `".`
			}
			if target != nil && strings.TrimSpace(target.EventTitle) != "" {
				return `You created the event "` + target.EventTitle + `".`
			}
			if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
				return `You created a new event in "` + target.GroupTitle + `".`
			}
			return "You created a new group event."
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` created the event "` + target.EventTitle + `" in "` + target.GroupTitle + `".`
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" {
			return name + ` created the event "` + target.EventTitle + `".`
		}
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` created a new event in "` + target.GroupTitle + `".`
		}
		return name + " created a new group event."
	case "new_comment":
		return name + " commented on your post."
	case "new_comment_reply":
		return name + " replied to your comment."
	case "new_group_message":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` sent a message in "` + target.GroupTitle + `".`
		}
		return name + " sent a message in a group."
	case "new_private_message":
		if actor != nil {
			fullName := strings.TrimSpace(actor.FirstName + " " + actor.LastName)
			if actor.Nickname != "" {
				return "@" + actor.Nickname + " sent you a message."
			}
			if fullName != "" {
				return fullName + " sent you a message."
			}
		}
		return "You have a new message."
	case "group_invitation_accepted":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` accepted your invite to "` + target.GroupTitle + `".`
		}
		return name + " accepted your group invite."
	case "group_invitation_declined":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` declined your invite to "` + target.GroupTitle + `".`
		}
		return name + " declined your group invite."
	case "group_join_request_accepted":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return `Your request to join "` + target.GroupTitle + `" was accepted.`
		}
		return "Your group join request was accepted."
	case "group_join_request_declined":
		if target != nil && strings.TrimSpace(target.GroupTitle) != "" {
			return `Your request to join "` + target.GroupTitle + `" was declined.`
		}
		return "Your group join request was declined."
	case "group_event_updated":
		if isSelfActor {
			if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
				return `You updated the event "` + target.EventTitle + `" in "` + target.GroupTitle + `".`
			}
			if target != nil && strings.TrimSpace(target.EventTitle) != "" {
				return `You updated the event "` + target.EventTitle + `".`
			}
			return "You updated a group event."
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
			return name + ` updated the event "` + target.EventTitle + `" in "` + target.GroupTitle + `".`
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" {
			return name + ` updated the event "` + target.EventTitle + `".`
		}
		return name + " updated a group event."
	case "group_event_due":
		if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
			return `The event "` + target.EventTitle + `" in "` + target.GroupTitle + `" is starting now.`
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" {
			return `The event "` + target.EventTitle + `" is starting now.`
		}
		return "A group event is starting now."
	case "group_event_deleted":
		if target != nil && strings.TrimSpace(target.EventTitle) != "" && strings.TrimSpace(target.GroupTitle) != "" {
			return `The event "` + target.EventTitle + `" in "` + target.GroupTitle + `" was cancelled.`
		}
		if target != nil && strings.TrimSpace(target.EventTitle) != "" {
			return `The event "` + target.EventTitle + `" was cancelled.`
		}
		return "A group event was cancelled."
	default:
		return "You have a new notification."
	}
}
