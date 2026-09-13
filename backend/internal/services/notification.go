package services

import (
	"context"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type NotificationService struct {
	repository NotificationRepository
}

type NotificationRepository interface {
	CreateFollow(ctx context.Context, recipientID, actorID string) error
	CreateLike(ctx context.Context, recipientID, actorID, postID string) error
	List(ctx context.Context, recipientID string) ([]models.Notification, error)
}

func NewNotificationService(repository NotificationRepository) *NotificationService {
	return &NotificationService{repository: repository}
}

func (s *NotificationService) CreateFollow(ctx context.Context, recipientID, actorID string) error {
	return s.repository.CreateFollow(ctx, recipientID, actorID)
}

func (s *NotificationService) CreateLike(ctx context.Context, recipientID, actorID, postID string) error {
	return s.repository.CreateLike(ctx, recipientID, actorID, postID)
}

func (s *NotificationService) List(ctx context.Context, recipientID string) ([]models.Notification, error) {
	return s.repository.List(ctx, recipientID)
}
