# Backend

Goで実装するバックエンドAPIです。

## 使用技術

- Go
- gorilla/mux

## 必要な環境

- Go 1.25.5以上

Goのバージョン確認：

```bash
go version
```

## 環境構築

backendディレクトリで実行します。

```bash
go mod download
```

`go.mod`と`go.sum`に記録された依存関係をダウンロードします。

## 開発サーバーの起動

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

## コード整形

```bash
gofmt -w .
```

## CI

### 実行条件

```yaml
pull_request:
  branches:
    - develop
  paths:
    - 'backend/**'
    - '.github/workflows/backend-ci.yml'
```

`develop`向けPull Requestの作成・更新時に、`backend`ディレクトリまたは`backend-ci.yml`を変更している場合、CIが実行されます。

```yaml
push:
  branches:
    - develop
  paths:
    - 'backend/**'
    - '.github/workflows/backend-ci.yml'
```

push先のブランチが`develop`で、かつ`backend`ディレクトリまたは`backend-ci.yml`を変更した場合、CIが実行されます。

### 実行内容

```text
test -z "$(gofmt -l .)"
  ↓
go test ./...
  ↓
go vet ./...
```

### ローカルでの確認

CIと同じ内容をローカルで確認する場合は、`backend`ディレクトリで次を実行してください。

```bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
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
│   ├── controllers/      # HTTPリクエスト・レスポンスの処理
│   │   └── health.go
│   └── routers/          # URLとcontrollerの紐付け
│       ├── router.go
│       └── router_test.go
├── go.mod
└── README.md
```
