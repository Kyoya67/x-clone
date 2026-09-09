package repositories

import (
	"context"
	"database/sql"
)

type FollowRepository struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Follow(ctx context.Context, followerID, followeeID string) error {
	return r.execute(ctx, `
		INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`, followerID, followeeID)
}

func (r *FollowRepository) Unfollow(ctx context.Context, followerID, followeeID string) error {
	return r.execute(ctx, `
		DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2
	`, followerID, followeeID)
}

func (r *FollowRepository) execute(ctx context.Context, query, followerID, followeeID string) error {
	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	if err != nil {
		return classifyPostgresError(err)
	}
	return nil
}
