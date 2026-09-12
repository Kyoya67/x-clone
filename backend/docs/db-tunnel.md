# DBトンネル接続

このドキュメントは、[`backend/internal/dbadmin/tunnel.go`](../internal/dbadmin/tunnel.go)の実装を説明する。

## シェルのトンネルとGoのトンネル

`make db-tunnel`は、SSMポートフォワードを開始する。

```text
開発者PCの127.0.0.1:15432
  -> SSM
  -> nat-instance
  -> RDS:5432
```

`cmd/db-user --tunnel 127.0.0.1:15432`は、すでに開いている入口をGoコードに渡すだけ。Goコード自体はSSMセッションを開始しない。

## `OpenDatabase`

```go
config, err := pgx.ParseConfig(databaseURL)
```

`databaseURL`をpgxの接続設定へ変換する。この時点ではまだDBへ接続しない。

`databaseURL`にはRDS本物のホスト名を残す。

```text
app-db.xxxxx.ap-northeast-1.rds.amazonaws.com
```

このホスト名は、RDSのTLS証明書を検証するときに使う。

```go
configureTunnel(config, tunnel)
```

`--tunnel`が指定されていれば、実際のTCP接続先だけローカルのSSM入口へ差し替える。

```go
return stdlib.OpenDB(*config), nil
```

`database/sql`で扱える`*sql.DB`を作る。実際の接続は、後続の`PingContext`や`QueryContext`などで発生する。

## `configureTunnel`

```go
if endpoint == "" {
    return nil
}
```

`--tunnel`が空なら何もしない。RDSへ直接接続する設定のままにする。

```go
host, port, err := net.SplitHostPort(endpoint)
```

`127.0.0.1:15432`をhostとportに分ける。

```text
host = 127.0.0.1
port = 15432
```

```go
ip := net.ParseIP(host)
```

hostがIPアドレスとして正しいか確認する。`localhost`のような名前は許可しない。

```go
p, err := strconv.Atoi(port)
```

portが数値か確認する。

```go
if ip == nil || !ip.IsLoopback() || err != nil || p < 1 || p > 65535 {
    return errors.New("tunnel must be a loopback IP and valid port")
}
```

`--tunnel`には、`127.0.0.1`のようなloopback IPと正しいportだけを許可する。RDSホスト名や外部ホストを渡さないためのチェック。

```go
dialer := net.Dialer{Timeout: 10 * time.Second}
```

TCP接続用の設定。10秒以内につながらなければ失敗にする。

```go
config.DialFunc = func(ctx context.Context, network, _ string) (net.Conn, error) {
    return dialer.DialContext(ctx, network, endpoint)
}
```

pgxが実際にTCP接続するときの接続先を差し替える。

通常はRDS本物のホストへ接続する。

```text
app-db.xxxxx.ap-northeast-1.rds.amazonaws.com:5432
```

この設定により、実際のTCP接続先だけSSM入口へ変わる。

```text
127.0.0.1:15432
```

ただし、`databaseURL`上のRDSホスト名は残している。そのため、TLS証明書の検証は`127.0.0.1`ではなくRDSホスト名に対して行われる。

```go
config.LookupFunc = func(context.Context, string) ([]string, error) {
    return []string{host}, nil
}
```

DNS解決を差し替える。`endpoint`のhostは検証済みのloopback IPなので、そのまま返す。

## まとめ

```text
TLS検証先: RDS本物のホスト名
TCP接続先: 127.0.0.1:15432
```

RDSはprivate subnetにあるため、開発者PCから直接接続できない。そこでTCP接続先だけSSMポートフォワードの入口へ差し替え、TLS検証ではRDS本物のホスト名を使う。
