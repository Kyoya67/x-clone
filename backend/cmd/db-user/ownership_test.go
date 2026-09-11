package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func expectMigrationOwnershipSetup(mock sqlmock.Sqlmock) {
	mock.ExpectExec("GRANT migration_user TO dbadmin WITH INHERIT TRUE, SET TRUE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS pgcrypto").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMissingTable(mock, "users")
	expectMissingTable(mock, "posts")
	expectMissingTable(mock, "follows")
	expectMissingTable(mock, "schema_migrations")
}

func expectMissingTable(mock sqlmock.Sqlmock, table string) {
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public." + table).
		WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(false))
}

func expectTableOwnership(mock sqlmock.Sqlmock, table string) {
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public." + table).
		WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(true))
	mock.ExpectExec("ALTER TABLE public." + table + " OWNER TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
}

func TestTransferMigrationTablesIncludesApplicationAndVersionTables(t *testing.T) {
	tx, mock := newTransaction(t)
	expectTableOwnership(mock, "users")
	expectTableOwnership(mock, "posts")
	expectTableOwnership(mock, "follows")
	expectTableOwnership(mock, "schema_migrations")
	if err := transferMigrationTables(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTransferMigrationTablesSkipsAbsentTables(t *testing.T) {
	tx, mock := newTransaction(t)
	expectMissingTable(mock, "users")
	expectMissingTable(mock, "posts")
	expectMissingTable(mock, "follows")
	expectMissingTable(mock, "schema_migrations")
	if err := transferMigrationTables(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTransferMigrationTablesHidesDatabaseErrors(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.users").WillReturnRows(sqlmock.NewRows([]string{"present"}).AddRow(true))
	mock.ExpectExec("ALTER TABLE").WillReturnError(errors.New("private database details"))
	err := transferMigrationTables(context.Background(), tx)
	if err == nil || strings.Contains(err.Error(), "private database details") {
		t.Fatal("expected sanitized error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
