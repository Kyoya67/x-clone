package dbaccess

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type RDSEndpoint struct {
	Host string
	Port int
}

func ConnectionURL(host string, port int, user, password, ca string) string {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, strconv.Itoa(port)), Path: "/app", User: url.UserPassword(user, password)}
	q := url.Values{"sslmode": {"verify-full"}, "sslrootcert": {ca}, "connect_timeout": {"10"}}
	u.RawQuery = q.Encode()
	return u.String()
}

func OpenDatabase(databaseURL, localForwardEndpoint string) (*sql.DB, error) {
	// databaseURLにはRDS本物のホスト名を入れておく。TLS検証でこのホスト名を使うため。
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.New("invalid database configuration")
	}
	if err := configureLocalForward(config, localForwardEndpoint); err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*config), nil
}

func configureLocalForward(config *pgx.ConnConfig, localForwardEndpoint string) error {
	if localForwardEndpoint == "" {
		return nil
	}
	// localForwardEndpointはRDSではなく、ローカルPC上のSSMポートフォワード入口。
	// 例: 127.0.0.1:15432
	localHost, localPort, err := net.SplitHostPort(localForwardEndpoint)
	if err != nil {
		return errors.New("local forward endpoint must be a loopback IP and port")
	}
	localIP := net.ParseIP(localHost)
	portNumber, err := strconv.Atoi(localPort)
	// ローカルのSSMトンネルだけを許可するため、127.0.0.1のようなloopback IPに限定する。
	if localIP == nil || !localIP.IsLoopback() || err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("local forward endpoint must be a loopback IP and valid port")
	}
	dialer := net.Dialer{Timeout: 10 * time.Second}
	// pgxの接続処理を差し替え、実際のTCP接続だけlocalForwardEndpointへ向ける。
	// これによりTLS検証先はRDSホスト名のまま、接続先だけ127.0.0.1:15432になる。
	config.DialFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, localForwardEndpoint)
	}
	// localHostは検証済みのloopback IPなので、DNS解決せずそのまま使わせる。
	config.LookupFunc = func(context.Context, string) ([]string, error) { return []string{localHost}, nil }
	return nil
}
