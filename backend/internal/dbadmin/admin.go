package dbadmin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
)

func ConnectionURL(host string, port int, user, password, ca string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, strconv.Itoa(port)), Path: "/app", User: url.UserPassword(user, password)}
	q := url.Values{"sslmode": {"verify-full"}, "sslrootcert": {ca}, "connect_timeout": {"10"}}
	u.RawQuery = q.Encode()
	return u.String()
}

type Metadata struct {
	Host string
	Port int
}

var (
	getRDSEndpoint  = GetRDSEndpoint
	getSecretString = GetSecretString
)

// OpenAdministrator reads dbadmin credentials into memory and connects to RDS.
func OpenAdministrator(ctx context.Context, instance, adminSecretID, caFile, tunnel string) (*sql.DB, Metadata, error) {
	var metadata Metadata
	if adminSecretID == "" {
		return nil, metadata, errors.New("administrator secret is required")
	}
	if _, err := os.Stat(caFile); err != nil {
		return nil, metadata, errors.New("cannot read RDS CA file")
	}
	metadata, err := getRDSEndpoint(ctx, instance)
	if err != nil {
		return nil, metadata, err
	}
	if metadata.Host == "" || metadata.Port == 0 {
		return nil, metadata, errors.New("RDS endpoint is missing")
	}
	adminSecret, err := getSecretString(ctx, adminSecretID)
	if err != nil {
		return nil, metadata, err
	}
	var admin struct{ Username, Password string }
	if json.Unmarshal([]byte(adminSecret), &admin) != nil || admin.Username != "dbadmin" || admin.Password == "" {
		return nil, metadata, errors.New("invalid administrator secret")
	}
	adminURL := ConnectionURL(metadata.Host, metadata.Port, admin.Username, admin.Password, caFile)
	db, err := OpenDatabase(adminURL, tunnel)
	if err != nil {
		return nil, metadata, errors.New("cannot initialize database connection")
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, metadata, errors.New("database connection failed; check VPC connectivity, CA and administrator credentials")
	}

	return db, metadata, nil
}
