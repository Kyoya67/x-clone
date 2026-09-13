# X Clone

X風の投稿・フォロー・タイムライン機能を持つWebアプリケーションです。

このリポジトリは、フロントエンド、バックエンド、AWSインフラを1つのリポジトリで管理します。ルートのREADMEでは全体の環境構築手順だけをまとめ、各領域の詳細はそれぞれのREADMEを参照します。

## 環境別デプロイURL

| 環境 | フロントエンド | API | 備考 |
| ---- | ---- | ---- | ---- |
| stg | https://stg.x-clone.kyo8.dev | https://api-v1.stg.x-clone.kyo8.dev | stg向けの開発検証環境 |
| prd | https://x-clone.kyo8.dev | https://api-v1.x-clone.kyo8.dev | 本番想定環境 |

stg・prdの各APIは、Cognito Hosted UI経由でのGoogleログインを前提にしています。

## ディレクトリ構成

`````text
.
├── frontend/        # React + TypeScript + Vite
├── backend/         # Go API、DB migration、DBユーザー管理コマンド
├── infrastructure/  # TerraformによるAWS環境構築
├── DESIGN.md        # 技術選定、設計判断、CI/CD方針
└── AGENTS.md        # 開発・Issue・PR・テスト・インフラ作業ルール
`````

## 詳細READMEの役割

| README | 役割 |
| ---- | ---- |
| [frontend/README.md](./frontend/README.md) | フロントエンドのローカル起動、テスト、ビルド |
| [backend/README.md](./backend/README.md) | バックエンドAPI、ローカルDB、マイグレーション、テスト |
| [infrastructure/README.md](./infrastructure/README.md) | Terraform、AWS環境、ECR push、RDS初期設定、ECS反映 |

## 必要な環境

| 用途 | 必要なもの |
| ---- | ---- |
| フロントエンド | Node.js、npm |
| バックエンド | Go、Docker |
| インフラ | Terraform、AWS CLI、jq、Docker |

## ローカル環境構築

### 1. フロントエンド

`````bash
cd frontend
npm ci
npm run dev
`````

テスト・ビルド：

`````bash
npm run format:check
npm run test
npm run build
`````

### 2. バックエンド

`````bash
cd backend
go mod download
cp .env.example .env
docker compose up -d
make migrate-up
go run ./cmd/api
`````

テスト：

`````bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
`````

## AWS環境構築

AWS環境はTerraformで管理します。

| 環境 | Terraform root module | AWS Profile |
| ---- | ---- | ---- |
| stg | `infrastructure/stg` | `x-clone-terraform-stg` |
| prd | `infrastructure/prd` | `x-clone-terraform-prd` |

基本手順：

`````bash
cd infrastructure/prd
cp .env.example .env
AWS_PROFILE=x-clone-terraform-prd make plan
AWS_PROFILE=x-clone-terraform-prd make apply
`````

`.env`にはRDS管理者ユーザー`dbadmin`のパスワードを設定します。`.env`はGitへコミットしません。

AWS初回構築では、ECRへのDockerイメージpush、DBユーザー作成、RDSマイグレーション、ECSサービス反映の順序が必要です。詳細は[infrastructure/README.md](./infrastructure/README.md)を参照してください。

## デプロイ概要

| 対象 | 方法 |
| ---- | ---- |
| フロントエンド | Amplify Hostingへデプロイ |
| API | Docker imageをECRへpushし、ECS Serviceで常時起動 |
| マイグレーション | `db-migrator` ECS単発タスクで実行 |
| DB | RDS PostgreSQL |

CI/CDの設計は[DESIGN.md](./DESIGN.md)を参照してください。

## 設計資料

| ドキュメント | 内容 |
| ---- | ---- |
| [DESIGN.md](./DESIGN.md) | 技術選定、トレードオフ、CI/CD方針 |
| [infrastructure/ARCHITECTURE.md](./infrastructure/ARCHITECTURE.md) | AWSリソース、Security Group、IAM、DBユーザー、Secret管理 |
| [backend/docs/go-testing.md](./backend/docs/go-testing.md) | Goテスト方針とカバレッジ |
| [backend/docs/container.md](./backend/docs/container.md) | バックエンドコンテナとRDS TLS接続 |
