package repository

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"socialnetwork/backend/internal/models"
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

func (s *SessionStore) CreateSession(session *models.Session) error {
	hashedSessionID, err := hashSessionToken(session.ID)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO sessions (id, user_id, expires_at, created_at, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	// Use SQLite-compatible datetime format because active-session checks use CURRENT_TIMESTAMP.
	const sqliteDateTime = "2006-01-02 15:04:05"

	_, err = s.db.Exec(
		query,
		hashedSessionID,
		session.UserID,
		session.ExpiresAt.UTC().Format(sqliteDateTime),
		session.CreatedAt.UTC().Format(sqliteDateTime),
		session.IPAddress,
		session.UserAgent,
	)
	return err
}

// ValidateSession looks up the session and (when the session has an IP binding)
// revokes and rejects it if the caller's IP does not match the stored one.
// This prevents session-cookie theft from a different IP address.
func (s *SessionStore) ValidateSession(sessionID, clientIP string) (string, error) {
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

	// Only enforce IP binding when the session has a stored IP.
	// Legacy sessions (NULL ip_address) are allowed through unchanged.
	if storedIP != "" && storedIP != clientIP {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE id = ?`, hashedSessionID)
		return "", errors.New("session revoked: IP address mismatch")
	}

	return userID, nil
}

func (s *SessionStore) DeleteSession(sessionID string) error {
	hashedSessionID, err := hashSessionToken(sessionID)
	if err != nil {
		return err
	}

	const query = `DELETE FROM sessions WHERE id = ?`

	_, err = s.db.Exec(query, hashedSessionID)
	return err
}

func (s *SessionStore) DeleteSessionsByUserID(userID string) error {
	const query = `DELETE FROM sessions WHERE user_id = ?`

	_, err := s.db.Exec(query, userID)
	return err
}

func (s *SessionStore) GetSessionByID(sessionID string) (*models.Session, error) {
	hashedSessionID, err := hashSessionToken(sessionID)
	if err != nil {
		return nil, err
	}

	const query = `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ? AND expires_at > CURRENT_TIMESTAMP
		LIMIT 1
	`

	var session models.Session
	if err := s.db.QueryRow(query, hashedSessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &session, nil
}
