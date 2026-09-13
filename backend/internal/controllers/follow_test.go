package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/gorilla/mux"
)

const followeeID = "00000000-0000-0000-0000-000000000002"

type fakeFollowService struct {
	followCalled   bool
	unfollowCalled bool
	listCalled     bool
	followerID     string
	followeeID     string
	followeeIDs    []string
	err            error
}

func (f *fakeFollowService) ListFolloweeIDs(_ context.Context, followerID string) ([]string, error) {
	f.listCalled = true
	f.followerID = followerID
	return f.followeeIDs, f.err
}

func (f *fakeFollowService) Follow(_ context.Context, followerID, followeeID string) error {
	f.followCalled = true
	f.followerID = followerID
	f.followeeID = followeeID
	return f.err
}

func (f *fakeFollowService) Unfollow(_ context.Context, followerID, followeeID string) error {
	f.unfollowCalled = true
	f.followerID = followerID
	f.followeeID = followeeID
	return f.err
}

func TestFollowControllerFollow(t *testing.T) {
	service := &fakeFollowService{}
	controller := NewFollowController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodPut, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Follow(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !service.followCalled || service.followerID != testUserID || service.followeeID != followeeID {
		t.Fatalf("unexpected service call: %+v", service)
	}
}

func TestFollowControllerUnfollow(t *testing.T) {
	service := &fakeFollowService{}
	controller := NewFollowController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodDelete, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Unfollow(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !service.unfollowCalled || service.followerID != testUserID || service.followeeID != followeeID {
		t.Fatalf("unexpected service call: %+v", service)
	}
}

func TestFollowControllerListFollowing(t *testing.T) {
	service := &fakeFollowService{followeeIDs: []string{followeeID}}
	controller := NewFollowController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/me/following", nil))

	controller.ListFollowing(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !service.listCalled || service.followerID != testUserID {
		t.Fatalf("unexpected service call: %+v", service)
	}
	var response struct {
		UserIDs []string `json:"userIds"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response.UserIDs) != 1 || response.UserIDs[0] != followeeID {
		t.Fatalf("unexpected response: %v", response.UserIDs)
	}
}

func TestFollowControllerListFollowingReturnsServiceError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{err: errors.New("secret database details")})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/me/following", nil))

	controller.ListFollowing(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "secret database details") {
		t.Fatal("internal error was exposed")
	}
}

func TestFollowControllerListFollowingRequiresLogin(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/me/following", nil)

	controller.ListFollowing(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestFollowControllerListFollowingHandlesEncodeError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{})
	recorder := &failingResponseWriter{header: http.Header{}}
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/me/following", nil))

	controller.ListFollowing(recorder, request)

	if recorder.status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.status)
	}
}

func TestFollowControllerRejectsInvalidUserID(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodPut, "/users/not-a-uuid/follow", nil), "not-a-uuid"))

	controller.Follow(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestFollowControllerUnfollowRejectsInvalidUserID(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodDelete, "/users/not-a-uuid/follow", nil), "not-a-uuid"))

	controller.Unfollow(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestFollowControllerReturnsServiceError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{err: errors.New("secret database details")})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodPut, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Follow(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "secret database details") {
		t.Fatal("internal error was exposed")
	}
}

func TestFollowControllerFollowRequiresLogin(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{})
	recorder := httptest.NewRecorder()
	request := withUserID(httptest.NewRequest(http.MethodPut, "/users/"+followeeID+"/follow", nil), followeeID)

	controller.Follow(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestFollowControllerUnfollowReturnsServiceError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{err: errors.New("secret database details")})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodDelete, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Unfollow(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "secret database details") {
		t.Fatal("internal error was exposed")
	}
}

func TestFollowControllerFollowReturnsValidationError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{
		err: apperrors.BadParam.Wrap(nil, "cannot follow yourself"),
	})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodPut, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Follow(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestFollowControllerUnfollowReturnsValidationError(t *testing.T) {
	controller := NewFollowController(&fakeFollowService{
		err: apperrors.BadParam.Wrap(nil, "cannot follow yourself"),
	})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(withUserID(httptest.NewRequest(http.MethodDelete, "/users/"+followeeID+"/follow", nil), followeeID))

	controller.Unfollow(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func withUserID(request *http.Request, userID string) *http.Request {
	return mux.SetURLVars(request, map[string]string{"userId": userID})
}

type failingResponseWriter struct {
	header http.Header
	status int
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(status int) {
	w.status = status
}
