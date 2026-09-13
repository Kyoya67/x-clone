package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

type LikeService struct {
	repository          LikeRepository
	notificationService *NotificationService
}

type LikeRepository interface {
	Like(ctx context.Context, userID, postID string) (authorID string, err error)
	Unlike(ctx context.Context, userID, postID string) error
}

func NewLikeService(repository LikeRepository, notificationService *NotificationService) *LikeService {
	return &LikeService{repository: repository, notificationService: notificationService}
}

func (s *LikeService) Like(ctx context.Context, userID, postID string) error {
	authorID, err := s.repository.Like(ctx, userID, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound.Wrap(err, "resource not found")
		}
		return err
	}
	return s.notificationService.CreateLike(ctx, authorID, userID, postID)
}

func (s *LikeService) Unlike(ctx context.Context, userID, postID string) error {
	if err := s.repository.Unlike(ctx, userID, postID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound.Wrap(err, "resource not found")
		}
		return err
	}
	return nil
}
