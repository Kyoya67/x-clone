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
	"io"
	"os"
	"strings"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
)

var openAdministrator = dbadmin.OpenAdministrator
var runCommand = run
var exit = os.Exit
var randomRead = rand.Read
var standardOutput io.Writer = os.Stdout
var standardError io.Writer = os.Stderr

var (
	currentSecretVersionExists = dbadmin.CurrentSecretVersionExists
	getSecretString            = dbadmin.GetSecretString
	putSecretString            = dbadmin.PutSecretString
	connectionURL              = dbadmin.ConnectionURL
	openDatabase               = dbadmin.OpenDatabase
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	exit(runCLI(os.Args[1:], standardOutput, standardError))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("db-user", flag.ContinueOnError)
	flags.SetOutput(stderr)
	instance := flags.String("instance", "app-db", "RDS instance identifier")
	adminSecretID := flags.String("admin-secret", "db/dbadmin", "Administrator username/password secret")
	appSecretID := flags.String("secret", "db/app_user", "Application username/password secret")
	migrationSecretID := flags.String("migration-secret", "db/migration_user", "Migration username/password secret")
	caFile := flags.String("ca-file", "", "Local RDS CA bundle path (required)")
	localForwardEndpoint := flags.String("local-forward-endpoint", "", "Optional local SSM port forwarding endpoint, e.g. 127.0.0.1:15432")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *caFile == "" {
		fmt.Fprintln(stderr, "--ca-file is required")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := runCommand(ctx, *instance, *appSecretID, *migrationSecretID, *caFile, *localForwardEndpoint, *adminSecretID); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintln(stdout, "app_user and migration_user configured; username/password JSON stored in Secrets Manager")
	return 0
}

func run(ctx context.Context, instance, appSecretID, migrationSecretID, caFile, localForwardEndpoint, adminSecretID string) error {
	if appSecretID == migrationSecretID {
		return errors.New("application and migration secrets must be different")
	}
	db, rdsEndpoint, err := openAdministrator(ctx, instance, adminSecretID, caFile, localForwardEndpoint)
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
	hasCurrentSecret, err := currentSecretVersionExists(ctx, secretID)
	if err != nil {
		return err
	}
	existingSecretPassword := ""
	if hasCurrentSecret {
		existingSecretPassword, err = getSecretString(ctx, secretID)
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
		appDB, err := openDatabase(connectionURL(rdsEndpoint.Host, rdsEndpoint.Port, role, password, caFile), localForwardEndpoint)
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
		if err := putSecretString(ctx, secretID, string(secretJSON)); err != nil {
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
	if _, err := randomRead(b); err != nil {
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
		// 最小権限でDBロールを作成し、細かい権限設定はそれぞれconfigureAppRoleやconfigureMigrationRoleで行う。
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
	if _, err := tx.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS migration AUTHORIZATION migration_user"); err != nil {
		return errors.New("cannot prepare migration schema")
	}
	if _, err := tx.ExecContext(ctx, "GRANT USAGE, CREATE ON SCHEMA migration TO migration_user"); err != nil {
		return errors.New("cannot grant migration metadata schema access")
	}
	if _, err := tx.ExecContext(ctx, "ALTER DEFAULT PRIVILEGES FOR ROLE migration_user IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user"); err != nil {
		return errors.New("cannot configure application default table privileges")
	}
	if _, err := tx.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pgcrypto"); err != nil {
		return errors.New("cannot prepare migration extension")
	}
	return nil
}
