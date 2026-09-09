package services

import (
	"context"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

var ErrInvalidTimelineFeed = apperrors.BadParam.Wrap(nil, "feed must be for-you or following")

type TimelineRepository interface {
	List(ctx context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error)
}

type TimelineService struct {
	repository TimelineRepository
}

func NewTimelineService(repository TimelineRepository) *TimelineService {
	return &TimelineService{repository: repository}
}

func (s *TimelineService) List(ctx context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error) {
	if feed != models.TimelineFeedForYou && feed != models.TimelineFeedFollowing {
		return nil, ErrInvalidTimelineFeed
	}
	return s.repository.List(ctx, userID, feed)
}
