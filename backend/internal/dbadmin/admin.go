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
func OpenAdministrator(ctx context.Context, instance, adminSecretID, caFile, localForwardEndpoint string) (*sql.DB, RDSEndpoint, error) {
	var rdsEndpoint RDSEndpoint
	if adminSecretID == "" {
		return nil, rdsEndpoint, errors.New("administrator secret is required")
	}
	if _, err := os.Stat(caFile); err != nil {
		return nil, rdsEndpoint, errors.New("cannot read RDS CA file")
	}
	rdsEndpoint, err := getRDSEndpoint(ctx, instance)
	if err != nil {
		return nil, rdsEndpoint, err
	}
	if rdsEndpoint.Host == "" || rdsEndpoint.Port == 0 {
		return nil, rdsEndpoint, errors.New("RDS endpoint is missing")
	}
	adminSecret, err := getSecretString(ctx, adminSecretID)
	if err != nil {
		return nil, rdsEndpoint, err
	}
	var admin struct{ Username, Password string }
	if json.Unmarshal([]byte(adminSecret), &admin) != nil || admin.Username != "dbadmin" || admin.Password == "" {
		return nil, rdsEndpoint, errors.New("invalid administrator secret")
	}
	adminURL := ConnectionURL(rdsEndpoint.Host, rdsEndpoint.Port, admin.Username, admin.Password, caFile)
	db, err := OpenDatabase(adminURL, localForwardEndpoint)
	if err != nil {
		return nil, rdsEndpoint, errors.New("cannot initialize database connection")
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, rdsEndpoint, errors.New("database connection failed; check VPC connectivity, CA and administrator credentials")
	}

	return db, rdsEndpoint, nil
}
