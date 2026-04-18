package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"socialnetwork/backend/internal/models"
)

type GroupRepository struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateGroup(creatorID, title, description string) (*models.Group, error) {
	group := &models.Group{
		ID:          uuid.NewString(),
		CreatorID:   creatorID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(
		`INSERT INTO groups (id, creator_id, title, description, created_at) VALUES (?, ?, ?, ?, ?)`,
		group.ID,
		group.CreatorID,
		group.Title,
		group.Description,
		group.CreatedAt.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if _, err := tx.Exec(
		`INSERT INTO group_members (group_id, user_id, role, joined_at) VALUES (?, ?, 'creator', ?)`,
		group.ID,
		group.CreatorID,
		group.CreatedAt.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return group, nil
}

func (r *GroupRepository) ListGroups(viewerID string) ([]models.GroupSummary, error) {
	const query = `
		SELECT
			g.id,
			g.creator_id,
			COALESCE(u.nickname, u.first_name || ' ' || u.last_name) AS creator_name,
			g.title,
			COALESCE(g.description, '') AS description,
			(SELECT COUNT(*) FROM group_members gm_count WHERE gm_count.group_id = g.id) AS members_count,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_members gm_me
				WHERE gm_me.group_id = g.id AND gm_me.user_id = ?
			) THEN 1 ELSE 0 END AS is_member,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_invites gi
				WHERE gi.group_id = g.id AND gi.receiver_id = ? AND gi.status = 'pending'
			) THEN 1 ELSE 0 END AS has_pending_invite,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_join_requests gjr
				WHERE gjr.group_id = g.id AND gjr.user_id = ? AND gjr.status = 'pending'
			) THEN 1 ELSE 0 END AS has_pending_request,
			g.created_at
		FROM groups g
		JOIN users u ON u.id = g.creator_id
		ORDER BY g.created_at DESC`

	rows, err := r.db.Query(query, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]models.GroupSummary, 0)
	for rows.Next() {
		var group models.GroupSummary
		var isMember int
		var hasPendingInvite int
		var hasPendingRequest int

		if err := rows.Scan(
			&group.ID,
			&group.CreatorID,
			&group.CreatorName,
			&group.Title,
			&group.Description,
			&group.MembersCount,
			&isMember,
			&hasPendingInvite,
			&hasPendingRequest,
			&group.CreatedAt,
		); err != nil {
			return nil, err
		}

		group.IsMember = isMember == 1
		group.HasPendingInvite = hasPendingInvite == 1
		group.HasPendingRequest = hasPendingRequest == 1
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

func (r *GroupRepository) GetGroup(groupID, viewerID string) (*models.GroupSummary, error) {
	const query = `
		SELECT
			g.id,
			g.creator_id,
			COALESCE(u.nickname, u.first_name || ' ' || u.last_name) AS creator_name,
			g.title,
			COALESCE(g.description, '') AS description,
			(SELECT COUNT(*) FROM group_members gm_count WHERE gm_count.group_id = g.id) AS members_count,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_members gm_me
				WHERE gm_me.group_id = g.id AND gm_me.user_id = ?
			) THEN 1 ELSE 0 END AS is_member,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_invites gi
				WHERE gi.group_id = g.id AND gi.receiver_id = ? AND gi.status = 'pending'
			) THEN 1 ELSE 0 END AS has_pending_invite,
			CASE WHEN EXISTS (
				SELECT 1 FROM group_join_requests gjr
				WHERE gjr.group_id = g.id AND gjr.user_id = ? AND gjr.status = 'pending'
			) THEN 1 ELSE 0 END AS has_pending_request,
			g.created_at
		FROM groups g
		JOIN users u ON u.id = g.creator_id
		WHERE g.id = ?
		LIMIT 1`

	var group models.GroupSummary
	var isMember int
	var hasPendingInvite int
	var hasPendingRequest int

	err := r.db.QueryRow(query, viewerID, viewerID, viewerID, groupID).Scan(
		&group.ID,
		&group.CreatorID,
		&group.CreatorName,
		&group.Title,
		&group.Description,
		&group.MembersCount,
		&isMember,
		&hasPendingInvite,
		&hasPendingRequest,
		&group.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	group.IsMember = isMember == 1
	group.HasPendingInvite = hasPendingInvite == 1
	group.HasPendingRequest = hasPendingRequest == 1

	return &group, nil
}

func (r *GroupRepository) UpdateGroup(groupID, actorID, title, description string) (*models.Group, error) {
	if !r.isGroupCreator(groupID, actorID) {
		return nil, errors.New("only the group creator can update this group")
	}

	if _, err := r.db.Exec(
		`UPDATE groups SET title = ?, description = ? WHERE id = ?`,
		title,
		description,
		groupID,
	); err != nil {
		return nil, err
	}

	var group models.Group
	err := r.db.QueryRow(
		`SELECT id, creator_id, title, COALESCE(description, ''), created_at FROM groups WHERE id = ?`,
		groupID,
	).Scan(&group.ID, &group.CreatorID, &group.Title, &group.Description, &group.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *GroupRepository) DeleteGroup(groupID, actorID string) error {
	if !r.isGroupCreator(groupID, actorID) {
		return errors.New("only the group creator can delete this group")
	}

	result, err := r.db.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *GroupRepository) CreateInvite(groupID, senderID, receiverID string) (*models.GroupInvite, error) {
	if !r.isGroupMember(groupID, senderID) {
		return nil, errors.New("only group members can send invites")
	}
	if !r.isFollowerOf(receiverID, senderID) {
		return nil, errors.New("only your followers can be invited to the group")
	}
	if r.isGroupMember(groupID, receiverID) {
		return nil, errors.New("user is already a group member")
	}
	if r.hasPendingJoinRequest(groupID, receiverID) {
		return nil, errors.New("this user already has a pending join request")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var existing models.GroupInvite
	err = tx.QueryRow(
		`SELECT id, group_id, sender_id, receiver_id, status, created_at FROM group_invites WHERE group_id = ? AND receiver_id = ?`,
		groupID,
		receiverID,
	).Scan(&existing.ID, &existing.GroupID, &existing.SenderID, &existing.ReceiverID, &existing.Status, &existing.CreatedAt)
	switch {
	case err == nil:
		if existing.Status == "pending" {
			_ = tx.Rollback()
			return nil, errors.New("this user already has a pending group invite")
		}

		existing.SenderID = senderID
		existing.Status = "pending"
		existing.CreatedAt = time.Now().UTC()

		if _, err := tx.Exec(
			`UPDATE group_invites SET sender_id = ?, status = ?, created_at = ? WHERE id = ?`,
			existing.SenderID,
			existing.Status,
			existing.CreatedAt.Format(time.RFC3339),
			existing.ID,
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		return &existing, nil
	case err != sql.ErrNoRows:
		_ = tx.Rollback()
		return nil, err
	}

	invite := &models.GroupInvite{
		ID:         uuid.NewString(),
		GroupID:    groupID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Status:     "pending",
		CreatedAt:  time.Now().UTC(),
	}

	if _, err := tx.Exec(
		`INSERT INTO group_invites (id, group_id, sender_id, receiver_id, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		invite.ID,
		invite.GroupID,
		invite.SenderID,
		invite.ReceiverID,
		invite.Status,
		invite.CreatedAt.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return invite, nil
}

func (r *GroupRepository) ListInvites(groupID, actorID string) ([]models.GroupInvite, error) {
	const query = `
		SELECT id, group_id, sender_id, receiver_id, status, created_at
		FROM group_invites
		WHERE group_id = ? AND (receiver_id = ? OR ? IN (
			SELECT user_id FROM group_members WHERE group_id = ?
		))
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, groupID, actorID, actorID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invites := make([]models.GroupInvite, 0)
	for rows.Next() {
		var invite models.GroupInvite
		if err := rows.Scan(&invite.ID, &invite.GroupID, &invite.SenderID, &invite.ReceiverID, &invite.Status, &invite.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}

	return invites, rows.Err()
}

func (r *GroupRepository) GetInviteByID(inviteID string) (*models.GroupInvite, error) {
	var invite models.GroupInvite
	err := r.db.QueryRow(
		`SELECT id, group_id, sender_id, receiver_id, status, created_at FROM group_invites WHERE id = ?`,
		strings.TrimSpace(inviteID),
	).Scan(&invite.ID, &invite.GroupID, &invite.SenderID, &invite.ReceiverID, &invite.Status, &invite.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *GroupRepository) RespondInvite(inviteID, actorID, status string) (*models.GroupInvite, error) {
	if status != "accepted" && status != "declined" {
		return nil, errors.New("status must be accepted or declined")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var invite models.GroupInvite
	if err := tx.QueryRow(
		`SELECT id, group_id, sender_id, receiver_id, status, created_at FROM group_invites WHERE id = ?`,
		inviteID,
	).Scan(&invite.ID, &invite.GroupID, &invite.SenderID, &invite.ReceiverID, &invite.Status, &invite.CreatedAt); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if invite.ReceiverID != actorID {
		_ = tx.Rollback()
		return nil, errors.New("only the invited user can respond")
	}
	if invite.Status != "pending" {
		_ = tx.Rollback()
		return nil, errors.New("invite has already been handled")
	}

	if _, err := tx.Exec(`UPDATE group_invites SET status = ? WHERE id = ?`, status, inviteID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if status == "accepted" {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO group_members (group_id, user_id, role, joined_at) VALUES (?, ?, 'member', ?)`,
			invite.GroupID,
			actorID,
			time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if _, err := tx.Exec(
			`UPDATE group_join_requests SET status = 'accepted' WHERE group_id = ? AND user_id = ? AND status = 'pending'`,
			invite.GroupID,
			actorID,
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	invite.Status = status
	return &invite, nil
}

func (r *GroupRepository) CreateJoinRequest(groupID, actorID string) (*models.GroupJoinRequest, error) {
	if r.isGroupMember(groupID, actorID) {
		return nil, errors.New("user is already a group member")
	}
	if r.hasPendingInvite(groupID, actorID) {
		return nil, errors.New("user already has a pending group invite")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var existing models.GroupJoinRequest
	err = tx.QueryRow(
		`SELECT id, group_id, user_id, status, created_at FROM group_join_requests WHERE group_id = ? AND user_id = ?`,
		groupID,
		actorID,
	).Scan(&existing.ID, &existing.GroupID, &existing.UserID, &existing.Status, &existing.CreatedAt)
	switch {
	case err == nil:
		if existing.Status == "pending" {
			_ = tx.Rollback()
			return nil, errors.New("you already have a pending join request for this group")
		}

		existing.Status = "pending"
		existing.CreatedAt = time.Now().UTC()
		if _, err := tx.Exec(
			`UPDATE group_join_requests SET status = ?, created_at = ? WHERE id = ?`,
			existing.Status,
			existing.CreatedAt.Format(time.RFC3339),
			existing.ID,
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		return &existing, nil
	case err != sql.ErrNoRows:
		_ = tx.Rollback()
		return nil, err
	}

	request := &models.GroupJoinRequest{
		ID:        uuid.NewString(),
		GroupID:   groupID,
		UserID:    actorID,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	if _, err := tx.Exec(
		`INSERT INTO group_join_requests (id, group_id, user_id, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		request.ID,
		request.GroupID,
		request.UserID,
		request.Status,
		request.CreatedAt.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return request, nil
}

func (r *GroupRepository) ListJoinRequests(groupID, actorID string) ([]models.GroupJoinRequest, error) {
	if !r.isGroupCreator(groupID, actorID) {
		return nil, errors.New("only the group creator can see join requests")
	}

	rows, err := r.db.Query(
		`SELECT id, group_id, user_id, status, created_at FROM group_join_requests WHERE group_id = ? ORDER BY created_at DESC`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]models.GroupJoinRequest, 0)
	for rows.Next() {
		var request models.GroupJoinRequest
		if err := rows.Scan(&request.ID, &request.GroupID, &request.UserID, &request.Status, &request.CreatedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, rows.Err()
}

func (r *GroupRepository) GetJoinRequestByID(requestID string) (*models.GroupJoinRequest, error) {
	var request models.GroupJoinRequest
	err := r.db.QueryRow(
		`SELECT id, group_id, user_id, status, created_at FROM group_join_requests WHERE id = ?`,
		strings.TrimSpace(requestID),
	).Scan(&request.ID, &request.GroupID, &request.UserID, &request.Status, &request.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *GroupRepository) RespondJoinRequest(requestID, actorID, status string) (*models.GroupJoinRequest, error) {
	if status != "accepted" && status != "declined" {
		return nil, errors.New("status must be accepted or declined")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	var request models.GroupJoinRequest
	if err := tx.QueryRow(
		`SELECT id, group_id, user_id, status, created_at FROM group_join_requests WHERE id = ?`,
		requestID,
	).Scan(&request.ID, &request.GroupID, &request.UserID, &request.Status, &request.CreatedAt); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	var creatorID string
	if err := tx.QueryRow(`SELECT creator_id FROM groups WHERE id = ?`, request.GroupID).Scan(&creatorID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if creatorID != actorID {
		_ = tx.Rollback()
		return nil, errors.New("only the group creator can respond")
	}
	if request.Status != "pending" {
		_ = tx.Rollback()
		return nil, errors.New("request has already been handled")
	}

	if _, err := tx.Exec(`UPDATE group_join_requests SET status = ? WHERE id = ?`, status, requestID); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if status == "accepted" {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO group_members (group_id, user_id, role, joined_at) VALUES (?, ?, 'member', ?)`,
			request.GroupID,
			request.UserID,
			time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if _, err := tx.Exec(
			`UPDATE group_invites SET status = 'accepted' WHERE group_id = ? AND receiver_id = ? AND status = 'pending'`,
			request.GroupID,
			request.UserID,
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	request.Status = status
	return &request, nil
}

func (r *GroupRepository) ListMembers(groupID, actorID string) ([]models.GroupMember, error) {
	if !r.isGroupMember(groupID, actorID) {
		return nil, errors.New("only group members can view group members")
	}

	const query = `
		SELECT
			gm.user_id,
			COALESCE(u.nickname, u.first_name || ' ' || u.last_name) AS nickname,
			COALESCE(u.avatar, '') AS avatar,
			gm.role,
			gm.joined_at
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ?
		ORDER BY CASE gm.role WHEN 'creator' THEN 0 ELSE 1 END, gm.joined_at ASC`

	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]models.GroupMember, 0)
	for rows.Next() {
		var member models.GroupMember
		if err := rows.Scan(&member.UserID, &member.Nickname, &member.Avatar, &member.Role, &member.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

func (r *GroupRepository) MemberIDs(groupID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT user_id FROM group_members WHERE group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		if strings.TrimSpace(userID) != "" {
			ids = append(ids, userID)
		}
	}
	return ids, rows.Err()
}

func (r *GroupRepository) RemoveMember(groupID, targetUserID, actorID string) error {
	if targetUserID == "" {
		return errors.New("target user is required")
	}

	isSelf := targetUserID == actorID
	if !isSelf && !r.isGroupCreator(groupID, actorID) {
		return errors.New("only the group creator can remove members")
	}

	var role string
	if err := r.db.QueryRow(
		`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`,
		groupID,
		targetUserID,
	).Scan(&role); err != nil {
		return err
	}

	if role == "creator" && !isSelf {
		return errors.New("the creator cannot be removed from the group")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if role == "creator" {
		var nextCreatorID string
		err := tx.QueryRow(
			`SELECT user_id
			 FROM group_members
			 WHERE group_id = ? AND user_id != ?
			 ORDER BY joined_at ASC
			 LIMIT 1`,
			groupID,
			targetUserID,
		).Scan(&nextCreatorID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if errors.Is(err, sql.ErrNoRows) {
			result, deleteErr := tx.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
			if deleteErr != nil {
				return deleteErr
			}
			affected, rowsErr := result.RowsAffected()
			if rowsErr != nil {
				return rowsErr
			}
			if affected == 0 {
				return sql.ErrNoRows
			}

			return tx.Commit()
		}

		if _, err := tx.Exec(
			`UPDATE group_members SET role = 'creator' WHERE group_id = ? AND user_id = ?`,
			groupID,
			nextCreatorID,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(`UPDATE groups SET creator_id = ? WHERE id = ?`, nextCreatorID, groupID); err != nil {
			return err
		}
	}

	result, err := tx.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, targetUserID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (r *GroupRepository) DeleteEvent(eventID, actorID string) error {
	var groupID string
	var creatorID string
	if err := r.db.QueryRow(`SELECT group_id, creator_id FROM events WHERE id = ?`, eventID).Scan(&groupID, &creatorID); err != nil {
		return err
	}

	if actorID != creatorID && !r.isGroupCreator(groupID, actorID) {
		return errors.New("only the event creator or group creator can delete this event")
	}

	result, err := r.db.Exec(`DELETE FROM events WHERE id = ?`, eventID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *GroupRepository) CreateEvent(groupID, actorID, title, description string, eventTime time.Time) (*models.Event, error) {
	if !r.isGroupMember(groupID, actorID) {
		return nil, errors.New("only group members can create events")
	}

	event := &models.Event{
		ID:          uuid.NewString(),
		GroupID:     groupID,
		CreatorID:   actorID,
		Title:       title,
		Description: description,
		EventTime:   eventTime.UTC(),
		CreatedAt:   time.Now().UTC(),
	}

	_, err := r.db.Exec(
		`INSERT INTO events (id, group_id, creator_id, title, description, event_time, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.ID,
		event.GroupID,
		event.CreatorID,
		event.Title,
		event.Description,
		event.EventTime.Format(time.RFC3339),
		event.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *GroupRepository) ListEvents(groupID, viewerID string) ([]models.Event, error) {
	if !r.isGroupMember(groupID, viewerID) {
		return nil, errors.New("only group members can see group events")
	}

	const query = `
		SELECT
			e.id,
			e.group_id,
			e.creator_id,
			e.title,
			COALESCE(e.description, '') AS description,
			e.event_time,
			e.created_at,
			COALESCE(SUM(CASE WHEN er.response = 'going' THEN 1 ELSE 0 END), 0) AS going_count,
			COALESCE(SUM(CASE WHEN er.response = 'not_going' THEN 1 ELSE 0 END), 0) AS not_going_count,
			COALESCE(MAX(CASE WHEN er.user_id = ? THEN er.response ELSE '' END), '') AS my_response
		FROM events e
		LEFT JOIN event_responses er ON er.event_id = e.id
		WHERE e.group_id = ?
		GROUP BY e.id
		ORDER BY e.event_time ASC`

	rows, err := r.db.Query(query, viewerID, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.Event, 0)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&event.EventTime,
			&event.CreatedAt,
			&event.GoingCount,
			&event.NotGoingCount,
			&event.MyResponse,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (r *GroupRepository) GetEvent(eventID, viewerID string) (*models.Event, error) {
	const query = `
		SELECT
			e.id,
			e.group_id,
			e.creator_id,
			e.title,
			COALESCE(e.description, '') AS description,
			e.event_time,
			e.created_at,
			COALESCE(SUM(CASE WHEN er.response = 'going' THEN 1 ELSE 0 END), 0) AS going_count,
			COALESCE(SUM(CASE WHEN er.response = 'not_going' THEN 1 ELSE 0 END), 0) AS not_going_count,
			COALESCE(MAX(CASE WHEN er.user_id = ? THEN er.response ELSE '' END), '') AS my_response
		FROM events e
		LEFT JOIN event_responses er ON er.event_id = e.id
		WHERE e.id = ?
		GROUP BY e.id`

	var event models.Event
	err := r.db.QueryRow(query, viewerID, eventID).Scan(
		&event.ID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&event.EventTime,
		&event.CreatedAt,
		&event.GoingCount,
		&event.NotGoingCount,
		&event.MyResponse,
	)
	if err != nil {
		return nil, err
	}

	if !r.isGroupMember(event.GroupID, viewerID) {
		return nil, errors.New("only group members can see this event")
	}

	return &event, nil
}

func (r *GroupRepository) UpdateEvent(eventID, actorID, title, description string, eventTime time.Time) (*models.Event, error) {
	var groupID string
	var creatorID string
	if err := r.db.QueryRow(`SELECT group_id, creator_id FROM events WHERE id = ?`, eventID).Scan(&groupID, &creatorID); err != nil {
		return nil, err
	}

	if actorID != creatorID && !r.isGroupCreator(groupID, actorID) {
		return nil, errors.New("only the event creator or group creator can update this event")
	}

	if _, err := r.db.Exec(
		`UPDATE events SET title = ?, description = ?, event_time = ? WHERE id = ?`,
		title,
		description,
		eventTime.UTC().Format(time.RFC3339),
		eventID,
	); err != nil {
		return nil, err
	}

	return r.GetEvent(eventID, actorID)
}

func (r *GroupRepository) UpsertEventResponse(eventID, actorID, response string) (*models.EventResponse, error) {
	if response != "going" && response != "not_going" {
		return nil, errors.New("response must be going or not_going")
	}

	var groupID string
	if err := r.db.QueryRow(`SELECT group_id FROM events WHERE id = ?`, eventID).Scan(&groupID); err != nil {
		return nil, err
	}

	if !r.isGroupMember(groupID, actorID) {
		return nil, errors.New("only group members can RSVP to this event")
	}

	resp := &models.EventResponse{
		EventID:   eventID,
		UserID:    actorID,
		Response:  response,
		CreatedAt: time.Now().UTC(),
	}

	_, err := r.db.Exec(
		`INSERT INTO event_responses (event_id, user_id, response, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(event_id, user_id)
		 DO UPDATE SET response = excluded.response, created_at = excluded.created_at`,
		resp.EventID,
		resp.UserID,
		resp.Response,
		resp.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *GroupRepository) isGroupMember(groupID, userID string) bool {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(userID) == "" {
		return false
	}

	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM group_members WHERE group_id = ? AND user_id = ? LIMIT 1`,
		groupID,
		userID,
	).Scan(&exists)

	return err == nil && exists == 1
}

func (r *GroupRepository) isGroupCreator(groupID, userID string) bool {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(userID) == "" {
		return false
	}

	var creatorID string
	err := r.db.QueryRow(`SELECT creator_id FROM groups WHERE id = ?`, groupID).Scan(&creatorID)
	return err == nil && creatorID == userID
}

func (r *GroupRepository) isFollowerOf(followerID, followingID string) bool {
	if strings.TrimSpace(followerID) == "" || strings.TrimSpace(followingID) == "" {
		return false
	}

	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM followers WHERE follower_id = ? AND following_id = ? LIMIT 1`,
		followerID,
		followingID,
	).Scan(&exists)

	return err == nil && exists == 1
}

func (r *GroupRepository) hasPendingInvite(groupID, receiverID string) bool {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM group_invites WHERE group_id = ? AND receiver_id = ? AND status = 'pending' LIMIT 1`,
		groupID,
		receiverID,
	).Scan(&exists)

	return err == nil && exists == 1
}

func (r *GroupRepository) hasPendingJoinRequest(groupID, userID string) bool {
	var exists int
	err := r.db.QueryRow(
		`SELECT 1 FROM group_join_requests WHERE group_id = ? AND user_id = ? AND status = 'pending' LIMIT 1`,
		groupID,
		userID,
	).Scan(&exists)

	return err == nil && exists == 1
}

// DueEvent holds the minimal fields needed to fire a due notification.
type DueEvent struct {
	ID        string
	GroupID   string
	CreatorID string
}

// FetchDueUnnotifiedEvents returns all events whose event_time has passed
// and for which a due notification has not yet been sent.
func (r *GroupRepository) FetchDueUnnotifiedEvents() ([]DueEvent, error) {
	rows, err := r.db.Query(
		`SELECT id, group_id, creator_id FROM events
		 WHERE event_time <= ? AND due_notified = 0`,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []DueEvent
	for rows.Next() {
		var e DueEvent
		if err := rows.Scan(&e.ID, &e.GroupID, &e.CreatorID); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// MarkEventDueNotified marks an event so the due notification is not sent again.
func (r *GroupRepository) MarkEventDueNotified(eventID string) error {
	_, err := r.db.Exec(`UPDATE events SET due_notified = 1 WHERE id = ?`, eventID)
	return err
}

func IsConflictError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}

func WrapRepositoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return fmt.Errorf("repository: %w", err)
}
