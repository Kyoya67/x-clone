package repositories

import (
	"context"
	"database/sql"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, authorID, content string) (models.Post, error) {
	var post models.Post
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO posts (author_id, content)
		VALUES ($1, $2)
		RETURNING id, author_id, content, created_at
	`, authorID, content).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.CreatedAt,
	)
	if err != nil {
		return models.Post{}, classifyPostgresError(err)
	}
	return post, err
}
