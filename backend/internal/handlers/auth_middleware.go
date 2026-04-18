package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const authUserIDContextKey contextKey = "auth_user_id"

func (h *HandlerStruct) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.SessionStore == nil {
			sendJSONError(w, "Auth services are not configured", http.StatusInternalServerError)
			return
		}

		cookie, err := r.Cookie("session_id")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
				return
			}
			sendJSONError(w, "Invalid session cookie", http.StatusBadRequest)
			return
		}

		sessionID := strings.TrimSpace(cookie.Value)
		if sessionID == "" {
			sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
			return
		}

		userID, err := h.SessionStore.ValidateSession(sessionID, extractClientIP(r))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "session revoked") {
				clearSessionCookie(w, isSecureRequest(r))
				sendJSONError(w, "Session expired or invalid", http.StatusUnauthorized)
				return
			}
			sendJSONError(w, "Could not validate session", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), authUserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func userIDFromRequest(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(authUserIDContextKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", false
	}
	return userID, true
}
