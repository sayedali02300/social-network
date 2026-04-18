package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"socialnetwork/backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	db *sql.DB
}

type UserProfileUpdates struct {
	FirstName   *string
	LastName    *string
	DateOfBirth *string
	Avatar      *string
	Nickname    *string
	AboutMe     *string
}

var (
	ErrFollowRequestNotFound       = errors.New("follow request not found")
	ErrFollowRequestForbidden      = errors.New("follow request does not belong to receiver")
	ErrFollowRequestNotPending     = errors.New("follow request is not pending")
	ErrAlreadyFollowing            = errors.New("already following")
	ErrFollowRequestAlreadyPending = errors.New("follow request already pending")
)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists int
	if err := r.db.QueryRow(query, email).Scan(&exists); err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (r *UserRepository) NicknameExists(nickname string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(COALESCE(nickname, '')) = LOWER(?))`

	var exists int
	if err := r.db.QueryRow(query, nickname).Scan(&exists); err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (r *UserRepository) NicknameExistsForOtherUser(nickname, excludeUserID string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE LOWER(COALESCE(nickname, '')) = LOWER(?)
			  AND id <> ?
		)
	`

	var exists int
	if err := r.db.QueryRow(query, nickname, excludeUserID).Scan(&exists); err != nil {
		return false, err
	}

	return exists == 1, nil
}

func (r *UserRepository) CreateUser(user *models.User) error {
	const query = `
		INSERT INTO users (id, email, password, first_name, last_name, date_of_birth, avatar, nickname, about_me, is_public, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		nullIfEmpty(user.Avatar),
		nullIfEmpty(user.Nickname),
		nullIfEmpty(user.AboutMe),
		boolToSQLiteInt(user.IsPublic),
		user.CreatedAt.Format(time.RFC3339),
	)
	return err
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	const query = `
		SELECT id, email, password, first_name, last_name, date_of_birth,
		       COALESCE(avatar, ''), COALESCE(nickname, ''), COALESCE(about_me, ''), is_public, created_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`

	var user models.User
	var isPublicInt int
	if err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&isPublicInt,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	user.IsPublic = isPublicInt == 1
	return &user, nil
}

func (r *UserRepository) GetUserByEmailOrNickname(identifier string) (*models.User, error) {
	const query = `
		SELECT id, email, password, first_name, last_name, date_of_birth,
		       COALESCE(avatar, ''), COALESCE(nickname, ''), COALESCE(about_me, ''), is_public, created_at
		FROM users
		WHERE LOWER(email) = LOWER(?)
		   OR LOWER(COALESCE(nickname, '')) = LOWER(?)
		ORDER BY CASE WHEN LOWER(email) = LOWER(?) THEN 0 ELSE 1 END
		LIMIT 1
	`

	var user models.User
	var isPublicInt int
	if err := r.db.QueryRow(query, identifier, identifier, identifier).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&isPublicInt,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	user.IsPublic = isPublicInt == 1
	return &user, nil
}

func (r *UserRepository) GetUserByID(userID string) (*models.User, error) {
	const query = `
		SELECT id, email, password, first_name, last_name, date_of_birth,
		       COALESCE(avatar, ''), COALESCE(nickname, ''), COALESCE(about_me, ''), is_public, created_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`

	var user models.User
	var isPublicInt int
	if err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.Avatar,
		&user.Nickname,
		&user.AboutMe,
		&isPublicInt,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	user.IsPublic = isPublicInt == 1
	return &user, nil
}

func (r *UserRepository) SearchUsers(query, excludeUserID string, limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed == "" {
		return []models.User{}, nil
	}

	like := "%" + trimmed + "%"
	const searchQuery = `
		SELECT id, email, first_name, last_name, date_of_birth,
		       COALESCE(avatar, ''), COALESCE(nickname, ''), COALESCE(about_me, ''), is_public, created_at
		FROM users
		WHERE id <> ?
		  AND (
		    LOWER(COALESCE(nickname, '')) LIKE ?
		    OR LOWER(first_name) LIKE ?
		    OR LOWER(last_name) LIKE ?
		    OR LOWER(first_name || ' ' || last_name) LIKE ?
		  )
		ORDER BY
		  CASE
		    WHEN LOWER(COALESCE(nickname, '')) = ? THEN 0
		    WHEN LOWER(first_name || ' ' || last_name) = ? THEN 1
		    WHEN LOWER(COALESCE(nickname, '')) LIKE ? THEN 2
		    ELSE 3
		  END,
		  created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(searchQuery, excludeUserID, like, like, like, like, trimmed, trimmed, like, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]models.User, 0, limit)
	for rows.Next() {
		var user models.User
		var isPublicInt int
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.Avatar,
			&user.Nickname,
			&user.AboutMe,
			&isPublicInt,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}

		user.IsPublic = isPublicInt == 1
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// FollowedIDsAmong returns a set of IDs from candidateIDs that followerID follows.
// A single query replaces per-user follow checks in the search handler.
// FollowStatusAmong returns "following", "requested", or "" for each candidateID
// relative to followerID, in a single round-trip per table.
func (r *UserRepository) FollowStatusAmong(followerID string, candidateIDs []string) (map[string]string, error) {
	result := make(map[string]string, len(candidateIDs))
	if len(candidateIDs) == 0 {
		return result, nil
	}

	placeholders := strings.Repeat("?,", len(candidateIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, 0, len(candidateIDs)+1)
	args = append(args, followerID)
	for _, id := range candidateIDs {
		args = append(args, id)
	}

	// Check confirmed follows first.
	rows, err := r.db.Query(
		fmt.Sprintf(`SELECT following_id FROM followers WHERE follower_id = ? AND following_id IN (%s)`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = "following"
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Check pending follow requests for the remaining IDs.
	rows2, err := r.db.Query(
		fmt.Sprintf(`SELECT receiver_id FROM follow_requests WHERE sender_id = ? AND status = 'pending' AND receiver_id IN (%s)`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var id string
		if err := rows2.Scan(&id); err != nil {
			return nil, err
		}
		if result[id] == "" {
			result[id] = "requested"
		}
	}
	return result, rows2.Err()
}

func (r *UserRepository) FollowedIDsAmong(followerID string, candidateIDs []string) (map[string]bool, error) {
	if len(candidateIDs) == 0 {
		return map[string]bool{}, nil
	}

	placeholders := strings.Repeat("?,", len(candidateIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, 0, len(candidateIDs)+1)
	args = append(args, followerID)
	for _, id := range candidateIDs {
		args = append(args, id)
	}

	rows, err := r.db.Query(
		fmt.Sprintf(`SELECT following_id FROM followers WHERE follower_id = ? AND following_id IN (%s)`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool, len(candidateIDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	return result, rows.Err()
}

func (r *UserRepository) GetFollowersByUserID(userID string) ([]models.User, error) {
	const query = `
		SELECT u.id, u.email, u.password, u.first_name, u.last_name, u.date_of_birth,
		       COALESCE(u.avatar, ''), COALESCE(u.nickname, ''), COALESCE(u.about_me, ''),
		       u.is_public, u.created_at
		FROM followers f
		JOIN users u ON u.id = f.follower_id
		WHERE f.following_id = ?
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	followers := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		var isPublicInt int
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.Avatar,
			&user.Nickname,
			&user.AboutMe,
			&isPublicInt,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}

		user.IsPublic = isPublicInt == 1
		user.Password = ""
		followers = append(followers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return followers, nil
}

func (r *UserRepository) GetFollowingByUserID(userID string) ([]models.User, error) {
	const query = `
		SELECT u.id, u.email, u.password, u.first_name, u.last_name, u.date_of_birth,
		       COALESCE(u.avatar, ''), COALESCE(u.nickname, ''), COALESCE(u.about_me, ''),
		       u.is_public, u.created_at
		FROM followers f
		JOIN users u ON u.id = f.following_id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	following := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		var isPublicInt int
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.DateOfBirth,
			&user.Avatar,
			&user.Nickname,
			&user.AboutMe,
			&isPublicInt,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}

		user.IsPublic = isPublicInt == 1
		user.Password = ""
		following = append(following, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}

func (r *UserRepository) IsFollowing(followerID, followingID string) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM followers
			WHERE follower_id = ? AND following_id = ?
		)
	`

	var exists int
	if err := r.db.QueryRow(query, followerID, followingID).Scan(&exists); err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (r *UserRepository) CreateFollowRequest(request *models.FollowRequest) error {
	const query = `
		INSERT INTO follow_requests (id, sender_id, receiver_id, status, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		request.ID,
		request.SenderID,
		request.ReceiverID,
		request.Status,
		request.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *UserRepository) StartFollowRequest(senderID, receiverID string, autoAccept bool) (*models.FollowRequest, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var isFollowing int
	if err := tx.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ?)`,
		senderID,
		receiverID,
	).Scan(&isFollowing); err != nil {
		return nil, err
	}
	if isFollowing == 1 {
		return nil, ErrAlreadyFollowing
	}

	now := time.Now().UTC()
	targetStatus := "pending"
	if autoAccept {
		targetStatus = "accepted"
	}

	var existing models.FollowRequest
	err = tx.QueryRow(
		`SELECT id, sender_id, receiver_id, status, created_at FROM follow_requests WHERE sender_id = ? AND receiver_id = ? LIMIT 1`,
		senderID,
		receiverID,
	).Scan(
		&existing.ID,
		&existing.SenderID,
		&existing.ReceiverID,
		&existing.Status,
		&existing.CreatedAt,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}

		request := &models.FollowRequest{
			ID:         uuid.NewString(),
			SenderID:   senderID,
			ReceiverID: receiverID,
			Status:     targetStatus,
			CreatedAt:  now,
		}

		if _, err := tx.Exec(
			`INSERT INTO follow_requests (id, sender_id, receiver_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
			request.ID,
			request.SenderID,
			request.ReceiverID,
			request.Status,
			request.CreatedAt.Format(time.RFC3339),
		); err != nil {
			return nil, err
		}
		if autoAccept {
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`,
				senderID,
				receiverID,
			); err != nil {
				return nil, err
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return request, nil
	}

	switch existing.Status {
	case "pending":
		if autoAccept {
			if _, err := tx.Exec(
				`UPDATE follow_requests SET status = 'accepted', created_at = ? WHERE id = ?`,
				now.Format(time.RFC3339),
				existing.ID,
			); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`,
				senderID,
				receiverID,
			); err != nil {
				return nil, err
			}
			existing.Status = "accepted"
			existing.CreatedAt = now
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return &existing, nil
		}
		return nil, ErrFollowRequestAlreadyPending
	case "accepted":
		if autoAccept {
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`,
				senderID,
				receiverID,
			); err != nil {
				return nil, err
			}
			if err := tx.Commit(); err != nil {
				return nil, err
			}
			return &existing, nil
		}

		if _, err := tx.Exec(
			`UPDATE follow_requests SET status = 'pending', created_at = ? WHERE id = ?`,
			now.Format(time.RFC3339),
			existing.ID,
		); err != nil {
			return nil, err
		}
		existing.Status = "pending"
		existing.CreatedAt = now
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &existing, nil
	case "declined":
		if _, err := tx.Exec(
			`UPDATE follow_requests SET status = 'pending', created_at = ? WHERE id = ?`,
			now.Format(time.RFC3339),
			existing.ID,
		); err != nil {
			return nil, err
		}
		existing.Status = "pending"
		existing.CreatedAt = now
		if autoAccept {
			if _, err := tx.Exec(
				`UPDATE follow_requests SET status = 'accepted', created_at = ? WHERE id = ?`,
				now.Format(time.RFC3339),
				existing.ID,
			); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(
				`INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`,
				senderID,
				receiverID,
			); err != nil {
				return nil, err
			}
			existing.Status = "accepted"
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return &existing, nil
	default:
		return nil, errors.New("invalid follow request status")
	}
}

func (r *UserRepository) GetIncomingFollowRequestsByUserID(userID string) ([]models.FollowRequest, error) {
	const query = `
		SELECT id, sender_id, receiver_id, status, created_at
		FROM follow_requests
		WHERE receiver_id = ? AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]models.FollowRequest, 0)
	for rows.Next() {
		var request models.FollowRequest
		if err := rows.Scan(
			&request.ID,
			&request.SenderID,
			&request.ReceiverID,
			&request.Status,
			&request.CreatedAt,
		); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *UserRepository) GetOutgoingFollowRequestsByUserID(userID string) ([]models.FollowRequest, error) {
	const query = `
		SELECT id, sender_id, receiver_id, status, created_at
		FROM follow_requests
		WHERE sender_id = ? AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]models.FollowRequest, 0)
	for rows.Next() {
		var request models.FollowRequest
		if err := rows.Scan(
			&request.ID,
			&request.SenderID,
			&request.ReceiverID,
			&request.Status,
			&request.CreatedAt,
		); err != nil {
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *UserRepository) GetFollowRequestByID(requestID string) (*models.FollowRequest, error) {
	const query = `
		SELECT id, sender_id, receiver_id, status, created_at
		FROM follow_requests
		WHERE id = ?
		LIMIT 1
	`

	var request models.FollowRequest
	if err := r.db.QueryRow(query, requestID).Scan(
		&request.ID,
		&request.SenderID,
		&request.ReceiverID,
		&request.Status,
		&request.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &request, nil
}

func (r *UserRepository) ResolveFollowRequest(requestID, receiverID, newStatus string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var request models.FollowRequest
	if err := tx.QueryRow(
		`SELECT id, sender_id, receiver_id, status, created_at FROM follow_requests WHERE id = ? LIMIT 1`,
		requestID,
	).Scan(
		&request.ID,
		&request.SenderID,
		&request.ReceiverID,
		&request.Status,
		&request.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ErrFollowRequestNotFound
		}
		return err
	}

	if request.ReceiverID != receiverID {
		return ErrFollowRequestForbidden
	}
	if request.Status != "pending" {
		return ErrFollowRequestNotPending
	}

	if newStatus == "accepted" {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO followers (follower_id, following_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`,
			request.SenderID,
			request.ReceiverID,
		); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE follow_requests SET status = ? WHERE id = ?`, newStatus, requestID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CancelOutgoingFollowRequest(requestID, senderID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var request models.FollowRequest
	if err := tx.QueryRow(
		`SELECT id, sender_id, receiver_id, status, created_at FROM follow_requests WHERE id = ? LIMIT 1`,
		requestID,
	).Scan(
		&request.ID,
		&request.SenderID,
		&request.ReceiverID,
		&request.Status,
		&request.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return ErrFollowRequestNotFound
		}
		return err
	}

	if request.SenderID != senderID {
		return ErrFollowRequestForbidden
	}
	if request.Status != "pending" {
		return ErrFollowRequestNotPending
	}

	if _, err := tx.Exec(`DELETE FROM follow_requests WHERE id = ?`, requestID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) CreateNotification(userID, actorID, notifType, payload string) (*models.Notification, error) {
	id := uuid.NewString()
	createdAt := time.Now().UTC()
	const query = `
		INSERT INTO notifications (id, user_id, actor_id, type, payload, is_read, created_at)
		VALUES (?, ?, NULLIF(?, ''), ?, NULLIF(?, ''), 0, ?)
	`

	if _, err := r.db.Exec(
		query,
		id,
		strings.TrimSpace(userID),
		strings.TrimSpace(actorID),
		strings.TrimSpace(notifType),
		strings.TrimSpace(payload),
		createdAt.Format(time.RFC3339),
	); err != nil {
		return nil, err
	}

	n := &models.Notification{
		ID:        id,
		UserID:    strings.TrimSpace(userID),
		ActorID:   strings.TrimSpace(actorID),
		Type:      strings.TrimSpace(notifType),
		Payload:   strings.TrimSpace(payload),
		IsRead:    false,
		CreatedAt: createdAt,
	}
	return n, nil
}

func (r *UserRepository) ListNotificationsByUserID(userID string, limit int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	const query = `
		SELECT id, user_id, COALESCE(actor_id, ''), type, COALESCE(payload, ''), is_read, created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, strings.TrimSpace(userID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Notification, 0, limit)
	for rows.Next() {
		var n models.Notification
		var isRead int
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.ActorID,
			&n.Type,
			&n.Payload,
			&isRead,
			&n.CreatedAt,
		); err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *UserRepository) MarkNotificationRead(userID, notificationID string) (bool, error) {
	const query = `
		UPDATE notifications
		SET is_read = 1
		WHERE id = ? AND user_id = ? AND is_read = 0
	`
	res, err := r.db.Exec(query, strings.TrimSpace(notificationID), strings.TrimSpace(userID))
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *UserRepository) MarkAllNotificationsRead(userID string) (int64, error) {
	const query = `
		UPDATE notifications
		SET is_read = 1
		WHERE user_id = ? AND is_read = 0
	`
	res, err := r.db.Exec(query, strings.TrimSpace(userID))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *UserRepository) CountUnreadNotificationsByUserID(userID string) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ? AND is_read = 0
	`

	var count int
	if err := r.db.QueryRow(query, strings.TrimSpace(userID)).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) RemoveFollower(profileOwnerID, followerUserID string) (bool, error) {
	const query = `DELETE FROM followers WHERE follower_id = ? AND following_id = ?`

	res, err := r.db.Exec(query, followerUserID, profileOwnerID)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *UserRepository) RemoveFollowing(followerUserID, followingUserID string) (bool, error) {
	const query = `DELETE FROM followers WHERE follower_id = ? AND following_id = ?`

	res, err := r.db.Exec(query, followerUserID, followingUserID)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *UserRepository) UpdateUserProfile(userID string, updates UserProfileUpdates) error {
	sets := make([]string, 0, 6)
	args := make([]interface{}, 0, 7)

	if updates.FirstName != nil {
		sets = append(sets, "first_name = ?")
		args = append(args, *updates.FirstName)
	}
	if updates.LastName != nil {
		sets = append(sets, "last_name = ?")
		args = append(args, *updates.LastName)
	}
	if updates.DateOfBirth != nil {
		sets = append(sets, "date_of_birth = ?")
		args = append(args, *updates.DateOfBirth)
	}
	if updates.Avatar != nil {
		sets = append(sets, "avatar = ?")
		args = append(args, nullIfEmpty(*updates.Avatar))
	}
	if updates.Nickname != nil {
		sets = append(sets, "nickname = ?")
		args = append(args, nullIfEmpty(*updates.Nickname))
	}
	if updates.AboutMe != nil {
		sets = append(sets, "about_me = ?")
		args = append(args, nullIfEmpty(*updates.AboutMe))
	}

	if len(sets) == 0 {
		return nil
	}

	query := "UPDATE users SET " + strings.Join(sets, ", ") + " WHERE id = ?"
	args = append(args, userID)

	res, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *UserRepository) UpdateUserPrivacy(userID string, isPublic bool) error {
	const query = `UPDATE users SET is_public = ? WHERE id = ?`

	res, err := r.db.Exec(query, boolToSQLiteInt(isPublic), userID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func boolToSQLiteInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
