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
	// databaseURLにはRDS本物のホスト名を入れておく。TLS検証でこのホスト名を使うため。
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
	// endpointはRDSではなく、ローカルPC上のSSMポートフォワード入口。
	// 例: 127.0.0.1:15432
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return errors.New("tunnel must be a loopback IP and port")
	}
	ip := net.ParseIP(host)
	p, err := strconv.Atoi(port)
	// ローカルのSSMトンネルだけを許可するため、127.0.0.1のようなloopback IPに限定する。
	if ip == nil || !ip.IsLoopback() || err != nil || p < 1 || p > 65535 {
		return errors.New("tunnel must be a loopback IP and valid port")
	}
	dialer := net.Dialer{Timeout: 10 * time.Second}
	// pgxの接続処理を差し替え、実際のTCP接続だけendpointへ向ける。
	// これによりTLS検証先はRDSホスト名のまま、接続先だけ127.0.0.1:15432になる。
	config.DialFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, endpoint)
	}
	// endpointのhostは検証済みのloopback IPなので、DNS解決せずそのまま使わせる。
	config.LookupFunc = func(context.Context, string) ([]string, error) { return []string{host}, nil }
	return nil
}
