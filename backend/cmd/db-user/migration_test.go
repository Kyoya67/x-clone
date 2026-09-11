package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

func TestConfigureMigrationRoleCreatesUserBeforeMigrations(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE migration_user LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA public TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMigrationOwnershipSetup(mock)

	if err := configureMigrationRole(context.Background(), tx, "test-password", false); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRolePreservesExistingUser(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA public TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMigrationOwnershipSetup(mock)

	if err := configureMigrationRole(context.Background(), tx, "saved-password", true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsUserWithoutSecret(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if err := configureMigrationRole(context.Background(), tx, "password", false); err == nil {
		t.Fatal("expected rejection")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationPasswordReusesSavedValue(t *testing.T) {
	existing := dbadmin.ConnectionURL("db.example", 5432, "migration_user", "saved-password", "/app/certs/rds-ca-bundle.pem")
	password, err := rolePassword(existing, "db.example", 5432, "migration_user")
	if err != nil || password != "saved-password" {
		t.Fatal("expected saved migration password")
	}
}

func TestMigrationPasswordRejectsApplicationSecret(t *testing.T) {
	existing := dbadmin.ConnectionURL("db.example", 5432, "app_user", "private-password", "/app/certs/rds-ca-bundle.pem")
	_, err := rolePassword(existing, "db.example", 5432, "migration_user")
	if err == nil || strings.Contains(err.Error(), "private-password") {
		t.Fatal("expected safe rejection of application secret")
	}
}

func TestConfigureMigrationRoleHidesDatabaseErrors(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE").WillReturnError(errors.New("private-password"))
	err := configureMigrationRole(context.Background(), tx, "private-password", false)
	if err == nil || strings.Contains(err.Error(), "private-password") {
		t.Fatal("expected sanitized error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsSharedSecretBeforeConnecting(t *testing.T) {
	err := run(context.Background(), "app-db", "same-secret", "same-secret", "", "", "db/dbadmin")
	if err == nil || err.Error() != "application and migration secrets must be different" {
		t.Fatal("expected separate secret validation")
	}
}
