package dbadmin

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func OpenDatabase(databaseURL, tunnel string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	if err := configureTunnel(config, tunnel); err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*config), nil
}

func configureTunnel(config *pgx.ConnConfig, endpoint string) error {
	if endpoint == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return errors.New("tunnel must be a loopback IP and port")
	}
	ip := net.ParseIP(host)
	p, err := strconv.Atoi(port)
	if ip == nil || !ip.IsLoopback() || err != nil || p < 1 || p > 65535 {
		return errors.New("tunnel must be a loopback IP and valid port")
	}
	// Keep Host/TLSConfig unchanged: verify the RDS hostname, not localhost.
	dialer := net.Dialer{Timeout: 10 * time.Second}
	config.DialFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, endpoint)
	}
	// The SSM-managed host resolves the RDS hostname; local DNS resolution is unnecessary.
	config.LookupFunc = func(context.Context, string) ([]string, error) { return []string{host}, nil }
	return nil
}
