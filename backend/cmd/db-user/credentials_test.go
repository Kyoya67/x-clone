package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

func TestRolePasswordReadsJSONCredentials(t *testing.T) {
	password, err := rolePassword(`{"username":"migration_user","password":"test-only!#$"}`, "db.example", 5432, "migration_user")
	if err != nil || password != "test-only!#$" {
		t.Fatal("expected saved JSON password")
	}
}

func TestRolePasswordRejectsWrongJSONUser(t *testing.T) {
	_, err := rolePassword(`{"username":"dbadmin","password":"test-only-private"}`, "db.example", 5432, "app_user")
	if err == nil || strings.Contains(err.Error(), "test-only-private") {
		t.Fatal("expected safe rejection")
	}
}

func TestRolePasswordRejectsMissingJSONPassword(t *testing.T) {
	if _, err := rolePassword(`{"username":"app_user"}`, "db.example", 5432, "app_user"); err == nil {
		t.Fatal("expected missing password error")
	}
}

func TestLegacyURLConvertsToJSONWithoutChangingPassword(t *testing.T) {
	legacy := dbadmin.ConnectionURL("db.example", 5432, "app_user", "test-only:@/?#'\\password", "/app/certs/rds-ca-bundle.pem")
	password, err := rolePassword(legacy, "db.example", 5432, "app_user")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(credentials{Username: "app_user", Password: password})
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]string
	if err := json.Unmarshal(encoded, &keys); err != nil || len(keys) != 2 || keys["username"] != "app_user" || keys["password"] != "test-only:@/?#'\\password" {
		t.Fatal("expected username/password keys only")
	}
	reused, err := rolePassword(string(encoded), "db.example", 5432, "app_user")
	if err != nil || reused != password {
		t.Fatal("JSON rerun changed password")
	}
}
