.PHONY: migrate-up migrate-down seed api-ecr-push migration-ecr-push _ecr-push db-user db-tunnel migrate-rds-up migrate-rds-status

# ECR pushのstg用既定値。API・マイグレーションで共通利用する。
ECR_PROFILE ?= x-clone-terraform-stg
ECR_REGION ?= ap-northeast-1
ECR_REGISTRY ?= 089244387218.dkr.ecr.$(ECR_REGION).amazonaws.com

# DB管理コマンドのstg用既定値。必要な場合はmakeの引数で上書きできる。
DB_LOCAL_PORT ?= 15432
DB_LOCAL_FORWARD_ENDPOINT ?= 127.0.0.1:$(DB_LOCAL_PORT)
db-user db-tunnel: export AWS_PROFILE ?= x-clone-terraform-stg
db-user db-tunnel: export AWS_REGION ?= ap-northeast-1
db-tunnel: export NAT_INSTANCE_ID ?= i-03653ea01d0f54ffa
db-tunnel: export DB_LOCAL_PORT := $(DB_LOCAL_PORT)

# ECS単発タスクを起動し、終了コードまで確認する。SSM転送は不要。
migrate-rds-up migrate-rds-status:
	bash scripts/ecs-migrate.sh "$(patsubst migrate-rds-%,%,$@)"

# API用：Dockerfileのappステージをビルドし、apiへpushする。
api-ecr-push:
	$(MAKE) _ecr-push ECR_REPOSITORY=api DOCKER_TARGET=app

# マイグレーション用：migrationステージをビルドし、db-migratorへpushする。
migration-ecr-push:
	$(MAKE) _ecr-push ECR_REPOSITORY=db-migrator DOCKER_TARGET=migration

# 上記のapi-ecr-push, migration-ecr-pushコマンドから呼ぶ共通処理。リポジトリとビルド対象は呼び出し元で指定する。
# ログイン・ビルド・pushを順に実行し、途中で失敗したら停止する。
_ecr-push: SHELL := /bin/bash
_ecr-push:
	@set -euo pipefail; \
	test -n "$(ECR_PROFILE)" && test -n "$(ECR_REGION)" && test -n "$(ECR_REGISTRY)" || \
		{ echo 'ECR_PROFILE, ECR_REGION and ECR_REGISTRY must not be empty'; exit 1; }; \
	test -n "$(ECR_REPOSITORY)" && test -n "$(DOCKER_TARGET)" || \
		{ echo 'Use api-ecr-push or migration-ecr-push'; exit 1; }; \
	image_tag=$$(git rev-parse HEAD | cut -c1-6); \
	image_uri="$(ECR_REGISTRY)/$(ECR_REPOSITORY):$$image_tag"; \
	aws ecr get-login-password --profile "$(ECR_PROFILE)" --region "$(ECR_REGION)" \
		| docker login --username AWS --password-stdin "$(ECR_REGISTRY)"; \
	docker buildx build --target "$(DOCKER_TARGET)" --platform linux/amd64 --load -t "$$image_uri" .; \
	docker push "$$image_uri"

# SSMを使ったDBトンネルの起動
db-tunnel:
	bash scripts/db-tunnel.sh

# 別ターミナルでmake db-tunnelを起動してから実行する。
db-user: SHELL := /bin/bash
db-user:
	@set -euo pipefail; \
	rds_ca_file=$$(mktemp); \
	trap 'rm -f -- "$$rds_ca_file"' EXIT; \
	curl --fail --silent --show-error --location \
		--connect-timeout 10 --max-time 60 \
		https://truststore.pki.rds.amazonaws.com/ap-northeast-1/ap-northeast-1-bundle.pem \
		--output "$$rds_ca_file"; \
	go run ./cmd/db-user --ca-file "$$rds_ca_file" --local-forward-endpoint "$(DB_LOCAL_FORWARD_ENDPOINT)"

migrate-up:
	docker compose run --rm migrate \
		-path=/migrations \
		-database '$(MIGRATE_DATABASE_URL)' \
		up

migrate-down:
	docker compose run --rm migrate \
		-path=/migrations \
		-database '$(MIGRATE_DATABASE_URL)' \
		down 1

seed:
	docker compose exec -T postgres psql -U app -d app < seeds/001_local_users.sql
	docker compose exec -T postgres psql -U app -d app < seeds/002_local_follows.sql
	docker compose exec -T postgres psql -U app -d app < seeds/003_local_posts.sql
