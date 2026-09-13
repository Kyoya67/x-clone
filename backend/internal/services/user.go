package services

import (
	"context"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type UserService struct {
	repository UserRepository
}

type UserRepository interface {
	FindOrCreateByOIDC(ctx context.Context, subject, email, displayName string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) FindOrCreateByOIDC(ctx context.Context, subject, email, displayName string) (models.User, error) {
	return s.repository.FindOrCreateByOIDC(ctx, subject, email, displayName)
}

func (s *UserService) FindByID(ctx context.Context, id string) (models.User, error) {
	return s.repository.FindByID(ctx, id)
}
