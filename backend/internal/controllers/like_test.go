package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

type fakeLikeService struct {
	likeCalled   bool
	unlikeCalled bool
	userID       string
	postID       string
	err          error
}

func (f *fakeLikeService) Like(_ context.Context, userID, postID string) error {
	f.likeCalled = true
	f.userID = userID
	f.postID = postID
	return f.err
}

func (f *fakeLikeService) Unlike(_ context.Context, userID, postID string) error {
	f.unlikeCalled = true
	f.userID = userID
	f.postID = postID
	return f.err
}

func TestLikeControllerLike(t *testing.T) {
	service := &fakeLikeService{}
	controller := NewLikeController(service)

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPut, "/posts/post-1/like", nil))
	request = mux.SetURLVars(request, map[string]string{"postId": "00000000-0000-0000-0000-000000000002"})

	controller.Like(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !service.likeCalled || service.userID != testUserID || service.postID != "00000000-0000-0000-0000-000000000002" {
		t.Fatalf("unexpected service call: %+v", service)
	}
}

func TestLikeControllerUnlike(t *testing.T) {
	service := &fakeLikeService{}
	controller := NewLikeController(service)

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodDelete, "/posts/post-1/like", nil))
	request = mux.SetURLVars(request, map[string]string{"postId": "00000000-0000-0000-0000-000000000002"})

	controller.Unlike(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if !service.unlikeCalled {
		t.Fatal("service should be called")
	}
}

func TestLikeControllerRejectsInvalidPostID(t *testing.T) {
	controller := NewLikeController(&fakeLikeService{})

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPut, "/posts/nope/like", nil))
	request = mux.SetURLVars(request, map[string]string{"postId": "nope"})

	controller.Like(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestLikeControllerRequiresLogin(t *testing.T) {
	controller := NewLikeController(&fakeLikeService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/posts/post-1/like", nil)
	request = mux.SetURLVars(request, map[string]string{"postId": "00000000-0000-0000-0000-000000000002"})

	controller.Like(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestLikeControllerReturnsServiceError(t *testing.T) {
	controller := NewLikeController(&fakeLikeService{err: errors.New("database secret")})

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPut, "/posts/post-1/like", nil))
	request = mux.SetURLVars(request, map[string]string{"postId": "00000000-0000-0000-0000-000000000002"})

	controller.Like(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
