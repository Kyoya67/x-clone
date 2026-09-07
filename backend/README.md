# Backend

Goで実装するバックエンドAPIです。

## 必要な環境

- Go 1.25以上

## 起動

backendディレクトリで実行します。

```bash
go run ./cmd/server
```

デフォルトでは`http://localhost:8080`で起動します。ポートを変更する場合は`PORT`環境変数を指定します。

```bash
PORT=8081 go run ./cmd/server
```

## テスト

```bash
go test ./...
```

## 現在の実装範囲

- HTTPサーバーの起動
- `PORT`環境変数によるポート設定
- `GET /health`によるヘルスチェック

投稿、フォロー、タイムラインなどのAPIは、OpenAPIで仕様を定義したうえで今後実装します。
