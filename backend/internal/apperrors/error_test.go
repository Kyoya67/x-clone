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

func TestErrorReturnsMessage(t *testing.T) {
	err := &Error{Message: "invalid post"}
	if err.Error() != "invalid post" {
		t.Fatalf("unexpected error message: %s", err.Error())
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

func TestStatusCodeFor(t *testing.T) {
	tests := []struct {
		name string
		code ErrCode
		want int
	}{
		{name: "bad parameter", code: BadParam, want: http.StatusBadRequest},
		{name: "request body decode failed", code: ReqBodyDecodeFailed, want: http.StatusBadRequest},
		{name: "request body too large", code: RequestBodyTooLarge, want: http.StatusBadRequest},
		{name: "unknown field", code: UnknownField, want: http.StatusBadRequest},
		{name: "not found", code: NotFound, want: http.StatusNotFound},
		{name: "dependency unavailable", code: DependencyUnavailable, want: http.StatusServiceUnavailable},
		{name: "unknown", code: Unknown, want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusCodeFor(tt.code); got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}
