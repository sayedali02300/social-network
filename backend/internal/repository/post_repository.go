package repository

import (
	"database/sql"
	"log"
	"socialnetwork/backend/internal/models"
	"time"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) GetSinglePost(postID string, currentUserID string) (*models.Post, error) {
	query := `
		SELECT
			p.id, p.user_id, COALESCE(p.group_id, '') AS group_id, p.title, p.content,
			COALESCE(p.image_path, '') AS image_path,
			p.privacy, p.created_at,
			u.id, u.first_name, u.last_name,
			COALESCE(u.nickname, '') AS nickname,
			COALESCE(u.avatar, '') AS avatar,
			COALESCE(SUM(CASE WHEN pl.value=1 THEN 1 ELSE 0 END),0) AS like_count,
			COALESCE(SUM(CASE WHEN pl.value=-1 THEN 1 ELSE 0 END),0) AS dislike_count,
			COALESCE(MAX(CASE WHEN pl.user_id=? THEN pl.value ELSE 0 END),0) AS my_reaction
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = ?
		LEFT JOIN followers f ON f.following_id = p.user_id AND f.follower_id = ?
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = ?
		LEFT JOIN post_likes pl ON pl.post_id = p.id
		WHERE p.id = ?
		AND (
			(p.group_id IS NOT NULL AND gm.user_id IS NOT NULL)
			OR (
				p.group_id IS NULL AND (
					p.user_id = ?
					OR p.privacy = 'public'
					OR (p.privacy = 'almost_private' AND f.follower_id IS NOT NULL)
					OR (p.privacy = 'private' AND pau.user_id IS NOT NULL)
				)
			)
		)
		GROUP BY p.id`

	var p models.Post
	var nickname, avatar sql.NullString

	err := r.db.QueryRow(query, currentUserID, currentUserID, currentUserID, currentUserID, postID, currentUserID).Scan(
		&p.ID,
		&p.UserID,
		&p.GroupID,
		&p.Title,
		&p.Content,
		&p.ImagePath,
		&p.Privacy,
		&p.CreatedAt,
		&p.Author.ID,
		&p.Author.FirstName,
		&p.Author.LastName,
		&nickname,
		&avatar,
		&p.LikeCount,
		&p.DislikeCount,
		&p.MyReaction,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	p.Author.Nickname = nickname.String
	p.Author.Avatar = avatar.String

	return &p, nil
}

// Updated return type to []models.Post
func (r *PostRepository) GetAllPosts(currentUserID string) ([]models.Post, error) {
	query := `
        SELECT
            p.id, p.user_id, COALESCE(p.group_id, '') AS group_id, p.title, p.content,
            COALESCE(p.image_path, '') AS image_path,
            p.privacy, p.created_at,
            u.id, u.first_name, u.last_name,
            COALESCE(u.nickname, '') AS nickname,
            COALESCE(u.avatar, '') AS avatar,
            COALESCE(SUM(CASE WHEN pl.value=1 THEN 1 ELSE 0 END),0) AS like_count,
            COALESCE(SUM(CASE WHEN pl.value=-1 THEN 1 ELSE 0 END),0) AS dislike_count,
            COALESCE(MAX(CASE WHEN pl.user_id=? THEN pl.value ELSE 0 END),0) AS my_reaction
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN followers f ON f.following_id = p.user_id AND f.follower_id = ?
        LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = ?
        LEFT JOIN post_likes pl ON pl.post_id = p.id
        WHERE
            p.group_id IS NULL
            AND (
                p.user_id = ?
                OR p.privacy = 'public'
                OR (p.privacy = 'almost_private' AND f.follower_id IS NOT NULL)
                OR (p.privacy = 'private' AND pau.user_id IS NOT NULL)
            )
        GROUP BY p.id
        ORDER BY p.created_at DESC;`

	rows, err := r.db.Query(query, currentUserID, currentUserID, currentUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, 0)

	for rows.Next() {
		var p models.Post
		var nickname, avatar sql.NullString

		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.GroupID,
			&p.Title,
			&p.Content,
			&p.ImagePath,
			&p.Privacy,
			&p.CreatedAt,
			&p.Author.ID,
			&p.Author.FirstName,
			&p.Author.LastName,
			&nickname,
			&avatar,
			&p.LikeCount,
			&p.DislikeCount,
			&p.MyReaction,
		)
		if err != nil {
			log.Printf("error scanning post row: %v", err)
			continue
		}

		p.Author.Nickname = nickname.String
		p.Author.Avatar = avatar.String

		posts = append(posts, p)
	}

	return posts, nil
}

// CreatePost inserts a new post into the database.
func (r *PostRepository) CreatePost(post *models.Post) error {
	query := `
		INSERT INTO posts (id, user_id, group_id, title, content, image_path, privacy, created_at)
		VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?)
	`

	// Handle Nullable ImagePath
	var image interface{}
	if post.ImagePath != "" {
		image = post.ImagePath
	} else {
		image = nil
	}

	_, err := r.db.Exec(query, post.ID, post.UserID, post.GroupID, post.Title, post.Content, image, post.Privacy, post.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *PostRepository) GetGroupPosts(groupID, viewerID string) ([]models.Post, error) {
	if !r.isGroupMember(groupID, viewerID) {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT
			p.id, p.user_id, COALESCE(p.group_id, '') AS group_id, p.title, p.content,
			COALESCE(p.image_path, '') AS image_path,
			p.privacy, p.created_at,
			u.id, u.first_name, u.last_name,
			COALESCE(u.nickname, '') AS nickname,
			COALESCE(u.avatar, '') AS avatar
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.group_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]models.Post, 0)
	for rows.Next() {
		var p models.Post
		var nickname, avatar sql.NullString

		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.GroupID,
			&p.Title,
			&p.Content,
			&p.ImagePath,
			&p.Privacy,
			&p.CreatedAt,
			&p.Author.ID,
			&p.Author.FirstName,
			&p.Author.LastName,
			&nickname,
			&avatar,
		); err != nil {
			return nil, err
		}

		p.Author.Nickname = nickname.String
		p.Author.Avatar = avatar.String
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (r *PostRepository) GetPostImagePath(postID string) (string, error) {
	var imagePath sql.NullString

	query := `SELECT image_path FROM posts WHERE id = ?`

	err := r.db.QueryRow(query, postID).Scan(&imagePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if imagePath.Valid {
		return imagePath.String, nil
	}
	return "", nil
}

func (r *PostRepository) GetPostOwnerID(postID string) (string, error) {
	const query = `SELECT user_id FROM posts WHERE id = ?`

	var ownerID string
	if err := r.db.QueryRow(query, postID).Scan(&ownerID); err != nil {
		return "", err
	}
	return ownerID, nil
}

func (r *PostRepository) DeletePost(postID string) error {
	query := `DELETE FROM posts WHERE id = ?`

	_, err := r.db.Exec(query, postID)

	return err
}

func (r *PostRepository) UpdatePost(postID, title, content string) error {
	query := `
		UPDATE posts
		SET title = ?, content = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, title, content, postID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostRepository) AllowFollowers(followers []string, postID string, authorID string) error {

	rows, err := r.db.Query(`
		SELECT follower_id
		FROM followers
		WHERE following_id = ?
	`, authorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// SAVE IN MAP FOR QUICK SEARCH
	validFollowers := make(map[string]bool)

	for rows.Next() {
		var followerID string
		if err := rows.Scan(&followerID); err != nil {
			return err
		}
		validFollowers[followerID] = true
	}

	stmt, err := r.db.Prepare(`
		INSERT INTO post_allowed_users (post_id, user_id)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, followerID := range followers {

		if !validFollowers[followerID] {
			log.Printf("Skipped invalid follower %s for post %s", followerID, postID)
			continue
		}

		_, err := stmt.Exec(postID, followerID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPostByImagePath checks authorization and fetches a post based on its image path.
func (r *PostRepository) GetPostByImagePath(imagePath string, currentUserID string) (*models.Post, error) {
	query := `
		SELECT
			p.id, p.user_id, COALESCE(p.group_id, '') AS group_id, p.title, p.content,
			COALESCE(p.image_path, '') AS image_path,
			p.privacy, p.created_at,
			u.id, u.first_name, u.last_name,
			COALESCE(u.nickname, '') AS nickname,
			COALESCE(u.avatar, '') AS avatar
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN group_members gm ON gm.group_id = p.group_id AND gm.user_id = ?
		LEFT JOIN followers f ON f.following_id = p.user_id AND f.follower_id = ?
		LEFT JOIN post_allowed_users pau ON pau.post_id = p.id AND pau.user_id = ?
		WHERE p.image_path = ?
		AND (
			(p.group_id IS NOT NULL AND gm.user_id IS NOT NULL)
			OR (
				p.group_id IS NULL AND (
					p.user_id = ?
					OR p.privacy = 'public'
					OR (p.privacy = 'almost_private' AND f.follower_id IS NOT NULL)
					OR (p.privacy = 'private' AND pau.user_id IS NOT NULL)
				)
			)
		)`

	var p models.Post
	var nickname, avatar sql.NullString

	err := r.db.QueryRow(query, currentUserID, currentUserID, currentUserID, imagePath, currentUserID).Scan(
		&p.ID,
		&p.UserID,
		&p.GroupID,
		&p.Title,
		&p.Content,
		&p.ImagePath,
		&p.Privacy,
		&p.CreatedAt,
		&p.Author.ID,
		&p.Author.FirstName,
		&p.Author.LastName,
		&nickname,
		&avatar,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	p.Author.Nickname = nickname.String
	p.Author.Avatar = avatar.String

	return &p, nil
}

func (r *PostRepository) isGroupMember(groupID, userID string) bool {
	if groupID == "" || userID == "" {
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

func (r *PostRepository) GetFeedUserIDs(userID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT id FROM users WHERE id != ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostRepository) GetFollowersIDs(userID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT follower_id FROM followers WHERE following_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostRepository) GetAllowedUsers(postID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT user_id FROM post_allowed_users WHERE post_id = ?`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostRepository) GetPostPrivacy(postID string) (string, error) {
	var privacy string
	err := r.db.QueryRow(`SELECT privacy FROM posts WHERE id = ?`, postID).Scan(&privacy)
	return privacy, err
}

// AddAllowedUsers adds new recipients to a private post, skipping duplicates and non-followers.
func (r *PostRepository) AddAllowedUsers(postID, authorID string, userIDs []string) ([]string, error) {
	rows, err := r.db.Query(`SELECT follower_id FROM followers WHERE following_id = ?`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	validFollowers := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		validFollowers[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	existing, err := r.GetAllowedUsers(postID)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]bool, len(existing))
	for _, id := range existing {
		existingSet[id] = true
	}

	stmt, err := r.db.Prepare(`INSERT INTO post_allowed_users (post_id, user_id) VALUES (?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var added []string
	for _, uid := range userIDs {
		if !validFollowers[uid] || existingSet[uid] {
			continue
		}
		if _, err := stmt.Exec(postID, uid); err != nil {
			return nil, err
		}
		added = append(added, uid)
	}

	return added, nil
}

// RemoveAllowedUsers removes recipients from a private post.
func (r *PostRepository) RemoveAllowedUsers(postID string, userIDs []string) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}

	stmt, err := r.db.Prepare(`DELETE FROM post_allowed_users WHERE post_id = ? AND user_id = ?`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var total int64
	for _, uid := range userIDs {
		res, err := stmt.Exec(postID, uid)
		if err != nil {
			return total, err
		}
		n, _ := res.RowsAffected()
		total += n
	}

	return total, nil
}

// RemoveUserFromAllowedPosts removes a follower from all private post_allowed_users entries
// for posts authored by authorID. Called when a follow relationship is removed.
func (r *PostRepository) RemoveUserFromAllowedPosts(authorID, followerID string) error {
	_, err := r.db.Exec(`
		DELETE FROM post_allowed_users
		WHERE user_id = ? AND post_id IN (
			SELECT id FROM posts WHERE user_id = ? AND privacy = 'private'
		)`, followerID, authorID)
	return err
}
