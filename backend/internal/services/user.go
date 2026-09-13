package services

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

var handlePattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,13}$`)

type UserService struct {
	repository UserRepository
}

type UserRepository interface {
	FindOrCreateByOIDC(ctx context.Context, subject, email, displayName string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
	UpdateProfile(ctx context.Context, id string, request models.UpdateUserProfileRequest) (models.User, error)
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

func (s *UserService) UpdateProfile(ctx context.Context, id string, request models.UpdateUserProfileRequest) (models.User, error) {
	request.Handle = strings.TrimSpace(request.Handle)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	request.Bio = strings.TrimSpace(request.Bio)
	if !handlePattern.MatchString(request.Handle) {
		return models.User{}, apperrors.BadParam.Wrap(nil, "handle must be 1 to 13 ASCII letters, numbers, or underscores")
	}
	if request.DisplayName == "" || utf8.RuneCountInString(request.DisplayName) > 100 {
		return models.User{}, apperrors.BadParam.Wrap(nil, "displayName must be between 1 and 100 characters")
	}
	if utf8.RuneCountInString(request.Bio) > 160 {
		return models.User{}, apperrors.BadParam.Wrap(nil, "bio must be 160 characters or less")
	}
	return s.repository.UpdateProfile(ctx, id, request)
}
