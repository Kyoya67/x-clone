package repositories

import (
	"context"
	"database/sql"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type TimelineRepository struct {
	db *sql.DB
}

func NewTimelineRepository(db *sql.DB) *TimelineRepository {
	return &TimelineRepository{db: db}
}

func (r *TimelineRepository) List(ctx context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error) {
	query := `
		SELECT posts.id, posts.content, posts.created_at, users.id, users.handle, users.display_name
		FROM posts
		INNER JOIN users ON users.id = posts.author_id
	`
	args := []any{}
	if feed == models.TimelineFeedFollowing {
		query += `
			INNER JOIN follows ON follows.followee_id = posts.author_id
			WHERE follows.follower_id = $1
		`
		args = append(args, userID)
	}
	query += " ORDER BY posts.created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, classifyPostgresError(err)
	}
	defer rows.Close()

	posts := make([]models.TimelinePost, 0)
	for rows.Next() {
		var post models.TimelinePost
		if err := rows.Scan(
			&post.ID,
			&post.Content,
			&post.CreatedAt,
			&post.Author.ID,
			&post.Author.Handle,
			&post.Author.DisplayName,
		); err != nil {
			return nil, classifyPostgresError(err)
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, classifyPostgresError(err)
	}

	return posts, nil
}
