# Backend

Goで実装するバックエンドAPIです。

## 必要な環境

- Go 1.25以上

## 起動

backendディレクトリで実行します。

```bash
go run ./cmd/api
```

`cmd/api/main.go`はポート設定とHTTPサーバーの起動を担当します。ルート定義やcontrollerの生成は`internal/routers/router.go`に集約しています。

デフォルトでは`http://localhost:8080`で起動します。ポートを変更する場合は`PORT`環境変数を指定します。

```bash
PORT=8081 go run ./cmd/api
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

## ディレクトリ構成

```text
backend/
├── cmd/api/             # APIサーバーの起動
│   └── main.go
├── internal/
│   ├── controllers/     # HTTPリクエスト・レスポンスの処理
│   │   └── health.go
│   └── routers/          # URLとcontrollerの紐付け
│       ├── router.go
│       └── router_test.go
├── go.mod
└── README.md
```
