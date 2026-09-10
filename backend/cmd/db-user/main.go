// db-user is an operator command, not part of the HTTP server or its IAM role.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	instance := flag.String("instance", "app-db", "RDS instance identifier")
	secretID := flag.String("secret", "backend/database-url", "Application URL secret")
	caFile := flag.String("ca-file", "", "Local RDS CA bundle path (required)")
	tunnel := flag.String("tunnel", "", "Optional local SSM tunnel endpoint, e.g. 127.0.0.1:15432")
	flag.Parse()
	if *caFile == "" {
		fmt.Fprintln(os.Stderr, "--ca-file is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := run(ctx, *instance, *secretID, *caFile, *tunnel); err != nil {
		// Never print SQL/driver/CLI errors, which can contain credentials.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("app_user configured; application URL stored in Secrets Manager")
}

// aws passes secret values via stdin, never process arguments or temporary files.
func aws(ctx context.Context, input any, output any, args ...string) error {
	args = append(args, "--output", "json", "--no-cli-pager")
	var stdin []byte
	if input != nil {
		var err error
		stdin, err = json.Marshal(input)
		if err != nil {
			return errors.New("cannot encode AWS request")
		}
		args = append(args, "--cli-input-json", "file:///dev/stdin")
	}
	cmd := exec.CommandContext(ctx, "aws", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	data, err := cmd.Output()
	if err != nil {
		return errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	if output != nil && json.Unmarshal(data, output) != nil {
		return errors.New("invalid AWS response")
	}
	return nil
}

func run(ctx context.Context, instance, secretID, caFile, tunnel string) error {
	if _, err := os.Stat(caFile); err != nil {
		return errors.New("cannot read RDS CA file")
	}
	var metadata struct {
		Host   string
		Port   int
		Secret string
	}
	if err := aws(ctx, nil, &metadata, "rds", "describe-db-instances", "--db-instance-identifier", instance,
		"--query", "DBInstances[0].{Host:Endpoint.Address,Port:Endpoint.Port,Secret:MasterUserSecret.SecretArn}"); err != nil {
		return err
	}
	if metadata.Host == "" || metadata.Secret == "" || metadata.Port == 0 {
		return errors.New("RDS endpoint or managed administrator secret is missing")
	}
	var adminValue struct{ SecretString string }
	if err := aws(ctx, nil, &adminValue, "secretsmanager", "get-secret-value", "--secret-id", metadata.Secret); err != nil {
		return err
	}
	var admin struct{ Username, Password string }
	if json.Unmarshal([]byte(adminValue.SecretString), &admin) != nil || admin.Username != "dbadmin" || admin.Password == "" {
		return errors.New("invalid administrator secret")
	}
	adminURL := connectionURL(metadata.Host, metadata.Port, admin.Username, admin.Password, caFile)
	db, err := openDatabase(adminURL, tunnel)
	if err != nil {
		return errors.New("cannot initialize database connection")
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return errors.New("database connection failed; check VPC connectivity, CA and administrator credentials")
	}
	// Serialize cooperating operators, including secret retrieval/publication.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("cannot start database transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(81421001)"); err != nil {
		return errors.New("cannot acquire setup lock")
	}
	var description struct{ VersionIdsToStages map[string][]string }
	if err := aws(ctx, nil, &description, "secretsmanager", "describe-secret", "--secret-id", secretID); err != nil {
		return err
	}
	existingURL := ""
	for _, stages := range description.VersionIdsToStages {
		for _, stage := range stages {
			if stage == "AWSCURRENT" {
				var value struct{ SecretString string }
				if err := aws(ctx, nil, &value, "secretsmanager", "get-secret-value", "--secret-id", secretID); err != nil {
					return err
				}
				existingURL = value.SecretString
				if existingURL == "" {
					return errors.New("application secret is empty")
				}
			}
		}
	}
	password, err := applicationPassword(existingURL, metadata.Host, metadata.Port)
	if err != nil {
		return err
	}
	// Existing credentials must authenticate; never silently reset a password.
	var roleExists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user')").Scan(&roleExists); err != nil {
		return errors.New("cannot inspect application role")
	}
	if roleExists && existingURL != "" {
		appDB, err := openDatabase(connectionURL(metadata.Host, metadata.Port, "app_user", password, caFile), tunnel)
		if err != nil {
			return errors.New("cannot initialize application connection")
		}
		defer appDB.Close()
		if err := appDB.PingContext(ctx); err != nil {
			return errors.New("saved application credentials cannot connect; refusing to change password")
		}
	}
	if err := configureRole(ctx, tx, password, existingURL != ""); err != nil {
		return err
	}
	appURL := connectionURL(metadata.Host, metadata.Port, "app_user", password, "/app/certs/rds-ca-bundle.pem")
	if existingURL == "" {
		// Publish before commit. If commit fails, retry reuses the saved password.
		if err := aws(ctx, map[string]string{"SecretId": secretID, "SecretString": appURL}, nil,
			"secretsmanager", "put-secret-value"); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.New("database commit failed; rerun to reconcile with saved secret")
	}
	return nil
}

func connectionURL(host string, port int, user, password, ca string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, strconv.Itoa(port)), Path: "/app", User: url.UserPassword(user, password)}
	q := url.Values{"sslmode": {"verify-full"}, "sslrootcert": {ca}, "connect_timeout": {"10"}}
	u.RawQuery = q.Encode()
	return u.String()
}

func applicationPassword(existing, host string, port int) (string, error) {
	if existing != "" {
		u, err := url.Parse(existing)
		if err != nil || u.Scheme != "postgres" || u.Host != net.JoinHostPort(host, strconv.Itoa(port)) || u.Path != "/app" || u.User == nil || u.User.Username() != "app_user" || u.Query().Get("sslmode") != "verify-full" || u.Query().Get("sslrootcert") != "/app/certs/rds-ca-bundle.pem" {
			return "", errors.New("existing application secret does not match target; refusing to overwrite")
		}
		password, ok := u.User.Password()
		if !ok || password == "" {
			return "", errors.New("application password is missing")
		}
		return password, nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("cannot generate password")
	}
	return hex.EncodeToString(b), nil
}

func configureRole(ctx context.Context, tx *sql.Tx, password string, hasSecret bool) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user')").Scan(&exists); err != nil {
		return errors.New("cannot inspect application role")
	}
	if exists && !hasSecret {
		return errors.New("app_user exists without a saved secret; refusing to change its password")
	}
	if !exists {
		// PostgreSQL utility statements do not accept password bind parameters.
		// E-string escaping also handles existing passwords on recovery runs.
		literal := strings.ReplaceAll(strings.ReplaceAll(password, "\\", "\\\\"), "'", "''")
		if _, err := tx.ExecContext(ctx, "CREATE ROLE app_user LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD E'"+literal+"'"); err != nil {
			return errors.New("cannot create application role")
		}
	}
	if _, err := tx.ExecContext(ctx, "GRANT CONNECT ON DATABASE app TO app_user"); err != nil {
		return errors.New("cannot grant database access")
	}
	if _, err := tx.ExecContext(ctx, "GRANT USAGE ON SCHEMA public TO app_user"); err != nil {
		return errors.New("cannot grant schema access")
	}
	// Explicit tables only: do not grant access to schema_migrations or future tables.
	for _, table := range []string{"users", "posts", "follows"} {
		var present bool
		if err := tx.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&present); err != nil {
			return errors.New("cannot inspect application tables")
		}
		if !present {
			continue
		}
		if _, err := tx.ExecContext(ctx, "GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public."+table+" TO app_user"); err != nil {
			return errors.New("cannot grant table privileges")
		}
	}
	return nil
}
