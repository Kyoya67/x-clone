package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeTimelineService struct {
	called bool
	userID string
	feed   models.TimelineFeed
	posts  []models.TimelinePost
	err    error
}

func (f *fakeTimelineService) List(_ context.Context, userID string, feed models.TimelineFeed) ([]models.TimelinePost, error) {
	f.called = true
	f.userID = userID
	f.feed = feed
	return f.posts, f.err
}

func TestTimelineControllerList(t *testing.T) {
	service := &fakeTimelineService{posts: []models.TimelinePost{{ID: "post-1"}}}
	controller := NewTimelineController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/timeline?feed=following", nil))

	controller.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !service.called || service.userID != testUserID || service.feed != models.TimelineFeedFollowing {
		t.Fatalf("unexpected service call: %+v", service)
	}
	var response models.TimelineResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.Posts) != 1 || response.Posts[0].ID != "post-1" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestTimelineControllerListDefaultsToForYou(t *testing.T) {
	service := &fakeTimelineService{}
	controller := NewTimelineController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/timeline", nil))

	controller.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if service.feed != models.TimelineFeedForYou {
		t.Fatalf("expected feed %q, got %q", models.TimelineFeedForYou, service.feed)
	}
}

func TestTimelineControllerListRequiresLogin(t *testing.T) {
	controller := NewTimelineController(&fakeTimelineService{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/timeline", nil)

	controller.List(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestTimelineControllerListReturnsServiceError(t *testing.T) {
	controller := NewTimelineController(&fakeTimelineService{err: errors.New("secret database details")})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/timeline", nil))

	controller.List(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "secret database details") {
		t.Fatal("internal error was exposed")
	}
}

func TestTimelineControllerListHandlesEncodeError(t *testing.T) {
	controller := NewTimelineController(&fakeTimelineService{})
	recorder := &failingResponseWriter{header: http.Header{}}
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/timeline", nil))

	controller.List(recorder, request)

	if recorder.status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.status)
	}
}
