package dbadmin

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAWSRequestHelper(t *testing.T) {
	if os.Getenv("AWS_REQUEST_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if strings.Contains(arg, "test-only-password") {
			os.Exit(2)
		}
		if arg != "--cli-input-json" || i+1 >= len(os.Args) {
			continue
		}
		path := strings.TrimPrefix(os.Args[i+1], "file://")
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0600 {
			os.Exit(3)
		}
		first, err := os.ReadFile(path)
		if err != nil || !json.Valid(first) {
			os.Exit(4)
		}
		second, err := os.ReadFile(path)
		if err != nil || string(first) != string(second) {
			os.Exit(5)
		}
		if os.Getenv("AWS_REQUEST_FAIL") == "1" {
			os.Stderr.WriteString("test-only-password")
			os.Exit(6)
		}
		os.Stdout.WriteString(`{"OK":true}`)
		os.Exit(0)
	}
	os.Exit(7)
}

func prepareAWSHelper(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	requests := t.TempDir()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	stub := "#!/bin/sh\nexec \"$AWS_TEST_BINARY\" -test.run=^TestAWSRequestHelper$ -- \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "aws"), []byte(stub), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("TMPDIR", requests)
	t.Setenv("AWS_TEST_BINARY", binary)
	t.Setenv("AWS_REQUEST_HELPER", "1")
	return requests
}

func assertRequestFilesRemoved(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal("temporary request files were not removed")
	}
}

func TestAWSUsesPrivateReusableFileAndRemovesIt(t *testing.T) {
	dir := prepareAWSHelper(t)
	var output struct{ OK bool }
	err := AWS(context.Background(), map[string]string{"SecretString": "test-only-password"}, &output, "secretsmanager", "put-secret-value")
	if err != nil || !output.OK {
		t.Fatalf("request failed: %v", err)
	}
	assertRequestFilesRemoved(t, dir)
}

func TestAWSRemovesRequestFileOnFailure(t *testing.T) {
	dir := prepareAWSHelper(t)
	t.Setenv("AWS_REQUEST_FAIL", "1")
	err := AWS(context.Background(), map[string]string{"SecretString": "test-only-password"}, nil, "secretsmanager", "put-secret-value")
	if err == nil || strings.Contains(err.Error(), "test-only-password") {
		t.Fatal("expected sanitized failure")
	}
	assertRequestFilesRemoved(t, dir)
}

func TestAWSCLIParsesRequestOffline(t *testing.T) {
	if _, err := exec.LookPath("aws"); err != nil {
		t.Skip("AWS CLI not installed")
	}
	dir := t.TempDir()
	t.Setenv("TMPDIR", dir)
	t.Setenv("AWS_ACCESS_KEY_ID", "test-only")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-only")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	// Input skeleton generation parses cli-input-json without sending an AWS request.
	err := AWS(context.Background(), map[string]string{"SecretId": "db/app_user", "SecretString": "test-only-password"}, nil,
		"secretsmanager", "put-secret-value", "--generate-cli-skeleton", "input", "--region", "ap-northeast-1")
	if err != nil {
		t.Fatal(err)
	}
	assertRequestFilesRemoved(t, dir)
}
