package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

func TestTimelineRepositoryListForYou(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT posts.id, posts.content, posts.created_at, users.id, users.handle, users.display_name")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "created_at", "author_id", "handle", "display_name"}).
			AddRow("post-1", "hello", createdAt, "author-1", "author_handle", "Author"))

	repository := NewTimelineRepository(db)
	posts, err := repository.List(context.Background(), "user-1", models.TimelineFeedForYou)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].ID != "post-1" || posts[0].Author.Handle != "author_handle" {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineRepositoryListFollowing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("INNER JOIN follows ON follows.followee_id = posts.author_id")).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "created_at", "author_id", "handle", "display_name"}))

	repository := NewTimelineRepository(db)
	posts, err := repository.List(context.Background(), "user-1", models.TimelineFeedFollowing)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected no posts, got %+v", posts)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineRepositoryListReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT posts.id, posts.content, posts.created_at, users.id, users.handle, users.display_name")).
		WillReturnError(errors.New("database unavailable"))

	repository := NewTimelineRepository(db)
	_, err = repository.List(context.Background(), "user-1", models.TimelineFeedForYou)
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected application error, got %v", err)
	}
	if appErr.ErrCode != string(apperrors.DependencyUnavailable) {
		t.Fatalf("unexpected error code: %s", appErr.ErrCode)
	}
}

func TestTimelineRepositoryListReturnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT posts.id, posts.content, posts.created_at, users.id, users.handle, users.display_name")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "content", "created_at", "author_id", "handle", "display_name"}).
			AddRow(nil, "hello", time.Now(), "author-1", "author_handle", "Author"))

	repository := NewTimelineRepository(db)
	_, err = repository.List(context.Background(), "user-1", models.TimelineFeedForYou)
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected application error, got %v", err)
	}
	if appErr.ErrCode != string(apperrors.DependencyUnavailable) {
		t.Fatalf("unexpected error code: %s", appErr.ErrCode)
	}
}

func TestTimelineRepositoryListReturnsRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "content", "created_at", "author_id", "handle", "display_name"}).
		AddRow("post-1", "hello", time.Now(), "author-1", "author_handle", "Author").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT posts.id, posts.content, posts.created_at, users.id, users.handle, users.display_name")).
		WillReturnRows(rows)

	repository := NewTimelineRepository(db)
	_, err = repository.List(context.Background(), "user-1", models.TimelineFeedForYou)
	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected application error, got %v", err)
	}
	if appErr.ErrCode != string(apperrors.DependencyUnavailable) {
		t.Fatalf("unexpected error code: %s", appErr.ErrCode)
	}
}
