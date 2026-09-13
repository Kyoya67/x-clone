package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

var ErrCannotFollowSelf = apperrors.BadParam.Wrap(nil, "cannot follow yourself")

type FollowService struct {
	repository FollowRepository
}

type FollowRepository interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error)
}

func NewFollowService(repository FollowRepository) *FollowService {
	return &FollowService{repository: repository}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID string) error {
	if err := validateFollowRelation(followerID, followeeID); err != nil {
		return err
	}
	return s.repository.Follow(ctx, followerID, followeeID)
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if err := validateFollowRelation(followerID, followeeID); err != nil {
		return err
	}
	if err := s.repository.Unfollow(ctx, followerID, followeeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound.Wrap(err, "resource not found")
		}
		return err
	}
	return nil
}

func (s *FollowService) ListFolloweeIDs(ctx context.Context, followerID string) ([]string, error) {
	return s.repository.ListFolloweeIDs(ctx, followerID)
}

func validateFollowRelation(followerID, followeeID string) error {
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}
	return nil
}
