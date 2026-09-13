package dbaccess

import (
	"context"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestConnectionURLBuildsPostgresURL(t *testing.T) {
	databaseURL := ConnectionURL("db.example", 5432, "app_user", "p@ss/word", "/tmp/rds-ca.pem")

	u, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	password, ok := u.User.Password()
	if u.Scheme != "postgres" ||
		u.Host != "db.example:5432" ||
		u.Path != "/app" ||
		u.User.Username() != "app_user" ||
		!ok ||
		password != "p@ss/word" ||
		u.Query().Get("sslmode") != "verify-full" ||
		u.Query().Get("sslrootcert") != "/tmp/rds-ca.pem" ||
		u.Query().Get("connect_timeout") != "10" {
		t.Fatalf("unexpected database url: %s", databaseURL)
	}
}

func TestOpenDatabaseReturnsDatabaseWithoutConnecting(t *testing.T) {
	db, err := OpenDatabase("postgres://app_user:test@db.example:5432/app?sslmode=verify-full", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db == nil {
		t.Fatal("expected database handle")
	}
}

func TestOpenDatabaseRejectsInvalidDatabaseURL(t *testing.T) {
	db, err := OpenDatabase("://invalid-url", "")
	if db != nil || err == nil || err.Error() != "invalid database configuration" {
		t.Fatalf("unexpected result: db=%v err=%v", db, err)
	}
}

func TestOpenDatabaseRejectsInvalidLocalForwardEndpoint(t *testing.T) {
	db, err := OpenDatabase("postgres://app_user:test@db.example:5432/app?sslmode=verify-full", "203.0.113.1:15432")
	if db != nil || err == nil || err.Error() != "local forward endpoint must be a loopback IP and valid port" {
		t.Fatalf("unexpected result: db=%v err=%v", db, err)
	}
}

func TestTunnelPreservesRDSCertificateVerification(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example:5432/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureLocalForward(config, "127.0.0.1:15432"); err != nil {
		t.Fatal(err)
	}
	if config.Host != "db.example" || config.TLSConfig == nil || config.TLSConfig.ServerName != "db.example" || config.TLSConfig.InsecureSkipVerify {
		t.Fatal("RDS certificate verification was changed")
	}
	addresses, err := config.LookupFunc(context.Background(), "db.example")
	if err != nil || len(addresses) != 1 || addresses[0] != "127.0.0.1" {
		t.Fatal("unexpected local forward address")
	}
}

func TestTunnelRejectsRemoteEndpoint(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if configureLocalForward(config, "203.0.113.1:15432") == nil {
		t.Fatal("remote local forward endpoint accepted")
	}
}

func TestTunnelRejectsInvalidPort(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if configureLocalForward(config, "127.0.0.1:0") == nil {
		t.Fatal("invalid port accepted")
	}
}

func TestTunnelRejectsMissingPort(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureLocalForward(config, "127.0.0.1"); err == nil || err.Error() != "local forward endpoint must be a loopback IP and port" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTunnelRejectsNonNumericPort(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureLocalForward(config, "127.0.0.1:postgres"); err == nil || err.Error() != "local forward endpoint must be a loopback IP and valid port" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmptyTunnelKeepsDirectConnection(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureLocalForward(config, ""); err != nil {
		t.Fatal(err)
	}
	if config.Host != "db.example" || config.Port != 5432 {
		t.Fatal("direct connection changed")
	}
}
