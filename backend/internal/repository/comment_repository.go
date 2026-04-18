package repository

import (
	"database/sql"
	"log"
	"socialnetwork/backend/internal/models"
	"time"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) GetCommentByID(commentID string) (*models.CommentWithAuthor, error) {
	query := `
	SELECT c.id, c.post_id, c.content, c.image_path, COALESCE(c.parent_id,'') AS parent_id, c.created_at,
	       u.id, u.avatar, u.nickname, u.first_name, u.last_name
	FROM comments c
	JOIN users u ON c.user_id = u.id
	WHERE c.id = ?
	LIMIT 1
	`

	row := r.db.QueryRow(query, commentID)

	var c models.CommentWithAuthor
	var imagePath, avatar, nickname, firstName, lastName, userID sql.NullString

	if err := row.Scan(
		&c.ID,
		&c.PostID,
		&c.Content,
		&imagePath,
		&c.ParentID,
		&c.CreatedAt,
		&userID,
		&avatar,
		&nickname,
		&firstName,
		&lastName,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // comment not found
		}
		log.Printf("Row Scan Error: %v", err)
		return nil, err
	}

	if imagePath.Valid {
		c.ImagePath = imagePath.String
	}

	c.Author = models.CommentAuthor{}
	if userID.Valid {
		c.Author.UserID = userID.String
	}
	if avatar.Valid {
		c.Author.Avatar = avatar.String
	}
	if nickname.Valid {
		c.Author.Nickname = nickname.String
	}
	if firstName.Valid {
		c.Author.FirstName = firstName.String
	}
	if lastName.Valid {
		c.Author.LastName = lastName.String
	}

	return &c, nil
}

// GetCommentAuthorAndPostID returns the user_id and post_id for a comment.
func (r *CommentRepository) GetCommentAuthorAndPostID(commentID string) (authorID, postID string, err error) {
	err = r.db.QueryRow(`SELECT user_id, post_id FROM comments WHERE id = ?`, commentID).Scan(&authorID, &postID)
	return
}

func (r *CommentRepository) GetPostIDByCommentID(commentID string) (string, error) {
	const query = `SELECT post_id FROM comments WHERE id = ?`

	var postID string
	if err := r.db.QueryRow(query, commentID).Scan(&postID); err != nil {
		return "", err
	}
	return postID, nil
}

func (r *CommentRepository) GetAllComments(postID string) ([]models.CommentWithAuthor, error) {
	query := `
	SELECT
		c.id, c.post_id, c.content, c.image_path, COALESCE(c.parent_id,'') AS parent_id, c.created_at,
		u.id, u.avatar, u.nickname, u.first_name, u.last_name
	FROM comments c
	JOIN users u ON c.user_id = u.id
	WHERE c.post_id = ?
	ORDER BY c.created_at ASC;
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	comments := make([]models.CommentWithAuthor, 0)

	for rows.Next() {
		var c models.CommentWithAuthor
		var imagePath, avatar, nickname, firstName, lastName, userID sql.NullString

		if err := rows.Scan(
			&c.ID,
			&c.PostID,
			&c.Content,
			&imagePath,
			&c.ParentID,
			&c.CreatedAt,
			&userID,
			&avatar,
			&nickname,
			&firstName,
			&lastName,
		); err != nil {
			log.Printf("Row Scan Error: %v", err)
			return nil, err
		}

		c.ImagePath = ""
		if imagePath.Valid {
			c.ImagePath = imagePath.String
		}

		c.Author = models.CommentAuthor{}

		if userID.Valid {
			c.Author.UserID = userID.String
		}
		if avatar.Valid {
			c.Author.Avatar = avatar.String
		}
		if nickname.Valid {
			c.Author.Nickname = nickname.String
		}
		if firstName.Valid {
			c.Author.FirstName = firstName.String
		}
		if lastName.Valid {
			c.Author.LastName = lastName.String
		}

		comments = append(comments, c)
	}

	return comments, nil
}

func (r *CommentRepository) CreateComment(comment *models.Comment) error {
	query := `
	INSERT INTO comments (id, post_id, user_id, content, image_path, parent_id, created_at)
	VALUES (?,?,?,?,?,NULLIF(?,?),?)
	`
	// Handle Nullable ImagePath
	var image interface{}
	if comment.ImagePath != "" {
		image = comment.ImagePath
	} else {
		image = nil
	}

	_, err := r.db.Exec(query, comment.ID, comment.PostID, comment.UserID, comment.Content, image, comment.ParentID, "", comment.CreatedAt.Format(time.RFC3339))
	return err
}

// GetPostIDByImagePath retrieves the post ID associated with a comment via its image path.
func (r *CommentRepository) GetPostIDByImagePath(imagePath string) (string, error) {
	var postID string
	query := `SELECT post_id FROM comments WHERE image_path = ?`

	err := r.db.QueryRow(query, imagePath).Scan(&postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return postID, nil
}

func (r *CommentRepository) GetCommentOwnerID(commentID string) (string, error) {
	const query = `SELECT user_id FROM comments WHERE id = ?`

	var ownerID string
	if err := r.db.QueryRow(query, commentID).Scan(&ownerID); err != nil {
		return "", err
	}
	return ownerID, nil
}

func (r *CommentRepository) GetCommentImagePath(commentID string) (string, error) {
	const query = `SELECT image_path FROM comments WHERE id = ?`
	var imagePath sql.NullString
	err := r.db.QueryRow(query, commentID).Scan(&imagePath)
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

func (r *CommentRepository) DeleteComment(commentID string) error {
	const query = `DELETE FROM comments WHERE id = ?`
	_, err := r.db.Exec(query, commentID)

	return err
}

func (r *CommentRepository) UpdateComment(commentID, content string) error {
	const query = `
		UPDATE comments
		SET content = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(query, content, commentID)
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
