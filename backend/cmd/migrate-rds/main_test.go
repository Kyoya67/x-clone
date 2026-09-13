package main

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func TestDatabaseURLUsesInjectedCredentialsAndVerifiesRDS(t *testing.T) {
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "migration_user")
	t.Setenv("DB_PASSWORD", "test:@/?password")
	value, err := databaseURL("/app/certs/rds-ca-bundle.pem")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if u.Host != "db.example:5432" || u.User.Username() != "migration_user" || password != "test:@/?password" || u.Query().Get("sslmode") != "verify-full" {
		t.Fatal("unexpected database configuration")
	}
}

func TestDatabaseURLRejectsMissingPassword(t *testing.T) {
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "migration_user")
	t.Setenv("DB_PASSWORD", "")
	if _, err := databaseURL("/app/certs/rds-ca-bundle.pem"); err == nil {
		t.Fatal("missing password accepted")
	}
}

func TestDatabaseURLRejectsAdministrator(t *testing.T) {
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "dbadmin")
	t.Setenv("DB_PASSWORD", "test-only-password")
	if _, err := databaseURL("/app/certs/rds-ca-bundle.pem"); err == nil {
		t.Fatal("administrator credentials accepted")
	}
}

func TestDatabaseURLRejectsApplicationUser(t *testing.T) {
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "app_user")
	t.Setenv("DB_PASSWORD", "test-only-password")
	if _, err := databaseURL("/app/certs/rds-ca-bundle.pem"); err == nil {
		t.Fatal("application credentials accepted")
	}
}

func TestApplySucceeds(t *testing.T) {
	called := false
	err := apply(func() error { called = true; return nil })
	if err != nil || !called {
		t.Fatal("migration was not executed successfully")
	}
}

func TestApplyAcceptsNoChange(t *testing.T) {
	if err := apply(func() error { return migrate.ErrNoChange }); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRejectsDirtyDatabase(t *testing.T) {
	err := apply(func() error { return migrate.ErrDirty{Version: 4} })
	if err == nil || !strings.Contains(err.Error(), "version 4 is dirty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyHidesDriverDetails(t *testing.T) {
	err := apply(func() error { return errors.New("secret password and row data") })
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("raw driver error exposed")
	}
}

func TestMigrationFilesHaveOrderedVersions(t *testing.T) {
	source, err := iofs.New(os.DirFS("../../migrations"), ".")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	version, err := source.First()
	if err != nil || version != 1 {
		t.Fatalf("unexpected first version: %d, %v", version, err)
	}
	lastVersion := uint(7)
	for expected := uint(1); expected <= lastVersion; expected++ {
		if version != expected {
			t.Fatalf("unexpected version: %d", version)
		}
		reader, _, err := source.ReadUp(version)
		if err != nil {
			t.Fatal(err)
		}
		reader.Close()
		version, err = source.Next(version)
		if expected < lastVersion && err != nil {
			t.Fatal(err)
		}
		if expected == lastVersion && !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected extra migration: %v", err)
		}
	}
}
