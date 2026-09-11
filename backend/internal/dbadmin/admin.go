package dbadmin

import (
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

// AWSはJSONを権限0600の一時ファイルで渡す。秘密値を引数・ログに出さない。
func AWS(ctx context.Context, input any, output any, args ...string) (resultErr error) {
	args = append(args, "--output", "json", "--no-cli-pager")
	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return errors.New("cannot encode AWS request")
		}
		// CLIが複数回読んでも同じ内容を取得できるよう、標準入力は使用しない。
		file, err := os.CreateTemp("", "x-clone-aws-request-*.json")
		if err != nil {
			return errors.New("cannot create private AWS request file")
		}
		defer func() {
			if err := os.Remove(file.Name()); err != nil {
				resultErr = errors.New("cannot remove private AWS request file; check temporary directory")
			}
		}()
		_, writeErr := file.Write(payload)
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			return errors.New("cannot write private AWS request file")
		}
		args = append(args, "--cli-input-json", "file://"+file.Name())
	}
	cmd := exec.CommandContext(ctx, "aws", args...)
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
	Host string
	Port int
}

// OpenAdministrator reads credentials into memory and connects as dbadmin.
func OpenAdministrator(ctx context.Context, instance, secretID, caFile, tunnel string) (*sql.DB, Metadata, error) {
	var metadata Metadata
	if secretID == "" {
		return nil, metadata, errors.New("administrator secret is required")
	}
	if _, err := os.Stat(caFile); err != nil {
		return nil, metadata, errors.New("cannot read RDS CA file")
	}
	if err := AWS(ctx, nil, &metadata, "rds", "describe-db-instances", "--db-instance-identifier", instance,
		"--query", "DBInstances[0].{Host:Endpoint.Address,Port:Endpoint.Port}"); err != nil {
		return nil, metadata, err
	}
	if metadata.Host == "" || metadata.Port == 0 {
		return nil, metadata, errors.New("RDS endpoint is missing")
	}
	var adminValue struct{ SecretString string }
	if err := AWS(ctx, nil, &adminValue, "secretsmanager", "get-secret-value", "--secret-id", secretID); err != nil {
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
