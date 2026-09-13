package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNotificationRepositoryCreateFollow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO notifications (recipient_id, actor_id, type)
		VALUES ($1, $2, 'follow')
		ON CONFLICT DO NOTHING`)).
		WithArgs("recipient-1", "actor-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewNotificationRepository(db).CreateFollow(context.Background(), "recipient-1", "actor-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationRepositoryCreateFollowSkipsSelfNotification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := NewNotificationRepository(db).CreateFollow(context.Background(), "user-1", "user-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationRepositoryCreateFollowReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO notifications (recipient_id, actor_id, type)
		VALUES ($1, $2, 'follow')
		ON CONFLICT DO NOTHING`)).
		WithArgs("recipient-1", "actor-1").
		WillReturnError(errors.New("insert failed"))

	if err := NewNotificationRepository(db).CreateFollow(context.Background(), "recipient-1", "actor-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationRepositoryCreateLike(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO notifications (recipient_id, actor_id, type, post_id)
		VALUES ($1, $2, 'like', $3)
		ON CONFLICT DO NOTHING`)).
		WithArgs("recipient-1", "actor-1", "post-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := NewNotificationRepository(db).CreateLike(context.Background(), "recipient-1", "actor-1", "post-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationRepositoryCreateLikeSkipsSelfNotification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := NewNotificationRepository(db).CreateLike(context.Background(), "user-1", "user-1", "post-1"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationRepositoryCreateLikeReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO notifications (recipient_id, actor_id, type, post_id)
		VALUES ($1, $2, 'like', $3)
		ON CONFLICT DO NOTHING`)).
		WithArgs("recipient-1", "actor-1", "post-1").
		WillReturnError(errors.New("insert failed"))

	if err := NewNotificationRepository(db).CreateLike(context.Background(), "recipient-1", "actor-1", "post-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationRepositoryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	rows := sqlmock.NewRows([]string{"id", "type", "actor_id", "handle", "display_name", "post_id", "created_at"}).
		AddRow("notification-1", "like", "actor-1", "actor", "Actor", "post-1", createdAt).
		AddRow("notification-2", "follow", "actor-2", "follower", "Follower", nil, createdAt)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT notifications.id, notifications.type, users.id, users.handle, users.display_name, notifications.post_id, notifications.created_at")).
		WithArgs("recipient-1").
		WillReturnRows(rows)

	notifications, err := NewNotificationRepository(db).List(context.Background(), "recipient-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(notifications) != 2 || notifications[0].PostID == nil || *notifications[0].PostID != "post-1" || notifications[1].PostID != nil {
		t.Fatalf("unexpected notifications: %+v", notifications)
	}
}

func TestNotificationRepositoryListReturnsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT notifications.id, notifications.type, users.id, users.handle, users.display_name, notifications.post_id, notifications.created_at")).
		WithArgs("recipient-1").
		WillReturnError(errors.New("query failed"))

	if _, err := NewNotificationRepository(db).List(context.Background(), "recipient-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationRepositoryListReturnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "type", "actor_id", "handle", "display_name", "post_id", "created_at"}).
		AddRow("notification-1", "like", "actor-1", "actor", "Actor", "post-1", "not-time")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT notifications.id, notifications.type, users.id, users.handle, users.display_name, notifications.post_id, notifications.created_at")).
		WithArgs("recipient-1").
		WillReturnRows(rows)

	if _, err := NewNotificationRepository(db).List(context.Background(), "recipient-1"); err == nil {
		t.Fatal("expected error")
	}
}
