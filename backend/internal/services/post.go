package services

import (
	"context"
	"strings"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

var ErrInvalidPostContent = apperrors.BadParam.Wrap(nil, "content must be between 1 and 280 characters")

type PostService struct {
	repository PostRepository
}

type PostRepository interface {
	Create(ctx context.Context, authorID, content string) (models.Post, error)
}

func NewPostService(repository PostRepository) *PostService {
	return &PostService{repository: repository}
}

func (s *PostService) Create(ctx context.Context, authorID string, request models.CreatePostRequest) (models.Post, error) {
	content := strings.TrimSpace(request.Content)
	if len([]rune(content)) < 1 || len([]rune(content)) > 280 {
		return models.Post{}, ErrInvalidPostContent
	}
	return s.repository.Create(ctx, authorID, content)
}
