package main

import (
	"net/url"
	"strings"
	"testing"
)

func clearDatabaseEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DATABASE_SSL_MODE", "")
}

func TestDatabaseConnectionURLUsesSecretKeys(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "app_user")
	t.Setenv("DB_PASSWORD", "test-only:@/?#'\\password")
	value, err := databaseConnectionURL()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(value)
	if err != nil {
		t.Fatal("invalid URL")
	}
	password, _ := u.User.Password()
	if u.Host != "db.example:5432" || u.Path != "/app" || u.User.Username() != "app_user" || password != "test-only:@/?#'\\password" {
		t.Fatal("incorrect connection parameters")
	}
	if u.Query().Get("sslmode") != "verify-full" || u.Query().Get("sslrootcert") != "/app/certs/rds-ca-bundle.pem" {
		t.Fatal("TLS verification must be enabled")
	}
}

func TestDatabaseConnectionURLRejectsIncompleteSecretSettings(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DB_PASSWORD", "test-only-private-value")
	_, err := databaseConnectionURL()
	if err == nil || strings.Contains(err.Error(), "test-only-private-value") {
		t.Fatal("expected safe failure without local fallback")
	}
}

func TestDatabaseConnectionURLPreservesLocalConfiguration(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DATABASE_URL", "postgres://app:app@localhost:5432/app")
	t.Setenv("DATABASE_SSL_MODE", "disable")
	value, err := databaseConnectionURL()
	if err != nil || value != "postgres://app:app@localhost:5432/app?sslmode=disable" {
		t.Fatal("local database configuration changed")
	}
}

func TestDatabaseConnectionURLRejectsInvalidPort(t *testing.T) {
	clearDatabaseEnvironment(t)
	t.Setenv("DB_HOST", "db.example")
	t.Setenv("DB_PORT", "65536")
	t.Setenv("DB_USER", "app_user")
	t.Setenv("DB_PASSWORD", "test-only")
	if _, err := databaseConnectionURL(); err == nil {
		t.Fatal("expected invalid port error")
	}
}
