package dbadmin

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestConnectionURLBuildsPostgresURL(t *testing.T) {
	databaseURL := ConnectionURL("db.example", 5432, "app_user", "p@ss/word", "/tmp/rds-ca.pem")

	u, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	password, ok := u.User.Password()
	if u.Scheme != "postgres" ||
		u.Host != "db.example:5432" ||
		u.Path != "/app" ||
		u.User.Username() != "app_user" ||
		!ok ||
		password != "p@ss/word" ||
		u.Query().Get("sslmode") != "verify-full" ||
		u.Query().Get("sslrootcert") != "/tmp/rds-ca.pem" ||
		u.Query().Get("connect_timeout") != "10" {
		t.Fatalf("unexpected database url: %s", databaseURL)
	}
}

func TestOpenAdministratorUsesExplicitSecretWithoutRDSManagedSecret(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	secretID := "db/dbadmin"
	restoreAWS := stubAWSForOpenAdministrator(
		t,
		RDSEndpoint{Host: "db.example", Port: 5432},
		secretID,
		`{"username":"wrong-user","password":"test-only-password"}`,
	)
	defer restoreAWS()
	db, rdsEndpoint, err := OpenAdministrator(context.Background(), "app-db", secretID, caFile, "")
	if db != nil || err == nil || err.Error() != "invalid administrator secret" {
		t.Fatalf("expected validation of the explicit secret, got %v", err)
	}
	if rdsEndpoint.Host != "db.example" || rdsEndpoint.Port != 5432 {
		t.Fatal("expected endpoint from RDS")
	}
}

func stubAWSForOpenAdministrator(t *testing.T, rdsEndpoint RDSEndpoint, expectedSecretID, secret string) func() {
	t.Helper()
	originalEndpoint := getRDSEndpoint
	originalSecret := getSecretString
	getRDSEndpoint = func(context.Context, string) (RDSEndpoint, error) {
		return rdsEndpoint, nil
	}
	getSecretString = func(_ context.Context, secretID string) (string, error) {
		if secretID != expectedSecretID {
			t.Fatalf("unexpected secret id: %s", secretID)
		}
		return secret, nil
	}
	return func() {
		getRDSEndpoint = originalEndpoint
		getSecretString = originalSecret
	}
}

func TestOpenAdministratorRejectsMissingSecret(t *testing.T) {
	_, _, err := OpenAdministrator(context.Background(), "app-db", "", "", "")
	if err == nil || err.Error() != "administrator secret is required" {
		t.Fatalf("expected missing secret error, got %v", err)
	}
}

func TestOpenAdministratorRejectsMissingCAFile(t *testing.T) {
	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", "/tmp/missing-rds-ca.pem", "")
	if err == nil || err.Error() != "cannot read RDS CA file" {
		t.Fatalf("expected missing CA file error, got %v", err)
	}
}

func TestOpenAdministratorReturnsRDSEndpointError(t *testing.T) {
	caFile := writeTestCAFile(t)
	originalEndpoint := getRDSEndpoint
	getRDSEndpoint = func(context.Context, string) (RDSEndpoint, error) {
		return RDSEndpoint{}, errors.New("RDS lookup failed")
	}
	defer func() {
		getRDSEndpoint = originalEndpoint
	}()

	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "")
	if err == nil || err.Error() != "RDS lookup failed" {
		t.Fatalf("expected RDS endpoint error, got %v", err)
	}
}

func TestOpenAdministratorRejectsMissingRDSEndpoint(t *testing.T) {
	caFile := writeTestCAFile(t)
	restoreAWS := stubAWSForOpenAdministrator(t, RDSEndpoint{}, "db/dbadmin", `{"username":"dbadmin","password":"test-only-password"}`)
	defer restoreAWS()

	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "")
	if err == nil || err.Error() != "RDS endpoint is missing" {
		t.Fatalf("expected missing endpoint error, got %v", err)
	}
}

func TestOpenAdministratorReturnsSecretError(t *testing.T) {
	caFile := writeTestCAFile(t)
	originalEndpoint := getRDSEndpoint
	originalSecret := getSecretString
	getRDSEndpoint = func(context.Context, string) (RDSEndpoint, error) {
		return RDSEndpoint{Host: "db.example", Port: 5432}, nil
	}
	getSecretString = func(context.Context, string) (string, error) {
		return "", errors.New("secret lookup failed")
	}
	defer func() {
		getRDSEndpoint = originalEndpoint
		getSecretString = originalSecret
	}()

	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "")
	if err == nil || err.Error() != "secret lookup failed" {
		t.Fatalf("expected secret error, got %v", err)
	}
}

func TestOpenAdministratorRejectsInvalidSecretJSON(t *testing.T) {
	caFile := writeTestCAFile(t)
	restoreAWS := stubAWSForOpenAdministrator(t, RDSEndpoint{Host: "db.example", Port: 5432}, "db/dbadmin", `{`)
	defer restoreAWS()

	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "")
	if err == nil || err.Error() != "invalid administrator secret" {
		t.Fatalf("expected invalid secret error, got %v", err)
	}
}

func TestOpenAdministratorRejectsEmptyAdminPassword(t *testing.T) {
	caFile := writeTestCAFile(t)
	restoreAWS := stubAWSForOpenAdministrator(t, RDSEndpoint{Host: "db.example", Port: 5432}, "db/dbadmin", `{"username":"dbadmin","password":""}`)
	defer restoreAWS()

	_, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "")
	if err == nil || err.Error() != "invalid administrator secret" {
		t.Fatalf("expected invalid secret error, got %v", err)
	}
}

func TestOpenAdministratorRejectsInvalidLocalForwardEndpoint(t *testing.T) {
	caFile := writeTestCAFile(t)
	restoreAWS := stubAWSForOpenAdministrator(t, RDSEndpoint{Host: "db.example", Port: 5432}, "db/dbadmin", `{"username":"dbadmin","password":"test-only-password"}`)
	defer restoreAWS()

	db, _, err := OpenAdministrator(context.Background(), "app-db", "db/dbadmin", caFile, "203.0.113.1:15432")
	if db != nil || err == nil || err.Error() != "cannot initialize database connection" {
		t.Fatalf("expected database initialization error, got db=%v err=%v", db, err)
	}
}

func writeTestCAFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	return caFile
}
