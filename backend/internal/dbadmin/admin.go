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

type RDSEndpoint struct {
	Host string
	Port int
}

var (
	getRDSEndpoint  = GetRDSEndpoint
	getSecretString = GetSecretString
)

// OpenAdministrator reads dbadmin credentials into memory and connects to RDS.
func OpenAdministrator(ctx context.Context, instance, adminSecretID, caFile, tunnel string) (*sql.DB, RDSEndpoint, error) {
	var endpoint RDSEndpoint
	if adminSecretID == "" {
		return nil, endpoint, errors.New("administrator secret is required")
	}
	if _, err := os.Stat(caFile); err != nil {
		return nil, endpoint, errors.New("cannot read RDS CA file")
	}
	endpoint, err := getRDSEndpoint(ctx, instance)
	if err != nil {
		return nil, endpoint, err
	}
	if endpoint.Host == "" || endpoint.Port == 0 {
		return nil, endpoint, errors.New("RDS endpoint is missing")
	}
	adminSecret, err := getSecretString(ctx, adminSecretID)
	if err != nil {
		return nil, endpoint, err
	}
	var admin struct{ Username, Password string }
	if json.Unmarshal([]byte(adminSecret), &admin) != nil || admin.Username != "dbadmin" || admin.Password == "" {
		return nil, endpoint, errors.New("invalid administrator secret")
	}
	adminURL := ConnectionURL(endpoint.Host, endpoint.Port, admin.Username, admin.Password, caFile)
	db, err := OpenDatabase(adminURL, tunnel)
	if err != nil {
		return nil, endpoint, errors.New("cannot initialize database connection")
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, endpoint, errors.New("database connection failed; check VPC connectivity, CA and administrator credentials")
	}

	return db, endpoint, nil
}
