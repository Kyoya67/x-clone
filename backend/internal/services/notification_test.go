package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeNotificationRepository struct {
	followCalled  bool
	likeCalled    bool
	listCalled    bool
	recipientID   string
	actorID       string
	postID        string
	notifications []models.Notification
	err           error
}

func (f *fakeNotificationRepository) CreateFollow(_ context.Context, recipientID, actorID string) error {
	f.followCalled = true
	f.recipientID = recipientID
	f.actorID = actorID
	return f.err
}

func (f *fakeNotificationRepository) CreateLike(_ context.Context, recipientID, actorID, postID string) error {
	f.likeCalled = true
	f.recipientID = recipientID
	f.actorID = actorID
	f.postID = postID
	return f.err
}

func (f *fakeNotificationRepository) List(_ context.Context, recipientID string) ([]models.Notification, error) {
	f.listCalled = true
	f.recipientID = recipientID
	return f.notifications, f.err
}

func TestNotificationServiceCreateFollow(t *testing.T) {
	repository := &fakeNotificationRepository{}
	service := NewNotificationService(repository)

	if err := service.CreateFollow(context.Background(), "recipient-1", "actor-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.followCalled || repository.recipientID != "recipient-1" || repository.actorID != "actor-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
}

func TestNotificationServiceCreateLike(t *testing.T) {
	repository := &fakeNotificationRepository{}
	service := NewNotificationService(repository)

	if err := service.CreateLike(context.Background(), "recipient-1", "actor-1", "post-1"); err != nil {
		t.Fatal(err)
	}
	if !repository.likeCalled || repository.recipientID != "recipient-1" || repository.actorID != "actor-1" || repository.postID != "post-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
}

func TestNotificationServiceList(t *testing.T) {
	createdAt := time.Now()
	repository := &fakeNotificationRepository{notifications: []models.Notification{{ID: "notification-1", CreatedAt: createdAt}}}
	service := NewNotificationService(repository)

	notifications, err := service.List(context.Background(), "recipient-1")
	if err != nil {
		t.Fatal(err)
	}
	if !repository.listCalled || repository.recipientID != "recipient-1" {
		t.Fatalf("unexpected repository call: %+v", repository)
	}
	if len(notifications) != 1 || notifications[0].ID != "notification-1" {
		t.Fatalf("unexpected notifications: %+v", notifications)
	}
}

func TestNotificationServiceReturnsRepositoryError(t *testing.T) {
	repository := &fakeNotificationRepository{err: errors.New("database down")}
	service := NewNotificationService(repository)

	if err := service.CreateFollow(context.Background(), "recipient-1", "actor-1"); err == nil {
		t.Fatal("expected error")
	}
}
