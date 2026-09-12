package main

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

func TestConnectionURLEscapesPassword(t *testing.T) {
	password := "a:@/?#'\\secret"
	u, err := url.Parse(dbadmin.ConnectionURL("db.example", 5432, "app_user", password, "/tmp/ca.pem"))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := u.User.Password()
	if got != password || u.Query().Get("sslmode") != "verify-full" || u.Query().Get("sslrootcert") != "/tmp/ca.pem" {
		t.Fatal("invalid connection URL")
	}
}

func TestApplicationPasswordReusesExistingValue(t *testing.T) {
	existing := `{"username":"app_user","password":"saved-password"}`
	got, err := databaseUserPassword(existing, "app_user")
	if err != nil || got != "saved-password" {
		t.Fatal("existing password was not preserved")
	}
}

func TestApplicationPasswordRejectsInvalidSecretFormat(t *testing.T) {
	existing := "postgres://app_user:secret@db.example:5432/app"
	_, err := databaseUserPassword(existing, "app_user")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("expected safe rejection")
	}
}

func TestApplicationPasswordGeneratesRandomValue(t *testing.T) {
	a, err := databaseUserPassword("", "app_user")
	if err != nil {
		t.Fatal(err)
	}
	b, err := databaseUserPassword("", "app_user")
	if err != nil || len(a) != 64 || a == b {
		t.Fatal("invalid random password")
	}
}

func newTransaction(t *testing.T) (*sql.Tx, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx, mock
}

func TestConfigureAppRoleCreatesUserAndGrantsOnlyApplicationTables(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE app_user LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE ON SCHEMA public TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.users").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(true))
	mock.ExpectExec("GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.users TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.posts").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(true))
	mock.ExpectExec("GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.posts TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.follows").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(true))
	mock.ExpectExec("GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.follows TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := configureAppRole(context.Background(), tx, "test-only-password", false); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureAppRoleRefusesUnknownExistingUser(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if err := configureAppRole(context.Background(), tx, "test-only-password", false); err == nil {
		t.Fatal("expected rejection")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureAppRoleReusesUserBeforeMigrations(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.users").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(false))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.posts").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(false))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.follows").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(false))
	if err := configureAppRole(context.Background(), tx, "unchanged", true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureAppRoleDoesNotExposeDatabaseError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE").WillReturnError(errors.New("password=private-value"))
	err := configureAppRole(context.Background(), tx, "private-value", false)
	if err == nil || strings.Contains(err.Error(), "private-value") {
		t.Fatal("expected sanitized error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
