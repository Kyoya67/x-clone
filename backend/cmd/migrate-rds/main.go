// migrate-rds runs as a standalone ECS task, not as part of the API server.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/dbadmin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	migrationsource "github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	ca := flag.String("ca-file", "/app/certs/rds-ca-bundle.pem", "RDS CA file")
	path := flag.String("path", "migrations", "Migration SQL directory")
	action := flag.String("action", "status", "status or up")
	flag.Parse()
	if *ca == "" || (*action != "status" && *action != "up") {
		fmt.Fprintln(os.Stderr, "--ca-file is required; --action must be status or up")
		os.Exit(1)
	}
	if err := run(*ca, *path, *action); err != nil {
		// Raw SQL/driver/AWS errors may contain credentials or data.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ECS injects username/password from the JSON secret. No AWS CLI/SDK or SSM is used here.
func databaseURL(ca string) (string, error) {
	host, user, password := os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if host == "" || user != "migration_user" || password == "" || err != nil || port < 1 || port > 65535 {
		return "", errors.New("DB_HOST, DB_PORT, DB_USER=migration_user and DB_PASSWORD are required")
	}
	return dbadmin.ConnectionURL(host, port, user, password, ca), nil
}

func run(ca, path, action string) error {
	// Validate migration files before connecting to AWS or modifying the DB.
	source, err := iofs.New(os.DirFS(path), ".")
	if err != nil {
		return errors.New("cannot read migration files")
	}
	defer source.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	url, err := databaseURL(ca)
	if err != nil {
		return err
	}
	db, err := dbadmin.OpenDatabase(url, "")
	if err != nil {
		return errors.New("cannot initialize migration connection; check DB settings and RDS CA")
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return errors.New("cannot connect to RDS; check network, certificate and credentials")
	}
	fmt.Printf("Target: %s, database=app, user=migration_user\n", os.Getenv("DB_HOST"))
	if action == "status" {
		var exists bool
		if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.schema_migrations') IS NOT NULL").Scan(&exists); err != nil {
			return errors.New("cannot inspect migration status")
		}
		if !exists {
			fmt.Println("No migrations applied (schema_migrations does not exist)")
			return nil
		}
		var version int64
		var dirty bool
		if err := db.QueryRowContext(ctx, "SELECT version, dirty FROM public.schema_migrations").Scan(&version, &dirty); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Println("No migrations applied (schema_migrations is empty)")
				return nil
			}
			return errors.New("cannot read migration version")
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
		return nil
	}
	return migrateUp(db, source)
}

func migrateUp(db *sql.DB, source migrationsource.Driver) error {
	// Reuse the same schema_migrations table and locking as the local migrate CLI.
	driver, err := postgres.WithInstance(db, &postgres.Config{
		DatabaseName: "app", SchemaName: "public", StatementTimeout: 60 * time.Second,
	})
	if err != nil {
		return errors.New("cannot initialize migration driver")
	}
	defer driver.Close()
	m, err := migrate.NewWithInstance("iofs", source, "app", driver)
	if err != nil {
		return errors.New("cannot initialize migrations")
	}
	m.LockTimeout = 15 * time.Second
	if err := apply(m.Up); err != nil {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return errors.New("cannot verify migration version")
	}
	fmt.Printf("Migrations complete: version=%d dirty=%t\n", version, dirty)
	return nil
}

func apply(up func() error) error {
	err := up()
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	var dirty migrate.ErrDirty
	if errors.As(err, &dirty) {
		return fmt.Errorf("migration version %d is dirty; inspect the database before retrying; do not force blindly", dirty.Version)
	}
	return errors.New("migration failed; inspect migration status and database logs before retrying")
}
