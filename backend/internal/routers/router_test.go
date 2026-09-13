package routers

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
)

const routerTestUserID = "00000000-0000-0000-0000-000000000001"
const routerTestSessionSecret = "test-session-secret"

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	NewRouter(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if body := recorder.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestSwaggerUI(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)

	NewRouter(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
}

func TestOpenAPISpec(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)

	NewRouter(nil).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/yaml; charset=utf-8" {
		t.Fatalf("unexpected content type: %q", got)
	}
}

func TestFollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING`)).
		WithArgs("00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO notifications (recipient_id, actor_id, type)
		VALUES ($1, $2, 'follow')
		ON CONFLICT DO NOTHING`)).
		WithArgs("00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000001").
		WillReturnResult(sqlmock.NewResult(0, 1))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/users/00000000-0000-0000-0000-000000000002/follow", nil)
	authenticateRouterRequest(t, request)

	NewRouter(db).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnfollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2`)).
		WithArgs("00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002").
		WillReturnResult(sqlmock.NewResult(0, 1))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/users/00000000-0000-0000-0000-000000000002/follow", nil)
	authenticateRouterRequest(t, request)

	NewRouter(db).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimeline(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("INNER JOIN follows ON follows.followee_id = posts.author_id")).
		WithArgs("00000000-0000-0000-0000-000000000001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "created_at", "author_id", "handle", "display_name"}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/timeline?feed=following", nil)
	authenticateRouterRequest(t, request)

	NewRouter(db).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func authenticateRouterRequest(t *testing.T, request *http.Request) {
	t.Helper()
	t.Setenv("AUTH_SESSION_SECRET", routerTestSessionSecret)
	session, err := auth.SignSession(routerTestUserID, routerTestSessionSecret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session})
}
