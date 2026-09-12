// db-user is an operator command, not part of the HTTP server or its IAM role.
package main

import (
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
	"strconv"
	"strings"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

func main() {
	instance := flag.String("instance", "app-db", "RDS instance identifier")
	adminSecretID := flag.String("admin-secret", "db/dbadmin", "Administrator username/password secret")
	appSecretID := flag.String("secret", "db/app_user", "Application username/password secret")
	migrationSecretID := flag.String("migration-secret", "db/migration_user", "Migration username/password secret")
	caFile := flag.String("ca-file", "", "Local RDS CA bundle path (required)")
	tunnel := flag.String("tunnel", "", "Optional local SSM tunnel endpoint, e.g. 127.0.0.1:15432")
	flag.Parse()
	if *caFile == "" {
		fmt.Fprintln(os.Stderr, "--ca-file is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := run(ctx, *instance, *appSecretID, *migrationSecretID, *caFile, *tunnel, *adminSecretID); err != nil {
		// Never print SQL/driver/CLI errors, which can contain credentials.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("app_user and migration_user configured; username/password JSON stored in Secrets Manager")
}

func run(ctx context.Context, instance, appSecretID, migrationSecretID, caFile, tunnel, adminSecretID string) error {
	if appSecretID == migrationSecretID {
		return errors.New("application and migration secrets must be different")
	}
	db, metadata, err := dbadmin.OpenAdministrator(ctx, instance, adminSecretID, caFile, tunnel)
	if err != nil {
		return err
	}
	defer db.Close()
	// Serialize cooperating operators, including secret retrieval/publication.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("cannot start database transaction")
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(81421001)"); err != nil {
		return errors.New("cannot acquire setup lock")
	}
	if err := setupUser(ctx, tx, metadata, appSecretID, caFile, tunnel, "app_user", configureRole); err != nil {
		return err
	}
	if err := setupUser(ctx, tx, metadata, migrationSecretID, caFile, tunnel, "migration_user", configureMigrationRole); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.New("database commit failed; rerun to reconcile with saved secrets")
	}
	return nil
}

// 両ユーザーで認証情報の取得・再利用・保存を共通化する。
func setupUser(ctx context.Context, tx *sql.Tx, metadata dbadmin.Metadata, secretID, caFile, tunnel, role string, configure func(context.Context, *sql.Tx, string, bool) error) error {
	hasCurrentSecret, err := dbadmin.CurrentSecretVersionExists(ctx, secretID)
	if err != nil {
		return err
	}
	existingURL := ""
	if hasCurrentSecret {
		existingURL, err = dbadmin.GetSecretString(ctx, secretID)
		if err != nil {
			return err
		}
		if existingURL == "" {
			return errors.New("database user secret is empty")
		}
	}
	password, err := rolePassword(existingURL, metadata.Host, metadata.Port, role)
	if err != nil {
		return err
	}
	// Existing credentials must authenticate; never silently reset a password.
	var roleExists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&roleExists); err != nil {
		return errors.New("cannot inspect database role")
	}
	if roleExists && existingURL != "" {
		appDB, err := dbadmin.OpenDatabase(dbadmin.ConnectionURL(metadata.Host, metadata.Port, role, password, caFile), tunnel)
		if err != nil {
			return errors.New("cannot initialize database connection")
		}
		defer appDB.Close()
		if err := appDB.PingContext(ctx); err != nil {
			return errors.New("saved database credentials cannot connect; refusing to change password")
		}
	}
	if err := configure(ctx, tx, password, existingURL != ""); err != nil {
		return err
	}
	secretJSON, err := json.Marshal(credentials{Username: role, Password: password})
	if err != nil {
		return errors.New("cannot encode database credentials")
	}
	// 旧URL形式も、認証を確認した同じパスワードでJSON形式へ移行する。
	if existingURL == "" || strings.HasPrefix(existingURL, "postgres://") {
		// Publish before commit. If commit fails, retry reuses the saved password.
		if err := dbadmin.PutSecretString(ctx, secretID, string(secretJSON)); err != nil {
			return err
		}
	}
	return nil
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func rolePassword(existing, host string, port int, role string) (string, error) {
	if existing != "" && !strings.HasPrefix(existing, "postgres://") {
		var saved credentials
		if err := json.Unmarshal([]byte(existing), &saved); err != nil || saved.Username != role || saved.Password == "" {
			return "", errors.New("invalid database credentials; refusing to overwrite")
		}
		return saved.Password, nil
	}
	if existing != "" {
		u, err := url.Parse(existing)
		if err != nil || u.Scheme != "postgres" || u.Host != net.JoinHostPort(host, strconv.Itoa(port)) || u.Path != "/app" || u.User == nil || u.User.Username() != role || u.Query().Get("sslmode") != "verify-full" || u.Query().Get("sslrootcert") != "/app/certs/rds-ca-bundle.pem" {
			return "", errors.New("existing database secret does not match target; refusing to overwrite")
		}
		password, ok := u.User.Password()
		if !ok || password == "" {
			return "", errors.New("database password is missing")
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
	if err := createRole(ctx, tx, "app_user", password, hasSecret); err != nil {
		return err
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

// roleは固定のユーザー名だけを許可し、SQL識別子として安全に使う。
func createRole(ctx context.Context, tx *sql.Tx, role, password string, hasSecret bool) error {
	if role != "app_user" && role != "migration_user" {
		return errors.New("unsupported database role")
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&exists); err != nil {
		return errors.New("cannot inspect database role")
	}
	if exists && !hasSecret {
		return errors.New("database role exists without a saved secret; refusing to change its password")
	}
	if !exists {
		// PostgreSQL utility statements do not accept password bind parameters.
		// E-string escaping also handles existing passwords on recovery runs.
		literal := strings.ReplaceAll(strings.ReplaceAll(password, "\\", "\\\\"), "'", "''")
		if _, err := tx.ExecContext(ctx, "CREATE ROLE "+role+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD E'"+literal+"'"); err != nil {
			return errors.New("cannot create database role")
		}
	}
	return nil
}

func configureMigrationRole(ctx context.Context, tx *sql.Tx, password string, hasSecret bool) error {
	if err := createRole(ctx, tx, "migration_user", password, hasSecret); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "GRANT CONNECT ON DATABASE app TO migration_user"); err != nil {
		return errors.New("cannot grant migration database access")
	}
	if _, err := tx.ExecContext(ctx, "GRANT USAGE, CREATE ON SCHEMA public TO migration_user"); err != nil {
		return errors.New("cannot grant migration schema access")
	}
	// 所有権変更に必要なSET ROLEと、変更後の管理・再実行をdbadminに許可する。
	// 逆方向（migration_userへの管理者権限付与）ではない。
	if _, err := tx.ExecContext(ctx, "GRANT migration_user TO dbadmin WITH INHERIT TRUE, SET TRUE"); err != nil {
		return errors.New("cannot grant administrator membership in migration role")
	}
	// 初回SQLが必要とする拡張は管理者が準備し、migration_userへDB全体のCREATEは与えない。
	if _, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto"); err != nil {
		return errors.New("cannot prepare migration extension")
	}
	return transferMigrationTables(ctx, tx)
}

func transferMigrationTables(ctx context.Context, tx *sql.Tx) error {
	// DB全体のREASSIGN OWNEDは使わず、このアプリの4テーブルだけを対象にする。
	for _, table := range []string{"users", "posts", "follows", "schema_migrations"} {
		var present bool
		if err := tx.QueryRowContext(ctx, "SELECT to_regclass($1) IS NOT NULL", "public."+table).Scan(&present); err != nil {
			return errors.New("cannot inspect migration tables")
		}
		if !present {
			continue
		}
		if _, err := tx.ExecContext(ctx, "ALTER TABLE public."+table+" OWNER TO migration_user"); err != nil {
			return errors.New("cannot transfer migration table ownership")
		}
	}
	return nil
}
