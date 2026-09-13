package repositories

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLikeRepositoryLike(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`WITH target_post AS (
			SELECT author_id FROM posts WHERE id = $1
		), inserted_like AS (
			INSERT INTO post_likes (post_id, user_id)
			SELECT $1, $2 FROM target_post
			ON CONFLICT DO NOTHING
		)
		SELECT author_id FROM target_post`)).
		WithArgs("post-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"author_id"}).AddRow("author-1"))

	authorID, err := NewLikeRepository(db).Like(context.Background(), "user-1", "post-1")
	if err != nil {
		t.Fatal(err)
	}
	if authorID != "author-1" {
		t.Fatalf("unexpected author ID: %s", authorID)
	}
}

func TestLikeRepositoryLikeReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`WITH target_post AS (
			SELECT author_id FROM posts WHERE id = $1
		), inserted_like AS (
			INSERT INTO post_likes (post_id, user_id)
			SELECT $1, $2 FROM target_post
			ON CONFLICT DO NOTHING
		)
		SELECT author_id FROM target_post`)).
		WithArgs("post-1", "user-1").
		WillReturnError(sql.ErrNoRows)

	err = nil
	_, err = NewLikeRepository(db).Like(context.Background(), "user-1", "post-1")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestLikeRepositoryUnlike(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2`)).
		WithArgs("post-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewLikeRepository(db).Unlike(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatal(err)
	}
}

func TestLikeRepositoryUnlikeReturnsNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2`)).
		WithArgs("post-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = NewLikeRepository(db).Unlike(context.Background(), "user-1", "post-1")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestLikeRepositoryUnlikeReturnsExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2`)).
		WithArgs("post-1", "user-1").
		WillReturnError(errors.New("delete failed"))

	err = NewLikeRepository(db).Unlike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLikeRepositoryUnlikeReturnsRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2`)).
		WithArgs("post-1", "user-1").
		WillReturnResult(sqlmock.NewErrorResult(errors.New("rows affected failed")))

	err = NewLikeRepository(db).Unlike(context.Background(), "user-1", "post-1")
	if err == nil {
		t.Fatal("expected error")
	}
}
