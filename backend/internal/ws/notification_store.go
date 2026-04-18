package ws

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type NotificationStore struct {
	db *sql.DB
}

type NotificationRecord struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ActorID   string `json:"actor_id,omitempty"`
	Type      string `json:"type"`
	Payload   string `json:"payload,omitempty"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type CreateNotificationInput struct {
	UserID  string
	ActorID string
	Type    string
	Payload string
}

func NewNotificationStore(db *sql.DB) *NotificationStore {
	return &NotificationStore{db: db}
}

func (s *NotificationStore) Create(input CreateNotificationInput) (NotificationRecord, error) {
	userID := strings.TrimSpace(input.UserID)
	actorID := strings.TrimSpace(input.ActorID)
	kind := strings.TrimSpace(input.Type)
	payload := strings.TrimSpace(input.Payload)
	if userID == "" {
		return NotificationRecord{}, errors.New("user_id is required")
	}
	if kind == "" {
		return NotificationRecord{}, errors.New("type is required")
	}

	notificationID, err := newNotificationID()
	if err != nil {
		return NotificationRecord{}, err
	}

	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	const query = `
		INSERT INTO notifications (id, user_id, actor_id, type, payload, is_read, created_at)
		VALUES (?, ?, NULLIF(?, ''), ?, NULLIF(?, ''), 0, ?);
	`
	if _, err := s.db.Exec(query, notificationID, userID, actorID, kind, payload, createdAt); err != nil {
		return NotificationRecord{}, err
	}

	return NotificationRecord{
		ID:        notificationID,
		UserID:    userID,
		ActorID:   actorID,
		Type:      kind,
		Payload:   payload,
		IsRead:    false,
		CreatedAt: createdAt,
	}, nil
}

func (s *NotificationStore) ListByUser(userID string, limit int, before string) ([]NotificationRecord, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, user_id, COALESCE(actor_id, ''), type, COALESCE(payload, ''), is_read, created_at
		FROM notifications
		WHERE user_id = ?
	`
	args := []any{userID}

	before = strings.TrimSpace(before)
	if before != "" {
		query += " AND created_at < ?"
		args = append(args, before)
	}

	query += " ORDER BY created_at DESC, id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]NotificationRecord, 0, limit)
	for rows.Next() {
		var record NotificationRecord
		var isReadInt int
		if err := rows.Scan(&record.ID, &record.UserID, &record.ActorID, &record.Type, &record.Payload, &isReadInt, &record.CreatedAt); err != nil {
			return nil, err
		}
		record.IsRead = isReadInt == 1
		notifications = append(notifications, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *NotificationStore) UnreadCount(userID string) (int, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, errors.New("user_id is required")
	}

	const query = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND is_read = 0;
	`
	var count int
	if err := s.db.QueryRow(query, userID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *NotificationStore) MarkRead(userID, notificationID string) (bool, error) {
	userID = strings.TrimSpace(userID)
	notificationID = strings.TrimSpace(notificationID)
	if userID == "" {
		return false, errors.New("user_id is required")
	}
	if notificationID == "" {
		return false, errors.New("notification_id is required")
	}

	const query = `
		UPDATE notifications
		SET is_read = 1
		WHERE id = ? AND user_id = ? AND is_read = 0;
	`
	result, err := s.db.Exec(query, notificationID, userID)
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *NotificationStore) MarkAllRead(userID string) (int64, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return 0, errors.New("user_id is required")
	}

	const query = `
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ? AND is_read = 0;
	`
	result, err := s.db.Exec(query, userID)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func newNotificationID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("failed to generate notification id")
	}
	return hex.EncodeToString(b), nil
}
