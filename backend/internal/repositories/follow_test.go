package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
)

func TestFollowRepositoryFollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING`)).
		WithArgs("follower-1", "followee-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewFollowRepository(db)
	if err := repository.Follow(context.Background(), "follower-1", "followee-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFollowRepositoryFollowReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO follows (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING`)).
		WithArgs("follower-1", "followee-1").
		WillReturnError(errors.New("database unavailable"))

	repository := NewFollowRepository(db)
	err = repository.Follow(context.Background(), "follower-1", "followee-1")
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

func TestFollowRepositoryUnfollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2`)).
		WithArgs("follower-1", "followee-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewFollowRepository(db)
	if err := repository.Unfollow(context.Background(), "follower-1", "followee-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFollowRepositoryUnfollowReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2`)).
		WithArgs("follower-1", "followee-1").
		WillReturnError(errors.New("database unavailable"))

	repository := NewFollowRepository(db)
	err = repository.Unfollow(context.Background(), "follower-1", "followee-1")
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

func TestFollowRepositoryListFolloweeIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT followee_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at ASC`)).
		WithArgs("follower-1").
		WillReturnRows(sqlmock.NewRows([]string{"followee_id"}).AddRow("followee-1").AddRow("followee-2"))

	repository := NewFollowRepository(db)
	followeeIDs, err := repository.ListFolloweeIDs(context.Background(), "follower-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(followeeIDs) != 2 || followeeIDs[0] != "followee-1" || followeeIDs[1] != "followee-2" {
		t.Fatalf("unexpected followee IDs: %v", followeeIDs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFollowRepositoryListFolloweeIDsReturnsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT followee_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at ASC`)).
		WithArgs("follower-1").
		WillReturnError(errors.New("database unavailable"))

	repository := NewFollowRepository(db)
	_, err = repository.ListFolloweeIDs(context.Background(), "follower-1")
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

func TestFollowRepositoryListFolloweeIDsReturnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT followee_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at ASC`)).
		WithArgs("follower-1").
		WillReturnRows(sqlmock.NewRows([]string{"followee_id"}).AddRow(nil))

	repository := NewFollowRepository(db)
	_, err = repository.ListFolloweeIDs(context.Background(), "follower-1")
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

func TestFollowRepositoryListFolloweeIDsReturnsRowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"followee_id"}).
		AddRow("followee-1").
		RowError(0, errors.New("row error"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT followee_id
		FROM follows
		WHERE follower_id = $1
		ORDER BY created_at ASC`)).
		WithArgs("follower-1").
		WillReturnRows(rows)

	repository := NewFollowRepository(db)
	_, err = repository.ListFolloweeIDs(context.Background(), "follower-1")
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
