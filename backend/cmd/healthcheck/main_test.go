package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" || r.Method != http.MethodGet {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := check(server.Client(), server.URL+"/health"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsUnhealthyStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	if err := check(server.Client(), server.URL); err == nil {
		t.Fatal("expected an error for status 503")
	}
}

func TestCheckReturnsConnectionError(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	if err := check(&http.Client{Timeout: time.Second}, server.URL); err == nil {
		t.Fatal("expected a connection error")
	}
}

func TestCheckTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()
	if err := check(&http.Client{Timeout: 50 * time.Millisecond}, server.URL); err == nil {
		t.Fatal("expected a timeout")
	}
}
