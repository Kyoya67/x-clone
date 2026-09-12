package dbadmin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAdministratorUsesExplicitSecretWithoutRDSManagedSecret(t *testing.T) {
	dir := t.TempDir()
	caFile := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(caFile, []byte("test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	secretID := "db/dbadmin"
	restoreAWS := stubAWSForOpenAdministrator(
		t,
		Metadata{Host: "db.example", Port: 5432},
		secretID,
		`{"username":"wrong-user","password":"test-only-password"}`,
	)
	defer restoreAWS()
	db, metadata, err := OpenAdministrator(context.Background(), "app-db", secretID, caFile, "")
	if db != nil || err == nil || err.Error() != "invalid administrator secret" {
		t.Fatalf("expected validation of the explicit secret, got %v", err)
	}
	if metadata.Host != "db.example" || metadata.Port != 5432 {
		t.Fatal("expected endpoint from RDS")
	}
}

func stubAWSForOpenAdministrator(t *testing.T, metadata Metadata, expectedSecretID, secret string) func() {
	t.Helper()
	originalEndpoint := getRDSEndpoint
	originalSecret := getSecretString
	getRDSEndpoint = func(context.Context, string) (Metadata, error) {
		return metadata, nil
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
