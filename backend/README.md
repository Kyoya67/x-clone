# Backend

Goで実装するバックエンドAPIです。

## 使用技術

- Go
- gorilla/mux
- PostgreSQL
- pgx

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

PostgreSQLを起動します。

```bash
docker compose up -d
```

停止する場合：

```bash
docker compose down
```

### データベースマイグレーション

データベースのスキーマ変更は`golang-migrate`で管理します。PostgreSQLを起動したあと、`backend`ディレクトリで以下を実行してください。

Makefileから実行する場合：

```bash
make migrate-up
```

Makefileでは、既存の`migrate`コンテナを一時的に起動して未適用のマイグレーションを適用します。接続先を変更する場合は、`MIGRATE_DATABASE_URL`を指定します。

```bash
make migrate-up MIGRATE_DATABASE_URL='postgres://user:password@postgres:5432/db?sslmode=disable'
```

最後に適用したマイグレーションを1つ戻す場合：

```bash
make migrate-down
```

Makefileを使わず直接実行する場合：

```bash
docker compose run --rm migrate \
  -path=/migrations \
  -database 'postgres://app:app@postgres:5432/app?sslmode=disable' \
  up
```

適用済みのマイグレーションは、PostgreSQLの`schema_migrations`テーブルで管理されます。未適用のマイグレーションだけが順番に適用されるため、同じコマンドを再実行しても適用済みのSQLは再実行されません。

`down`はテーブル削除などの変更を行うため、開発データが失われる可能性があります。

`.env.example`をコピーして、環境変数を設定します。

```bash
cp .env.example .env
set -a
source .env
set +a
```

## 環境変数

`.env.example`をコピーして、データベース接続に必要な環境変数を設定します。

```bash
cp .env.example .env
set -a
source .env
set +a
```

`DATABASE_URL`でデータベース接続先を、`DATABASE_SSL_MODE`で接続時のSSL方式を変更できます。未指定の場合は、ローカル開発用のデフォルト接続先と`disable`を使用します。

```bash
DATABASE_URL=postgres://app:app@localhost:5432/app
DATABASE_SSL_MODE=disable
```

本番環境では、環境に応じて`DATABASE_SSL_MODE=require`または`verify-full`を設定します。`.env`は機密情報を含む可能性があるため、Gitへコミットしません。

## 開発サーバーの起動

```bash
go run ./cmd/api
```

`cmd/api/main.go`はポート設定とHTTPサーバーの起動を担当します。ルート定義やcontrollerの生成は`internal/routers/router.go`に集約しています。

デフォルトでは`http://localhost:8080`で起動します。ポートを変更する場合は`PORT`環境変数を指定します。

サーバー起動時には、Swagger UIのURLもログへ表示されます。デフォルト設定では`http://localhost:8080/docs`です。

```bash
PORT=8081 go run ./cmd/api
```

## OpenAPIの確認

Goバックエンドが配信するSwagger UIで、OpenAPI仕様を確認できます。

```bash
go run ./cmd/api
```

起動後、ブラウザで`http://localhost:8080/docs`を開きます。Swagger UIは同じGoサーバーから`/openapi.yaml`を読み込むため、Swagger UI専用のポートやCORS設定は必要ありません。

## テスト

バックエンドのユニットテストを実行します。現在はルーターからヘルスチェックエンドポイントを呼び出し、期待したHTTPステータスとレスポンスが返ることを確認しています。

```bash
go test ./...
```

### カバレッジ

テスト方針、テスト対象、カバレッジ結果は[`docs/api-testing.md`](docs/api-testing.md)を参照してください。

## コード整形

Goの標準フォーマッターを使用します。VS Codeでは、ルートの[`.vscode/settings.json`](../.vscode/settings.json)で保存時フォーマットを有効にしており、Go拡張機能が保存時に自動実行します。

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
    - "backend/**"
    - ".github/workflows/backend-ci.yml"
```

`develop`向けPull Requestの作成・更新時に、`backend`ディレクトリまたは`backend-ci.yml`を変更している場合、CIが実行されます。

```yaml
push:
  branches:
    - develop
  paths:
    - "backend/**"
    - ".github/workflows/backend-ci.yml"
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

CIと同じ内容をローカルで確認する場合は、`backend`ディレクトリで次を実行してください。最初のコマンドはフォーマット違反の検出、2つ目はテスト、3つ目は静的解析です。

```bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
```

## 現在の実装範囲

- HTTPサーバーの起動
- `PORT`環境変数によるポート設定
- PostgreSQLへの接続確認
- `GET /health`によるヘルスチェック

投稿、フォロー、タイムラインのAPI仕様は[`openapi/openapi.yaml`](openapi/openapi.yaml)で定義しています。各APIの実装は今後の機能Issueで行います。

## ディレクトリ構成

```text
backend/
├── cmd/api/             # APIサーバーの起動
│   └── main.go
├── docker-compose.yml    # ローカルPostgreSQL
├── Makefile               # マイグレーションコマンド
├── .env.example           # 環境変数のサンプル
├── openapi/              # API仕様
│   └── openapi.yaml
├── internal/
│   ├── controllers/      # HTTPリクエスト・レスポンスの処理
│   │   └── health.go
│   └── routers/          # URLとcontrollerの紐付け
│       ├── router.go
│       └── router_test.go
├── go.mod
└── README.md
```
