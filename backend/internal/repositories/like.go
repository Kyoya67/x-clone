package repositories

import (
	"context"
	"database/sql"
)

type LikeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Like(ctx context.Context, userID, postID string) (string, error) {
	var authorID string
	err := r.db.QueryRowContext(ctx, `
		WITH target_post AS (
			SELECT author_id FROM posts WHERE id = $1
		), inserted_like AS (
			INSERT INTO post_likes (post_id, user_id)
			SELECT $1, $2 FROM target_post
			ON CONFLICT DO NOTHING
		)
		SELECT author_id FROM target_post
	`, postID, userID).Scan(&authorID)
	if err != nil {
		return "", classifyPostgresError(err)
	}
	return authorID, nil
}

func (r *LikeRepository) Unlike(ctx context.Context, userID, postID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2
	`, postID, userID)
	if err != nil {
		return classifyPostgresError(err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return classifyPostgresError(err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
