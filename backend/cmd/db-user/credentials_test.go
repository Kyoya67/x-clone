package main

import (
	"strings"
	"testing"
)

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
