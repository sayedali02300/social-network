package repository

import "database/sql"

type LikeRepository struct {
	DB *sql.DB
}

func NewLikeRepository(db *sql.DB) *LikeRepository {
	return &LikeRepository{DB: db}
}

func (r *LikeRepository) UpsertLike(postID, userID string, value int) error {
	_, err := r.DB.Exec(
		`INSERT INTO post_likes (post_id, user_id, value) VALUES (?, ?, ?)
         ON CONFLICT(post_id, user_id) DO UPDATE SET value = excluded.value`,
		postID, userID, value,
	)
	return err
}

func (r *LikeRepository) DeleteLike(postID, userID string) error {
	_, err := r.DB.Exec(`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID)
	return err
}

func (r *LikeRepository) GetLikeCounts(postID string) (likes, dislikes int, err error) {
	row := r.DB.QueryRow(
		`SELECT COALESCE(SUM(CASE WHEN value=1 THEN 1 ELSE 0 END),0),
                COALESCE(SUM(CASE WHEN value=-1 THEN 1 ELSE 0 END),0)
         FROM post_likes WHERE post_id = ?`, postID,
	)
	err = row.Scan(&likes, &dislikes)
	return
}

func (r *LikeRepository) GetMyReaction(postID, userID string) (int, error) {
	var v int
	err := r.DB.QueryRow(
		`SELECT value FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID,
	).Scan(&v)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return v, err
}
