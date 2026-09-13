package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbaccess"
)

func TestRunCLISucceeds(t *testing.T) {
	original := runCommand
	runCommand = func(_ context.Context, instance, appSecretID, migrationSecretID, caFile, localForwardEndpoint, adminSecretID string) error {
		if instance != "custom-db" ||
			appSecretID != "db/app_user" ||
			migrationSecretID != "db/migration_user" ||
			caFile != "/tmp/ca.pem" ||
			localForwardEndpoint != "127.0.0.1:15432" ||
			adminSecretID != "db/dbadmin" {
			t.Fatalf("unexpected args: instance=%s app=%s migration=%s ca=%s local=%s admin=%s", instance, appSecretID, migrationSecretID, caFile, localForwardEndpoint, adminSecretID)
		}
		return nil
	}
	defer func() {
		runCommand = original
	}()

	var stdout, stderr bytes.Buffer
	code := runCLI([]string{
		"--instance", "custom-db",
		"--ca-file", "/tmp/ca.pem",
		"--local-forward-endpoint", "127.0.0.1:15432",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("unexpected exit code: %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "app_user and migration_user configured") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestRunCLIRejectsMissingCAFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI(nil, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("unexpected exit code: %d", code)
	}
	if stdout.String() != "" || !strings.Contains(stderr.String(), "--ca-file is required") {
		t.Fatalf("unexpected output: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestRunCLIReturnsRunError(t *testing.T) {
	original := runCommand
	runCommand = func(context.Context, string, string, string, string, string, string) error {
		return errors.New("setup failed")
	}
	defer func() {
		runCommand = original
	}()

	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"--ca-file", "/tmp/ca.pem"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("unexpected exit code: %d", code)
	}
	if stdout.String() != "" || !strings.Contains(stderr.String(), "setup failed") {
		t.Fatalf("unexpected output: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestMainCallsExitWithRunCLIResult(t *testing.T) {
	originalArgs := os.Args
	originalRun := runCommand
	originalExit := exit
	originalStdout := standardOutput
	originalStderr := standardError
	defer func() {
		os.Args = originalArgs
		runCommand = originalRun
		exit = originalExit
		standardOutput = originalStdout
		standardError = originalStderr
	}()

	os.Args = []string{"db-user", "--ca-file", "/tmp/ca.pem"}
	runCommand = func(context.Context, string, string, string, string, string, string) error {
		return nil
	}
	var stdout, stderr bytes.Buffer
	standardOutput = &stdout
	standardError = &stderr
	var exitCode int
	exit = func(code int) {
		exitCode = code
	}

	main()
	if exitCode != 0 {
		t.Fatalf("unexpected exit code: %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "app_user and migration_user configured") || stderr.String() != "" {
		t.Fatalf("unexpected output: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestConnectionURLEscapesPassword(t *testing.T) {
	password := "a:@/?#'\\secret"
	u, err := url.Parse(dbaccess.ConnectionURL("db.example", 5432, "app_user", password, "/tmp/ca.pem"))
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

func TestDatabaseUserPasswordReturnsRandomError(t *testing.T) {
	original := randomRead
	randomRead = func([]byte) (int, error) {
		return 0, errors.New("random failed")
	}
	defer func() {
		randomRead = original
	}()

	_, err := databaseUserPassword("", "app_user")
	if err == nil || err.Error() != "cannot generate password" {
		t.Fatalf("unexpected error: %v", err)
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

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func stubOpenAdministrator(t *testing.T, db *sql.DB, rdsEndpoint dbaccess.RDSEndpoint, err error) func() {
	t.Helper()
	original := openAdministrator
	openAdministrator = func(context.Context, string, string, string, string) (*sql.DB, dbaccess.RDSEndpoint, error) {
		return db, rdsEndpoint, err
	}
	return func() {
		openAdministrator = original
	}
}

func TestRunConfiguresApplicationAndMigrationUsers(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()
	restoreSecrets := stubSecretStore(t, false, "", func(string, string) error { return nil })
	defer restoreSecrets()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE ON SCHEMA public TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA public TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMigrationRolePrivileges(mock)
	mock.ExpectCommit()

	if err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsSetupLockError(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnError(errors.New("lock failed"))
	mock.ExpectRollback()

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "cannot acquire setup lock" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsOpenAdministratorError(t *testing.T) {
	restoreOpen := stubOpenAdministrator(t, nil, dbaccess.RDSEndpoint{}, errors.New("administrator connection failed"))
	defer restoreOpen()

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "administrator connection failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsBeginTransactionError(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()
	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "cannot start database transaction" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsCommitError(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()
	restoreSecrets := stubSecretStore(t, false, "", func(string, string) error { return nil })
	defer restoreSecrets()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE ON SCHEMA public TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA public TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMigrationRolePrivileges(mock)
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "database commit failed; rerun to reconcile with saved secrets" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsApplicationSetupError(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()
	originalCurrent := currentSecretVersionExists
	currentSecretVersionExists = func(context.Context, string) (bool, error) {
		return false, errors.New("app setup failed")
	}
	defer func() {
		currentSecretVersionExists = originalCurrent
	}()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "app setup failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsMigrationSetupError(t *testing.T) {
	db, mock := newMockDB(t)
	restoreOpen := stubOpenAdministrator(t, db, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, nil)
	defer restoreOpen()
	originalCurrent := currentSecretVersionExists
	originalPut := putSecretString
	call := 0
	currentSecretVersionExists = func(context.Context, string) (bool, error) {
		call++
		if call == 2 {
			return false, errors.New("migration setup failed")
		}
		return false, nil
	}
	putSecretString = func(context.Context, string, string) error {
		return nil
	}
	defer func() {
		currentSecretVersionExists = originalCurrent
		putSecretString = originalPut
	}()

	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE ON SCHEMA public TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := run(context.Background(), "app-db", "db/app_user", "db/migration_user", "/tmp/ca.pem", "", "db/dbadmin")
	if err == nil || err.Error() != "migration setup failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureAppRoleCreatesUserAndGrantsBasicPrivileges(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE app_user LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE ON SCHEMA public TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
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

func TestConfigureAppRoleRejectsDatabaseGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnError(errors.New("grant failed"))

	err := configureAppRole(context.Background(), tx, "test-only-password", true)
	if err == nil || err.Error() != "cannot grant database access" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureAppRoleRejectsSchemaGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE").WillReturnError(errors.New("grant failed"))

	err := configureAppRole(context.Background(), tx, "test-only-password", true)
	if err == nil || err.Error() != "cannot grant schema access" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseUserPasswordReadsJSONCredentials(t *testing.T) {
	password, err := databaseUserPassword(`{"username":"migration_user","password":"test-only!#$"}`, "migration_user")
	if err != nil || password != "test-only!#$" {
		t.Fatal("expected saved JSON password")
	}
}

func TestDatabaseUserPasswordRejectsWrongJSONUser(t *testing.T) {
	_, err := databaseUserPassword(`{"username":"dbadmin","password":"test-only-private"}`, "app_user")
	if err == nil || strings.Contains(err.Error(), "test-only-private") {
		t.Fatal("expected safe rejection")
	}
}

func TestDatabaseUserPasswordRejectsMissingJSONPassword(t *testing.T) {
	if _, err := databaseUserPassword(`{"username":"app_user"}`, "app_user"); err == nil {
		t.Fatal("expected missing password error")
	}
}

func TestConfigureMigrationRoleCreatesUserBeforeMigrations(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("CREATE ROLE migration_user LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT CONNECT ON DATABASE app TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA public TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	expectMigrationRolePrivileges(mock)

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
	expectMigrationRolePrivileges(mock)

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
	existing := `{"username":"migration_user","password":"saved-password"}`
	password, err := databaseUserPassword(existing, "migration_user")
	if err != nil || password != "saved-password" {
		t.Fatal("expected saved migration password")
	}
}

func TestMigrationPasswordRejectsApplicationSecret(t *testing.T) {
	existing := `{"username":"app_user","password":"private-password"}`
	_, err := databaseUserPassword(existing, "migration_user")
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

func TestConfigureMigrationRoleRejectsDatabaseGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnError(errors.New("grant failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot grant migration database access" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsSchemaGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnError(errors.New("grant failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot grant migration schema access" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsMembershipGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT migration_user TO dbadmin").WillReturnError(errors.New("grant failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot grant administrator membership in migration role" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsExtensionError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT migration_user TO dbadmin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA IF NOT EXISTS migration").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA migration").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER DEFAULT PRIVILEGES").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE EXTENSION").WillReturnError(errors.New("extension failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot prepare migration extension" {
		t.Fatalf("unexpected error: %v", err)
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

func expectMigrationRolePrivileges(mock sqlmock.Sqlmock) {
	mock.ExpectExec("GRANT migration_user TO dbadmin WITH INHERIT TRUE, SET TRUE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA IF NOT EXISTS migration AUTHORIZATION migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA migration TO migration_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER DEFAULT PRIVILEGES FOR ROLE migration_user IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE EXTENSION IF NOT EXISTS pgcrypto").WillReturnResult(sqlmock.NewResult(0, 0))
}

func TestConfigureMigrationRoleRejectsMigrationSchemaError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT migration_user TO dbadmin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA").WillReturnError(errors.New("schema failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot prepare migration schema" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsMigrationMetadataGrantError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT migration_user TO dbadmin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA migration").WillReturnError(errors.New("grant failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot grant migration metadata schema access" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigureMigrationRoleRejectsDefaultPrivilegeError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("migration_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec("GRANT CONNECT").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT migration_user TO dbadmin").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE SCHEMA").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("GRANT USAGE, CREATE ON SCHEMA migration").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("ALTER DEFAULT PRIVILEGES").WillReturnError(errors.New("default privilege failed"))

	err := configureMigrationRole(context.Background(), tx, "saved-password", true)
	if err == nil || err.Error() != "cannot configure application default table privileges" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRoleRejectsUnsupportedRole(t *testing.T) {
	tx, _ := newTransaction(t)
	err := createRole(context.Background(), tx, "dbadmin", "password", false)
	if err == nil || err.Error() != "unsupported database role" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateRoleReturnsInspectionError(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").WillReturnError(errors.New("inspect failed"))

	err := createRole(context.Background(), tx, "app_user", "password", false)
	if err == nil || err.Error() != "cannot inspect database role" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateRolePreservesExistingRoleWithSecret(t *testing.T) {
	tx, mock := newTransaction(t)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if err := createRole(context.Background(), tx, "app_user", "password", true); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func stubSecretStore(t *testing.T, hasCurrent bool, secret string, put func(string, string) error) func() {
	t.Helper()
	originalCurrent := currentSecretVersionExists
	originalGet := getSecretString
	originalPut := putSecretString
	currentSecretVersionExists = func(context.Context, string) (bool, error) {
		return hasCurrent, nil
	}
	getSecretString = func(context.Context, string) (string, error) {
		return secret, nil
	}
	putSecretString = func(_ context.Context, secretID, secretString string) error {
		if put == nil {
			return nil
		}
		return put(secretID, secretString)
	}
	return func() {
		currentSecretVersionExists = originalCurrent
		getSecretString = originalGet
		putSecretString = originalPut
	}
}

func TestSetupUserCreatesRoleAndStoresNewSecret(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, false, "", func(secretID, secretString string) error {
		if secretID != "db/app_user" {
			t.Fatalf("unexpected secret id: %s", secretID)
		}
		var saved credentials
		if err := json.Unmarshal([]byte(secretString), &saved); err != nil {
			t.Fatal(err)
		}
		if saved.Username != "app_user" || len(saved.Password) != 64 {
			t.Fatalf("unexpected saved credentials: %#v", saved)
		}
		return nil
	})
	defer restore()

	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	configureCalled := false
	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, "db/app_user", "/tmp/ca.pem", "", "app_user", func(_ context.Context, _ *sql.Tx, password string, hasSecret bool) error {
		configureCalled = true
		if len(password) != 64 || hasSecret {
			t.Fatalf("unexpected configure args: password=%s hasSecret=%v", password, hasSecret)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !configureCalled {
		t.Fatal("expected configure callback")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserReusesExistingSecret(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, true, `{"username":"app_user","password":"saved-password"}`, func(string, string) error {
		t.Fatal("did not expect secret update")
		return nil
	})
	defer restore()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	configureCalled := false
	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, "db/app_user", "/tmp/ca.pem", "", "app_user", func(_ context.Context, _ *sql.Tx, password string, hasSecret bool) error {
		configureCalled = true
		if password != "saved-password" || !hasSecret {
			t.Fatalf("unexpected configure args: password=%s hasSecret=%v", password, hasSecret)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !configureCalled {
		t.Fatal("expected configure callback")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserRejectsInvalidConnectionForExistingRole(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, true, `{"username":"app_user","password":"saved-password"}`, nil)
	defer restore()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, "db/app_user", "/tmp/ca.pem", "203.0.113.1:15432", "app_user", nil)
	if err == nil || err.Error() != "cannot initialize database connection" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserVerifiesExistingRoleCredentials(t *testing.T) {
	tx, mock := newTransaction(t)
	restoreSecret := stubSecretStore(t, true, `{"username":"app_user","password":"saved-password"}`, func(string, string) error {
		t.Fatal("did not expect secret update")
		return nil
	})
	defer restoreSecret()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	db, dbMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dbMock.ExpectPing()

	originalConnectionURL := connectionURL
	originalOpenDatabase := openDatabase
	connectionURL = func(host string, port int, user, password, ca string) string {
		if host != "db.example" || port != 5432 || user != "app_user" || password != "saved-password" || ca != "/tmp/ca.pem" {
			t.Fatalf("unexpected connection url args: host=%s port=%d user=%s password=%s ca=%s", host, port, user, password, ca)
		}
		return "postgres://app_user:saved-password@db.example:5432/app"
	}
	openDatabase = func(databaseURL, localForwardEndpoint string) (*sql.DB, error) {
		if databaseURL != "postgres://app_user:saved-password@db.example:5432/app" || localForwardEndpoint != "127.0.0.1:15432" {
			t.Fatalf("unexpected open database args: url=%s local=%s", databaseURL, localForwardEndpoint)
		}
		return db, nil
	}
	defer func() {
		connectionURL = originalConnectionURL
		openDatabase = originalOpenDatabase
	}()

	configureCalled := false
	err = setupUser(context.Background(), tx, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, "db/app_user", "/tmp/ca.pem", "127.0.0.1:15432", "app_user", func(_ context.Context, _ *sql.Tx, password string, hasSecret bool) error {
		configureCalled = true
		if password != "saved-password" || !hasSecret {
			t.Fatalf("unexpected configure args: password=%s hasSecret=%v", password, hasSecret)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !configureCalled {
		t.Fatal("expected configure callback")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserRejectsExistingRolePingError(t *testing.T) {
	tx, mock := newTransaction(t)
	restoreSecret := stubSecretStore(t, true, `{"username":"app_user","password":"saved-password"}`, nil)
	defer restoreSecret()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	db, dbMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dbMock.ExpectPing().WillReturnError(errors.New("ping failed"))

	originalOpenDatabase := openDatabase
	openDatabase = func(string, string) (*sql.DB, error) {
		return db, nil
	}
	defer func() {
		openDatabase = originalOpenDatabase
	}()

	err = setupUser(context.Background(), tx, dbaccess.RDSEndpoint{Host: "db.example", Port: 5432}, "db/app_user", "/tmp/ca.pem", "", "app_user", nil)
	if err == nil || err.Error() != "saved database credentials cannot connect; refusing to change password" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserRejectsEmptyExistingSecret(t *testing.T) {
	tx, _ := newTransaction(t)
	restore := stubSecretStore(t, true, "", nil)
	defer restore()

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", nil)
	if err == nil || err.Error() != "database user secret is empty" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupUserReturnsCurrentSecretError(t *testing.T) {
	tx, _ := newTransaction(t)
	originalCurrent := currentSecretVersionExists
	currentSecretVersionExists = func(context.Context, string) (bool, error) {
		return false, errors.New("secret version lookup failed")
	}
	defer func() {
		currentSecretVersionExists = originalCurrent
	}()

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", nil)
	if err == nil || err.Error() != "secret version lookup failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupUserReturnsGetSecretError(t *testing.T) {
	tx, _ := newTransaction(t)
	originalCurrent := currentSecretVersionExists
	originalGet := getSecretString
	currentSecretVersionExists = func(context.Context, string) (bool, error) {
		return true, nil
	}
	getSecretString = func(context.Context, string) (string, error) {
		return "", errors.New("secret lookup failed")
	}
	defer func() {
		currentSecretVersionExists = originalCurrent
		getSecretString = originalGet
	}()

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", nil)
	if err == nil || err.Error() != "secret lookup failed" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupUserReturnsInvalidCredentialError(t *testing.T) {
	tx, _ := newTransaction(t)
	restore := stubSecretStore(t, true, `{"username":"migration_user","password":"saved-password"}`, nil)
	defer restore()

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", nil)
	if err == nil || err.Error() != "invalid database credentials; refusing to overwrite" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetupUserReturnsRoleInspectionError(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, false, "", nil)
	defer restore()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").WillReturnError(errors.New("inspect failed"))

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", nil)
	if err == nil || err.Error() != "cannot inspect database role" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserReturnsConfigureError(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, false, "", nil)
	defer restore()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", func(context.Context, *sql.Tx, string, bool) error {
		return errors.New("configure failed")
	})
	if err == nil || err.Error() != "configure failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSetupUserReturnsPutSecretError(t *testing.T) {
	tx, mock := newTransaction(t)
	restore := stubSecretStore(t, false, "", func(string, string) error {
		return errors.New("secret save failed")
	})
	defer restore()
	mock.ExpectQuery("SELECT EXISTS").WithArgs("app_user").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	err := setupUser(context.Background(), tx, dbaccess.RDSEndpoint{}, "db/app_user", "", "", "app_user", func(context.Context, *sql.Tx, string, bool) error {
		return nil
	})
	if err == nil || err.Error() != "secret save failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
