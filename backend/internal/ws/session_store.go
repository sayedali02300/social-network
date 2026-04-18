package ws

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strings"
)

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func hashSessionToken(sessionToken string) (string, error) {
	token := strings.TrimSpace(sessionToken)
	if token == "" {
		return "", errors.New("session token is empty")
	}

	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

// UserIDBySessionID returns the owning user if the session exists and is not expired.
func (s *SessionStore) UserIDBySessionID(sessionID string) (string, error) {
	hashedSessionID, err := hashSessionToken(sessionID)
	if err != nil {
		return "", err
	}

	const query = `
		SELECT user_id
		FROM sessions
		WHERE id = ? AND expires_at > CURRENT_TIMESTAMP
		LIMIT 1;
	`

	var userID string
	if err := s.db.QueryRow(query, hashedSessionID).Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

// extractClientIP returns the real client IP, respecting common reverse-proxy headers.
func extractClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// ValidateSession checks the session token and (when the session has an IP binding)
// revokes and rejects it if the caller's IP does not match the stored one.
func (s *SessionStore) ValidateSession(sessionID string, r *http.Request) (string, error) {
	hashedSessionID, err := hashSessionToken(sessionID)
	if err != nil {
		return "", err
	}

	const query = `
		SELECT user_id, COALESCE(ip_address, '')
		FROM sessions
		WHERE id = ? AND expires_at > CURRENT_TIMESTAMP
		LIMIT 1;
	`

	var userID, storedIP string
	if err := s.db.QueryRow(query, hashedSessionID).Scan(&userID, &storedIP); err != nil {
		return "", err
	}

	if storedIP != "" && storedIP != extractClientIP(r) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, hashedSessionID)
		return "", errors.New("session revoked: IP address mismatch")
	}

	return userID, nil
}
