package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeTimelineRepository struct {
	called bool
	userID string
	feed   models.TimelineFeed
	posts  []models.TimelinePost
	err    error
}

func (f *fakeTimelineRepository) List(_ context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error) {
	f.called = true
	f.userID = userID
	f.feed = feed
	return f.posts, f.err
}

func TestTimelineServiceList(t *testing.T) {
	repository := &fakeTimelineRepository{posts: []models.TimelinePost{{ID: "post-1"}}}
	service := NewTimelineService(repository)

	posts, err := service.List(context.Background(), "user-1", models.TimelineFeedFollowing)
	if err != nil {
		t.Fatal(err)
	}
	if !repository.called || repository.userID != "user-1" || repository.feed != models.TimelineFeedFollowing {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if len(posts) != 1 || posts[0].ID != "post-1" {
		t.Fatalf("unexpected posts: %+v", posts)
	}
}

func TestTimelineServiceListRejectsInvalidFeed(t *testing.T) {
	repository := &fakeTimelineRepository{}
	service := NewTimelineService(repository)

	_, err := service.List(context.Background(), "user-1", "invalid")
	if !errors.Is(err, ErrInvalidTimelineFeed) {
		t.Fatalf("expected invalid feed error, got %v", err)
	}
	if repository.called {
		t.Fatal("repository should not be called")
	}
}
