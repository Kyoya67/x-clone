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
	// No managed-secret ARN in the RDS response. Check the explicitly supplied ARN.
	cli := `#!/bin/sh
if [ "$1" = rds ]; then
  printf '%s' '{"Host":"db.example","Port":5432}'
elif [ "$1" = secretsmanager ] && [ "$2" = get-secret-value ] && [ "$3" = --secret-id ] && [ "$4" = "$EXPECTED_ADMIN_SECRET" ]; then
  printf '%s' '{"SecretString":"{\"username\":\"wrong-user\",\"password\":\"test-only-password\"}"}'
else
  exit 1
fi
`
	if err := os.WriteFile(filepath.Join(dir, "aws"), []byte(cli), 0700); err != nil {
		t.Fatal(err)
	}
	secretID := "arn:aws:secretsmanager:ap-northeast-1:089244387218:secret:db/dbadmin-pTEXku"
	t.Setenv("PATH", dir)
	t.Setenv("EXPECTED_ADMIN_SECRET", secretID)
	db, metadata, err := OpenAdministrator(context.Background(), "app-db", secretID, caFile, "")
	if db != nil || err == nil || err.Error() != "invalid administrator secret" {
		t.Fatalf("expected validation of the explicit secret, got %v", err)
	}
	if metadata.Host != "db.example" || metadata.Port != 5432 {
		t.Fatal("expected endpoint from RDS")
	}
}

func TestOpenAdministratorRejectsMissingSecret(t *testing.T) {
	_, _, err := OpenAdministrator(context.Background(), "app-db", "", "", "")
	if err == nil || err.Error() != "administrator secret is required" {
		t.Fatalf("expected missing secret error, got %v", err)
	}
}
