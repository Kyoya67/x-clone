package dbadmin

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestTunnelPreservesRDSCertificateVerification(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example:5432/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureTunnel(config, "127.0.0.1:15432"); err != nil {
		t.Fatal(err)
	}
	if config.Host != "db.example" || config.TLSConfig == nil || config.TLSConfig.ServerName != "db.example" || config.TLSConfig.InsecureSkipVerify {
		t.Fatal("RDS certificate verification was changed")
	}
	addresses, err := config.LookupFunc(context.Background(), "db.example")
	if err != nil || len(addresses) != 1 || addresses[0] != "127.0.0.1" {
		t.Fatal("unexpected tunnel address")
	}
}

func TestTunnelRejectsRemoteEndpoint(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if configureTunnel(config, "203.0.113.1:15432") == nil {
		t.Fatal("remote endpoint accepted")
	}
}

func TestTunnelRejectsInvalidPort(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if configureTunnel(config, "127.0.0.1:0") == nil {
		t.Fatal("invalid port accepted")
	}
}

func TestEmptyTunnelKeepsDirectConnection(t *testing.T) {
	config, err := pgx.ParseConfig("postgres://app_user:test@db.example/app?sslmode=verify-full")
	if err != nil {
		t.Fatal(err)
	}
	if err := configureTunnel(config, ""); err != nil {
		t.Fatal(err)
	}
	if config.Host != "db.example" || config.Port != 5432 {
		t.Fatal("direct connection changed")
	}
}
