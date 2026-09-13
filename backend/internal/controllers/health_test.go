package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthControllerHealth(t *testing.T) {
	controller := NewHealthController()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	controller.Health(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("unexpected content type: %s", recorder.Header().Get("Content-Type"))
	}
	if recorder.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}
