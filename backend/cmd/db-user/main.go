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
	"os"
	"strings"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	instance := flag.String("instance", "app-db", "RDS instance identifier")
	adminSecretID := flag.String("admin-secret", "db/dbadmin", "Administrator username/password secret")
	appSecretID := flag.String("secret", "db/app_user", "Application username/password secret")
	migrationSecretID := flag.String("migration-secret", "db/migration_user", "Migration username/password secret")
	caFile := flag.String("ca-file", "", "Local RDS CA bundle path (required)")
	localForwardEndpoint := flag.String("local-forward-endpoint", "", "Optional local SSM port forwarding endpoint, e.g. 127.0.0.1:15432")
	flag.Parse()
	if *caFile == "" {
		fmt.Fprintln(os.Stderr, "--ca-file is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := run(ctx, *instance, *appSecretID, *migrationSecretID, *caFile, *localForwardEndpoint, *adminSecretID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("app_user and migration_user configured; username/password JSON stored in Secrets Manager")
}

func run(ctx context.Context, instance, appSecretID, migrationSecretID, caFile, localForwardEndpoint, adminSecretID string) error {
	if appSecretID == migrationSecretID {
		return errors.New("application and migration secrets must be different")
	}
	db, rdsEndpoint, err := dbadmin.OpenAdministrator(ctx, instance, adminSecretID, caFile, localForwardEndpoint)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.New("cannot start database transaction")
	}
	defer tx.Rollback()
	// このdb-user/main.goを実行中他のプロセスでこのdb-user/main.goを実行しても、同じロックを取得できず待機する。
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(81421001)"); err != nil {
		return errors.New("cannot acquire setup lock")
	}
	if err := setupUser(ctx, tx, rdsEndpoint, appSecretID, caFile, localForwardEndpoint, "app_user", configureAppRole); err != nil {
		return err
	}
	if err := setupUser(ctx, tx, rdsEndpoint, migrationSecretID, caFile, localForwardEndpoint, "migration_user", configureMigrationRole); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.New("database commit failed; rerun to reconcile with saved secrets")
	}
	return nil
}

// DBユーザーのセットアップを行う。既存のシークレットがある場合は、それを利用して接続確認を行い、必要に応じてパスワードを設定する。
func setupUser(ctx context.Context, tx *sql.Tx, rdsEndpoint dbadmin.RDSEndpoint, secretID, caFile, localForwardEndpoint, role string, configure func(context.Context, *sql.Tx, string, bool) error) error {
	hasCurrentSecret, err := dbadmin.CurrentSecretVersionExists(ctx, secretID)
	if err != nil {
		return err
	}
	existingSecretPassword := ""
	if hasCurrentSecret {
		existingSecretPassword, err = dbadmin.GetSecretString(ctx, secretID)
		if err != nil {
			return err
		}
		if existingSecretPassword == "" {
			return errors.New("database user secret is empty")
		}
	}
	password, err := databaseUserPassword(existingSecretPassword, role)
	if err != nil {
		return err
	}
	var roleExists bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)", role).Scan(&roleExists); err != nil {
		return errors.New("cannot inspect database role")
	}
	// もしRDSにDBユーザーが存在し、既存のシークレットがある場合は、その認証情報で接続できるか確認する
	if roleExists && existingSecretPassword != "" {
		appDB, err := dbadmin.OpenDatabase(dbadmin.ConnectionURL(rdsEndpoint.Host, rdsEndpoint.Port, role, password, caFile), localForwardEndpoint)
		if err != nil {
			return errors.New("cannot initialize database connection")
		}
		defer appDB.Close()
		// AWSコンソール上からsecrets managerでパスワードを変更すると既存の認証情報で接続できなくなる。
		if err := appDB.PingContext(ctx); err != nil {
			return errors.New("saved database credentials cannot connect; refusing to change password")
		}
	}
	if err := configure(ctx, tx, password, existingSecretPassword != ""); err != nil {
		return err
	}
	secretJSON, err := json.Marshal(credentials{Username: role, Password: password})
	if err != nil {
		return errors.New("cannot encode database credentials")
	}
	if existingSecretPassword == "" {
		if err := dbadmin.PutSecretString(ctx, secretID, string(secretJSON)); err != nil {
			return err
		}
	}
	return nil
}

func databaseUserPassword(existingSecretPassword, role string) (string, error) {
	if existingSecretPassword != "" {
		var saved credentials
		if err := json.Unmarshal([]byte(existingSecretPassword), &saved); err != nil || saved.Username != role || saved.Password == "" {
			return "", errors.New("invalid database credentials; refusing to overwrite")
		}
		return saved.Password, nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.New("cannot generate password")
	}
	return hex.EncodeToString(b), nil
}

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
		// 最小権限でDBロールを作成する
		literal := strings.ReplaceAll(strings.ReplaceAll(password, "\\", "\\\\"), "'", "''")
		if _, err := tx.ExecContext(ctx, "CREATE ROLE "+role+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD E'"+literal+"'"); err != nil {
			return errors.New("cannot create database role")
		}
	}
	return nil
}

func configureAppRole(ctx context.Context, tx *sql.Tx, password string, hasSecret bool) error {
	if err := createRole(ctx, tx, "app_user", password, hasSecret); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "GRANT CONNECT ON DATABASE app TO app_user"); err != nil {
		return errors.New("cannot grant database access")
	}
	if _, err := tx.ExecContext(ctx, "GRANT USAGE ON SCHEMA public TO app_user"); err != nil {
		return errors.New("cannot grant schema access")
	}
	// 今存在するテーブルに対してのみ権限を付与する
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
	if _, err := tx.ExecContext(ctx, "GRANT migration_user TO dbadmin WITH INHERIT TRUE, SET TRUE"); err != nil {
		return errors.New("cannot grant administrator membership in migration role")
	}
	if _, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto"); err != nil {
		return errors.New("cannot prepare migration extension")
	}
	return transferMigrationTables(ctx, tx)
}

func transferMigrationTables(ctx context.Context, tx *sql.Tx) error {
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
