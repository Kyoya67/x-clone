package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeNotificationService struct {
	called        bool
	recipientID   string
	notifications []models.Notification
	err           error
}

func (f *fakeNotificationService) List(_ context.Context, recipientID string) ([]models.Notification, error) {
	f.called = true
	f.recipientID = recipientID
	return f.notifications, f.err
}

func TestNotificationControllerList(t *testing.T) {
	service := &fakeNotificationService{
		notifications: []models.Notification{{
			ID:        "notification-1",
			Type:      models.NotificationTypeFollow,
			Actor:     models.NotificationActor{ID: "actor-1", Handle: "actor", DisplayName: "Actor"},
			CreatedAt: time.Now(),
		}},
	}
	controller := NewNotificationController(service)

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/notifications", nil))

	controller.List(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !service.called || service.recipientID != testUserID {
		t.Fatalf("unexpected service call: %+v", service)
	}
	if !strings.Contains(recorder.Body.String(), `"notifications"`) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestNotificationControllerListRequiresLogin(t *testing.T) {
	controller := NewNotificationController(&fakeNotificationService{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/notifications", nil)

	controller.List(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestNotificationControllerListReturnsServiceError(t *testing.T) {
	controller := NewNotificationController(&fakeNotificationService{err: errors.New("database secret")})

	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodGet, "/notifications", nil))

	controller.List(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "database secret") {
		t.Fatalf("internal details leaked: %s", recorder.Body.String())
	}
}
