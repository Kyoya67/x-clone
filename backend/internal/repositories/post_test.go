package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

func TestPostRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	expectation := mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO posts (author_id, content)"))
	expectation.WithArgs("author-1", "hello").WillReturnRows(
		sqlmock.NewRows([]string{"id", "author_id", "content", "created_at"}).AddRow("post-1", "author-1", "hello", createdAt),
	)

	repository := NewPostRepository(db)
	post, err := repository.Create(context.Background(), "author-1", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if post.ID != "post-1" || post.Content != "hello" || !post.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected post: %+v", post)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostRepositoryCreateReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expectation := mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO posts (author_id, content)"))
	expectation.WithArgs("author-1", "hello").WillReturnError(errors.New("database unavailable"))

	repository := NewPostRepository(db)
	_, err = repository.Create(context.Background(), "author-1", "hello")

	var appErr *apperrors.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected application error, got %v", err)
	}
	if appErr.ErrCode != string(apperrors.DependencyUnavailable) {
		t.Fatalf("unexpected error code: %s", appErr.ErrCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
