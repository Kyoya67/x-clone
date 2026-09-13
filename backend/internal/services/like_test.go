package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

type fakeLikeRepository struct {
	likeCalled   bool
	unlikeCalled bool
	userID       string
	postID       string
	authorID     string
	err          error
}

func (f *fakeLikeRepository) Like(_ context.Context, userID, postID string) (string, error) {
	f.likeCalled = true
	f.userID = userID
	f.postID = postID
	return f.authorID, f.err
}

func (f *fakeLikeRepository) Unlike(_ context.Context, userID, postID string) error {
	f.unlikeCalled = true
	f.userID = userID
	f.postID = postID
	return f.err
}

func TestLikeServiceLike(t *testing.T) {
	repository := &fakeLikeRepository{authorID: "author-1"}
	notificationRepository := &fakeNotificationRepository{}
	service := NewLikeService(repository, NewNotificationService(notificationRepository))

	if err := service.Like(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.likeCalled || repository.userID != "user-1" || repository.postID != "post-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if !notificationRepository.likeCalled || notificationRepository.recipientID != "author-1" || notificationRepository.actorID != "user-1" || notificationRepository.postID != "post-1" {
		t.Fatalf("unexpected notification call: %+v", notificationRepository)
	}
}

func TestLikeServiceLikeReturnsNotFound(t *testing.T) {
	service := NewLikeService(&fakeLikeRepository{err: sql.ErrNoRows}, NewNotificationService(&fakeNotificationRepository{}))

	err := service.Like(context.Background(), "user-1", "missing-post")

	var appErr *apperrors.Error
	if !errors.As(err, &appErr) || appErr.ErrCode != string(apperrors.NotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestLikeServiceLikeReturnsRepositoryError(t *testing.T) {
	expected := errors.New("database failed")
	service := NewLikeService(&fakeLikeRepository{err: expected}, NewNotificationService(&fakeNotificationRepository{}))

	err := service.Like(context.Background(), "user-1", "post-1")
	if !errors.Is(err, expected) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestLikeServiceLikeReturnsNotificationError(t *testing.T) {
	service := NewLikeService(
		&fakeLikeRepository{authorID: "author-1"},
		NewNotificationService(&fakeNotificationRepository{err: errors.New("notification failed")}),
	)

	if err := service.Like(context.Background(), "user-1", "post-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestLikeServiceUnlike(t *testing.T) {
	repository := &fakeLikeRepository{}
	service := NewLikeService(repository, NewNotificationService(&fakeNotificationRepository{}))

	if err := service.Unlike(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.unlikeCalled || repository.userID != "user-1" || repository.postID != "post-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
}

func TestLikeServiceUnlikeReturnsNotFound(t *testing.T) {
	service := NewLikeService(&fakeLikeRepository{err: sql.ErrNoRows}, NewNotificationService(&fakeNotificationRepository{}))

	err := service.Unlike(context.Background(), "user-1", "missing-post")

	var appErr *apperrors.Error
	if !errors.As(err, &appErr) || appErr.ErrCode != string(apperrors.NotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestLikeServiceUnlikeReturnsRepositoryError(t *testing.T) {
	expected := errors.New("database failed")
	service := NewLikeService(&fakeLikeRepository{err: expected}, NewNotificationService(&fakeNotificationRepository{}))

	err := service.Unlike(context.Background(), "user-1", "post-1")
	if !errors.Is(err, expected) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
