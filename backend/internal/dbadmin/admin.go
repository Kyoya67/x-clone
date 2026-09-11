package dbadmin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
)

// AWS passes secret values via stdin, never process arguments or temporary files.
func AWS(ctx context.Context, input any, output any, args ...string) error {
	args = append(args, "--output", "json", "--no-cli-pager")
	var stdin []byte
	if input != nil {
		var err error
		stdin, err = json.Marshal(input)
		if err != nil {
			return errors.New("cannot encode AWS request")
		}
		args = append(args, "--cli-input-json", "file:///dev/stdin")
	}
	cmd := exec.CommandContext(ctx, "aws", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	data, err := cmd.Output()
	if err != nil {
		return errors.New("AWS request failed; check operator credentials, region and permissions")
	}
	if output != nil && json.Unmarshal(data, output) != nil {
		return errors.New("invalid AWS response")
	}
	return nil
}

func ConnectionURL(host string, port int, user, password, ca string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, strconv.Itoa(port)), Path: "/app", User: url.UserPassword(user, password)}
	q := url.Values{"sslmode": {"verify-full"}, "sslrootcert": {ca}, "connect_timeout": {"10"}}
	u.RawQuery = q.Encode()
	return u.String()
}

type Metadata struct {
	Host   string
	Port   int
	Secret string
}

// OpenAdministrator reads credentials into memory and connects as dbadmin.
func OpenAdministrator(ctx context.Context, instance, caFile, tunnel string) (*sql.DB, Metadata, error) {
	var metadata Metadata
	if _, err := os.Stat(caFile); err != nil {
		return nil, metadata, errors.New("cannot read RDS CA file")
	}
	if err := AWS(ctx, nil, &metadata, "rds", "describe-db-instances", "--db-instance-identifier", instance,
		"--query", "DBInstances[0].{Host:Endpoint.Address,Port:Endpoint.Port,Secret:MasterUserSecret.SecretArn}"); err != nil {
		return nil, metadata, err
	}
	if metadata.Host == "" || metadata.Secret == "" || metadata.Port == 0 {
		return nil, metadata, errors.New("RDS endpoint or managed administrator secret is missing")
	}
	var adminValue struct{ SecretString string }
	if err := AWS(ctx, nil, &adminValue, "secretsmanager", "get-secret-value", "--secret-id", metadata.Secret); err != nil {
		return nil, metadata, err
	}
	var admin struct{ Username, Password string }
	if json.Unmarshal([]byte(adminValue.SecretString), &admin) != nil || admin.Username != "dbadmin" || admin.Password == "" {
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
