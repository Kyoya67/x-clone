package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type fakeAuthUserService struct{}

func (fakeAuthUserService) FindOrCreateByOIDC(context.Context, string, string, string) (models.User, error) {
	return models.User{}, nil
}

func (fakeAuthUserService) FindByID(context.Context, string) (models.User, error) {
	return models.User{}, nil
}

func TestAuthLoginDisablesCache(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		AuthorizeURL: "https://example.com/oauth2/authorize",
		ClientID:     "client-id",
		RedirectURL:  "https://example.com/auth/callback",
	}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/login", nil)

	controller.Login(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache control, got %q", response.Header().Get("Cache-Control"))
	}
	if !strings.HasPrefix(response.Header().Get("Location"), "https://example.com/oauth2/authorize?") {
		t.Fatalf("unexpected redirect location: %s", response.Header().Get("Location"))
	}
}

func TestAuthCallbackRejectsInvalidStateWithoutLeakingDetails(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=query-state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: "x_clone_oauth_state", Value: "cookie-state"})

	controller.Callback(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache control, got %q", response.Header().Get("Cache-Control"))
	}
	if !strings.Contains(response.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("expected bad parameter error body, got %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "state_match") {
		t.Fatalf("internal callback details leaked: %s", response.Body.String())
	}
}
