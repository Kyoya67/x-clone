package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSwaggerUI(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)

	SwaggerUI(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", recorder.Header().Get("Content-Type"))
	}
	if !strings.Contains(recorder.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestOpenAPISpec(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Mkdir("openapi", 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte("openapi: 3.0.0\n")
	if err := os.WriteFile(filepath.Join("openapi", "openapi.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)

	OpenAPISpec(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("Content-Type") != "application/yaml; charset=utf-8" {
		t.Fatalf("unexpected content type: %s", recorder.Header().Get("Content-Type"))
	}
	if recorder.Body.String() != string(spec) {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestOpenAPISpecReturnsErrorWhenSpecMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)

	OpenAPISpec(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "OpenAPI specification is unavailable") {
		t.Fatalf("unexpected response: %s", recorder.Body.String())
	}
}

func TestReadOpenAPISpecFindsBackendPath(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.MkdirAll(filepath.Join("backend", "openapi"), 0o755); err != nil {
		t.Fatal(err)
	}
	spec := []byte("info:\n  title: API\n")
	if err := os.WriteFile(filepath.Join("backend", "openapi", "openapi.yaml"), spec, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := readOpenAPISpec()
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(spec) {
		t.Fatalf("unexpected spec: %s", string(got))
	}
}
