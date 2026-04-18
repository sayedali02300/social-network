package ws

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type HistoryResponse struct {
	Items      []ChatMessage `json:"items"`
	NextBefore string        `json:"next_before,omitempty"`
}

func NewPrivateHistoryHandler(sessions *SessionStore, chats *ChatStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		requesterID, ok := authenticatedUserIDFromRequest(r, sessions, w)
		if !ok {
			return
		}

		otherUserID := strings.TrimSpace(r.PathValue("chatId"))
		if otherUserID == "" {
			writeHTTPError(w, http.StatusNotFound, "invalid private history path")
			return
		}

		allowed, err := chats.CanSendPrivateMessage(requesterID, otherUserID)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, "failed to validate access")
			return
		}
		if !allowed {
			writeHTTPError(w, http.StatusForbidden, "private history not allowed")
			return
		}

		limit := parseLimit(r, 50, 100)
		before := strings.TrimSpace(r.URL.Query().Get("before"))

		items, err := chats.GetPrivateMessages(requesterID, otherUserID, limit, before)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, "failed to load private history")
			return
		}

		// Mark conversation as read when the user loads the first page (no pagination cursor).
		if before == "" {
			_ = chats.MarkChatRead(requesterID, "private", otherUserID)
		}

		resp := HistoryResponse{Items: items}
		if len(items) == limit {
			resp.NextBefore = items[0].CreatedAt
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func NewGroupHistoryHandler(sessions *SessionStore, chats *ChatStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		requesterID, ok := authenticatedUserIDFromRequest(r, sessions, w)
		if !ok {
			return
		}

		groupID := strings.TrimSpace(r.PathValue("chatId"))
		if groupID == "" {
			writeHTTPError(w, http.StatusNotFound, "invalid group history path")
			return
		}

		allowed, err := chats.CanSendGroupMessage(requesterID, groupID)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, "failed to validate access")
			return
		}
		if !allowed {
			writeHTTPError(w, http.StatusForbidden, "group history not allowed")
			return
		}

		limit := parseLimit(r, 50, 100)
		before := strings.TrimSpace(r.URL.Query().Get("before"))

		items, err := chats.GetGroupMessages(groupID, limit, before)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, "failed to load group history")
			return
		}

		if before == "" {
			_ = chats.MarkChatRead(requesterID, "group", groupID)
		}

		resp := HistoryResponse{Items: items}
		if len(items) == limit {
			resp.NextBefore = items[0].CreatedAt
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// NewUnreadCountsHandler returns unread message counts per conversation for the authenticated user.
// Response: { "private:<userId>": <count>, "group:<groupId>": <count>, ... }
func NewUnreadCountsHandler(sessions *SessionStore, chats *ChatStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		userID, ok := authenticatedUserIDFromRequest(r, sessions, w)
		if !ok {
			return
		}

		counts, err := chats.UnreadCounts(userID)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, "failed to load unread counts")
			return
		}

		writeJSON(w, http.StatusOK, counts)
	}
}

// NewChatMuteHandler handles GET and PUT for /api/chats/{chatType}/{chatId}/mute
// GET  → returns { "muted": true/false }
// PUT  → body { "muted": true/false }, upserts preference, returns same
func NewChatMuteHandler(sessions *SessionStore, chats *ChatStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := authenticatedUserIDFromRequest(r, sessions, w)
		if !ok {
			return
		}

		chatType := strings.TrimSpace(r.PathValue("chatType"))
		chatID := strings.TrimSpace(r.PathValue("chatId"))
		if chatType != "private" && chatType != "group" {
			writeHTTPError(w, http.StatusBadRequest, "chatType must be private or group")
			return
		}
		if chatID == "" {
			writeHTTPError(w, http.StatusBadRequest, "chatId is required")
			return
		}

		switch r.Method {
		case http.MethodGet:
			muted, err := chats.IsMuted(userID, chatType, chatID)
			if err != nil {
				writeHTTPError(w, http.StatusInternalServerError, "failed to load mute preference")
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"muted": muted})

		case http.MethodPut:
			var body struct {
				Muted bool `json:"muted"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeHTTPError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			if err := chats.SetMuted(userID, chatType, chatID, body.Muted); err != nil {
				writeHTTPError(w, http.StatusInternalServerError, "failed to save mute preference")
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"muted": body.Muted})

		default:
			writeHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	}
}

func authenticatedUserIDFromRequest(r *http.Request, sessions *SessionStore, w http.ResponseWriter) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		writeHTTPError(w, http.StatusUnauthorized, "unauthorized: missing session cookie")
		return "", false
	}

	userID, err := sessions.ValidateSession(cookie.Value, r)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "session revoked") {
			writeHTTPError(w, http.StatusUnauthorized, "unauthorized: invalid or expired session")
			return "", false
		}
		writeHTTPError(w, http.StatusInternalServerError, "session lookup failed")
		return "", false
	}
	return userID, true
}

func parseLimit(r *http.Request, defaultValue, maxValue int) int {
	limitText := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitText == "" {
		return defaultValue
	}

	n, err := strconv.Atoi(limitText)
	if err != nil || n <= 0 {
		return defaultValue
	}
	if n > maxValue {
		return maxValue
	}
	return n
}

func writeHTTPError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
