package repositories

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

func TestUserRepositoryFindOrCreateByOIDC(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (id, oidc_subject, email, handle, display_name, bio, profile_completed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, '', FALSE)
		ON CONFLICT (oidc_subject) DO UPDATE
		SET email = EXCLUDED.email,
		    display_name = CASE
		        WHEN users.profile_completed THEN users.display_name
		        ELSE EXCLUDED.display_name
		    END
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("subject-1234567890", "kyoya@example.com", "user_subject-", "京谷").
		WillReturnRows(sqlmock.NewRows([]string{"id", "handle", "display_name", "bio", "created_at", "needs_profile_setup"}).
			AddRow("user-1", "user_subject-", "京谷", "", createdAt, true))

	user, err := NewUserRepository(db).FindOrCreateByOIDC(context.Background(), "subject-1234567890", "kyoya@example.com", "京谷")
	if err != nil {
		t.Fatal(err)
	}
	if user.Handle != "user_subject-" || !user.NeedsProfileSetup {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryFindOrCreateByOIDCUsesHandleAsDisplayNameFallback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (id, oidc_subject, email, handle, display_name, bio, profile_completed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, '', FALSE)
		ON CONFLICT (oidc_subject) DO UPDATE
		SET email = EXCLUDED.email,
		    display_name = CASE
		        WHEN users.profile_completed THEN users.display_name
		        ELSE EXCLUDED.display_name
		    END
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("short", "not-valid-email-prefix@example.com", "user_short", "user_short").
		WillReturnRows(sqlmock.NewRows([]string{"id", "handle", "display_name", "bio", "created_at", "needs_profile_setup"}).
			AddRow("user-1", "user_short", "user_short", "", createdAt, true))

	user, err := NewUserRepository(db).FindOrCreateByOIDC(context.Background(), "short", "not-valid-email-prefix@example.com", "")
	if err != nil {
		t.Fatal(err)
	}
	if user.DisplayName != "user_short" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryFindOrCreateByOIDCReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (id, oidc_subject, email, handle, display_name, bio, profile_completed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, '', FALSE)
		ON CONFLICT (oidc_subject) DO UPDATE
		SET email = EXCLUDED.email,
		    display_name = CASE
		        WHEN users.profile_completed THEN users.display_name
		        ELSE EXCLUDED.display_name
		    END
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("subject", "user@example.com", "user_subject", "User").
		WillReturnError(errors.New("query failed"))

	if _, err := NewUserRepository(db).FindOrCreateByOIDC(context.Background(), "subject", "user@example.com", "User"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUserRepositoryFindOrCreateByOIDCPreservesCompletedProfileFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (id, oidc_subject, email, handle, display_name, bio, profile_completed)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, '', FALSE)
		ON CONFLICT (oidc_subject) DO UPDATE
		SET email = EXCLUDED.email,
		    display_name = CASE
		        WHEN users.profile_completed THEN users.display_name
		        ELSE EXCLUDED.display_name
		    END
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("subject", "user@example.com", "user_subject", "Google Name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "handle", "display_name", "bio", "created_at", "needs_profile_setup"}).
			AddRow("user-1", "ore_handle", "オレは何？", "bio", createdAt, false))

	user, err := NewUserRepository(db).FindOrCreateByOIDC(context.Background(), "subject", "user@example.com", "Google Name")
	if err != nil {
		t.Fatal(err)
	}
	if user.DisplayName != "オレは何？" || user.Handle != "ore_handle" || user.NeedsProfileSetup {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, handle, display_name, bio, created_at, NOT profile_completed
		FROM users
		WHERE id = $1`)).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "handle", "display_name", "bio", "created_at", "needs_profile_setup"}).
			AddRow("user-1", "kyoya_dev", "京谷", "bio", createdAt, false))

	user, err := NewUserRepository(db).FindByID(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if user.Handle != "kyoya_dev" || user.NeedsProfileSetup {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryFindByIDReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, handle, display_name, bio, created_at, NOT profile_completed
		FROM users
		WHERE id = $1`)).
		WithArgs("user-1").
		WillReturnError(errors.New("query failed"))

	if _, err := NewUserRepository(db).FindByID(context.Background(), "user-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUserRepositoryUpdateProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	createdAt := time.Now()
	request := models.UpdateUserProfileRequest{Handle: "kyoya_dev", DisplayName: "京谷", Bio: "bio"}
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE users
		SET handle = $2,
		    display_name = $3,
		    bio = $4,
		    profile_completed = TRUE
		WHERE id = $1
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("user-1", "kyoya_dev", "京谷", "bio").
		WillReturnRows(sqlmock.NewRows([]string{"id", "handle", "display_name", "bio", "created_at", "needs_profile_setup"}).
			AddRow("user-1", "kyoya_dev", "京谷", "bio", createdAt, false))

	user, err := NewUserRepository(db).UpdateProfile(context.Background(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if user.Handle != "kyoya_dev" || user.NeedsProfileSetup {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryUpdateProfileReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	request := models.UpdateUserProfileRequest{Handle: "kyoya_dev", DisplayName: "京谷", Bio: "bio"}
	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE users
		SET handle = $2,
		    display_name = $3,
		    bio = $4,
		    profile_completed = TRUE
		WHERE id = $1
		RETURNING id, handle, display_name, bio, created_at, NOT profile_completed`)).
		WithArgs("user-1", "kyoya_dev", "京谷", "bio").
		WillReturnError(errors.New("query failed"))

	if _, err := NewUserRepository(db).UpdateProfile(context.Background(), "user-1", request); err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleFromOIDCUsesCleanEmailPrefix(t *testing.T) {
	handle := handleFromOIDC("subject-1234567890", "Kyoya.Dev+oauth@example.com")

	if handle != "user_subject-" {
		t.Fatalf("unexpected handle: %s", handle)
	}
}

func TestHandleFromOIDCHandlesShortSubject(t *testing.T) {
	handle := handleFromOIDC("abc", "!!!@example.com")

	if handle != "user_abc" {
		t.Fatalf("unexpected handle: %s", handle)
	}
}
