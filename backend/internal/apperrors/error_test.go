package apperrors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorWrap(t *testing.T) {
	cause := errors.New("database unavailable")
	err := BadParam.Wrap(cause, "invalid post")

	var appErr *Error
	if !errors.As(err, &appErr) || appErr.ErrCode != string(BadParam) {
		t.Fatalf("unexpected application error: %#v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("original error was not preserved")
	}
}

func TestErrorHandlerDoesNotExposeCause(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/posts", nil)

	ErrorHandler(recorder, request, errors.New("secret database details"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	var response Error
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Err != nil || response.ErrCode != string(Unknown) {
		t.Fatalf("unexpected response: %+v", response)
	}
}
