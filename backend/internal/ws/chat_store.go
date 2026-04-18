package ws

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type ChatStore struct {
	db *sql.DB
}

type ChatMessage struct {
	ID         string `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

func NewChatStore(db *sql.DB) *ChatStore {
	return &ChatStore{db: db}
}

// CanSendPrivateMessage enforces "at least one follows the other".
func (s *ChatStore) CanSendPrivateMessage(senderID, receiverID string) (bool, error) {
	if senderID == "" || receiverID == "" {
		return false, nil
	}
	if senderID == receiverID {
		return true, nil
	}

	const receiverExistsQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE id = ?);`
	var receiverExists bool
	if err := s.db.QueryRow(receiverExistsQuery, receiverID).Scan(&receiverExists); err != nil {
		return false, err
	}
	if !receiverExists {
		return false, nil
	}

	const followEitherDirectionQuery = `
		SELECT EXISTS(
			SELECT 1
			FROM followers
			WHERE (follower_id = ? AND following_id = ?)
			   OR (follower_id = ? AND following_id = ?)
		);
	`
	var allowed bool
	if err := s.db.QueryRow(followEitherDirectionQuery, senderID, receiverID, receiverID, senderID).Scan(&allowed); err != nil {
		return false, err
	}
	return allowed, nil
}

func (s *ChatStore) SavePrivateMessage(senderID, receiverID, content string) (PrivateMessageEventPayload, error) {
	msgID, err := newMessageID()
	if err != nil {
		return PrivateMessageEventPayload{}, err
	}

	createdAt := time.Now().UTC()
	const insertQuery = `
		INSERT INTO messages (id, sender_id, receiver_id, group_id, content, created_at)
		VALUES (?, ?, ?, NULL, ?, ?);
	`
	if _, err := s.db.Exec(insertQuery, msgID, senderID, receiverID, content, createdAt); err != nil {
		return PrivateMessageEventPayload{}, err
	}

	return PrivateMessageEventPayload{
		ID:         msgID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  createdAt.Format(time.RFC3339),
	}, nil
}

func (s *ChatStore) CanSendGroupMessage(senderID, groupID string) (bool, error) {
	if senderID == "" || groupID == "" {
		return false, nil
	}

	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE group_id = ? AND user_id = ?
		);
	`
	var allowed bool
	if err := s.db.QueryRow(query, groupID, senderID).Scan(&allowed); err != nil {
		return false, err
	}
	return allowed, nil
}

func (s *ChatStore) GroupMemberIDs(groupID string) ([]string, error) {
	const query = `
		SELECT user_id
		FROM group_members
		WHERE group_id = ?;
	`
	rows, err := s.db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		ids = append(ids, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *ChatStore) SaveGroupMessage(senderID, groupID, content string) (GroupMessageEventPayload, error) {
	msgID, err := newMessageID()
	if err != nil {
		return GroupMessageEventPayload{}, err
	}

	createdAt := time.Now().UTC()
	const insertQuery = `
		INSERT INTO messages (id, sender_id, receiver_id, group_id, content, created_at)
		VALUES (?, ?, NULL, ?, ?, ?);
	`
	if _, err := s.db.Exec(insertQuery, msgID, senderID, groupID, content, createdAt); err != nil {
		return GroupMessageEventPayload{}, err
	}

	return GroupMessageEventPayload{
		ID:        msgID,
		SenderID:  senderID,
		GroupID:   groupID,
		Content:   content,
		CreatedAt: createdAt.Format(time.RFC3339),
	}, nil
}

func (s *ChatStore) GetPrivateMessages(userID, otherUserID string, limit int, before string) ([]ChatMessage, error) {
	baseQuery := `
		SELECT id, sender_id, COALESCE(receiver_id, ''), COALESCE(group_id, ''), content, created_at
		FROM messages
		WHERE group_id IS NULL
		  AND ((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))
	`

	args := []any{userID, otherUserID, otherUserID, userID}
	if before != "" {
		baseQuery += " AND datetime(created_at) < datetime(?)"
		args = append(args, before)
	}

	baseQuery += " ORDER BY datetime(created_at) DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0, limit)
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	reverseMessages(messages)
	return messages, nil
}

func (s *ChatStore) GetGroupMessages(groupID string, limit int, before string) ([]ChatMessage, error) {
	baseQuery := `
		SELECT id, sender_id, COALESCE(receiver_id, ''), COALESCE(group_id, ''), content, created_at
		FROM messages
		WHERE group_id = ?
	`

	args := []any{groupID}
	if before != "" {
		baseQuery += " AND datetime(created_at) < datetime(?)"
		args = append(args, before)
	}

	baseQuery += " ORDER BY datetime(created_at) DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0, limit)
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.GroupID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	reverseMessages(messages)
	return messages, nil
}

// MarkChatRead upserts the last_read_at timestamp for the user in a given conversation.
// chatType must be "private" or "group". chatID is the other user ID or group ID.
func (s *ChatStore) MarkChatRead(userID, chatType, chatID string) error {
	const query = `
		INSERT INTO chat_last_read (user_id, chat_type, chat_id, last_read_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, chat_type, chat_id) DO UPDATE SET last_read_at = excluded.last_read_at;
	`
	_, err := s.db.Exec(query, userID, chatType, chatID, time.Now().UTC().Format(time.RFC3339))
	return err
}

// UnreadCounts returns a map of chatID → unread message count for the given user
// across all private and group conversations. Keys are prefixed: "private:<id>" / "group:<id>".
func (s *ChatStore) UnreadCounts(userID string) (map[string]int, error) {
	const query = `
		SELECT
			CASE
				WHEN m.group_id IS NOT NULL THEN 'group:' || m.group_id
				ELSE 'private:' || CASE
					WHEN m.sender_id = ? THEN m.receiver_id
					ELSE m.sender_id
				END
			END AS conv_key,
			COUNT(*) AS cnt
		FROM messages m
		LEFT JOIN chat_last_read clr ON (
			clr.user_id = ?
			AND (
				(m.group_id IS NOT NULL AND clr.chat_type = 'group'   AND clr.chat_id = m.group_id)
				OR
				(m.group_id IS NULL     AND clr.chat_type = 'private' AND (clr.chat_id = m.sender_id OR clr.chat_id = m.receiver_id))
			)
		)
		WHERE
			m.sender_id != ?
			AND (
				m.receiver_id = ?
				OR m.group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)
			)
			AND (clr.last_read_at IS NULL OR datetime(m.created_at) > datetime(clr.last_read_at))
		GROUP BY conv_key;
	`
	rows, err := s.db.Query(query, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var key string
		var cnt int
		if err := rows.Scan(&key, &cnt); err != nil {
			return nil, err
		}
		counts[key] = cnt
	}
	return counts, rows.Err()
}

// IsMuted returns true if the user has muted notifications for the given conversation.
func (s *ChatStore) IsMuted(userID, chatType, chatID string) (bool, error) {
	const query = `
		SELECT muted FROM chat_notification_preferences
		WHERE user_id = ? AND chat_type = ? AND chat_id = ?;
	`
	var muted int
	err := s.db.QueryRow(query, userID, chatType, chatID).Scan(&muted)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return muted == 1, nil
}

// SetMuted upserts the mute preference for the user in a given conversation.
func (s *ChatStore) SetMuted(userID, chatType, chatID string, muted bool) error {
	muteVal := 0
	if muted {
		muteVal = 1
	}
	const query = `
		INSERT INTO chat_notification_preferences (user_id, chat_type, chat_id, muted)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, chat_type, chat_id) DO UPDATE SET muted = excluded.muted;
	`
	_, err := s.db.Exec(query, userID, chatType, chatID, muteVal)
	return err
}

func reverseMessages(messages []ChatMessage) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}

func newMessageID() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("failed to generate message id")
	}
	return hex.EncodeToString(b), nil
}
