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

func (r *FollowRepository) ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT followee_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at ASC
	`, followerID)
	if err != nil {
		return nil, classifyPostgresError(err)
	}
	defer rows.Close()

	followeeIDs := make([]string, 0)
	for rows.Next() {
		var followeeID string
		if err := rows.Scan(&followeeID); err != nil {
			return nil, classifyPostgresError(err)
		}
		followeeIDs = append(followeeIDs, followeeID)
	}
	if err := rows.Err(); err != nil {
		return nil, classifyPostgresError(err)
	}

	return followeeIDs, nil
}

func (r *FollowRepository) execute(ctx context.Context, query, followerID, followeeID string) error {
	_, err := r.db.ExecContext(ctx, query, followerID, followeeID)
	if err != nil {
		return classifyPostgresError(err)
	}
	return nil
}
