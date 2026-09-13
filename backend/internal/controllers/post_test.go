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
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

const testUserID = "00000000-0000-0000-0000-000000000001"

type fakePostService struct {
	post models.Post
	err  error
}

func (f *fakePostService) Create(_ context.Context, _ string, _ models.CreatePostRequest) (models.Post, error) {
	return f.post, f.err
}

func TestPostControllerCreate(t *testing.T) {
	service := &fakePostService{post: models.Post{ID: "post-1", AuthorID: testUserID, Content: "hello"}}
	controller := NewPostController(service)
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"content":"hello"}`)))

	controller.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
	var response models.Post
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != "post-1" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func authenticatedRequest(request *http.Request) *http.Request {
	return request.WithContext(auth.WithUserID(request.Context(), testUserID))
}

func TestPostControllerReturnsGenericError(t *testing.T) {
	controller := NewPostController(&fakePostService{err: errors.New("secret database details")})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"content":"hello"}`)))

	controller.Create(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "secret database details") {
		t.Fatal("internal error was exposed")
	}
}

func TestPostControllerRejectsUnknownField(t *testing.T) {
	controller := NewPostController(&fakePostService{})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"cotent":"hello"}`)))

	controller.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "cotent") {
		t.Fatal("unknown field name was exposed")
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R006"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestPostControllerRejectsInvalidJSON(t *testing.T) {
	controller := NewPostController(&fakePostService{})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"content":}`)))

	controller.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R002"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestPostControllerReturnsServiceValidationError(t *testing.T) {
	controller := NewPostController(&fakePostService{
		err: apperrors.BadParam.Wrap(nil, "content must be between 1 and 280 characters"),
	})
	recorder := httptest.NewRecorder()
	request := authenticatedRequest(httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(`{"content":""}`)))

	controller.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "content must be between 1 and 280 characters") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}
