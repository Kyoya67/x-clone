.PHONY: migrate-up migrate-down seed ecr-push db-user db-tunnel migrate-rds-up migrate-rds-status migration-ecr-push

MIGRATE_DATABASE_URL ?= postgres://app:app@postgres:5432/app?sslmode=disable

ECR_PROFILE ?= x-clone-terraform-stg
ECR_REGION ?= ap-northeast-1
ECR_REGISTRY ?= 089244387218.dkr.ecr.ap-northeast-1.amazonaws.com
ECR_REPOSITORY ?= backend
DOCKER_TARGET ?= app

DB_TUNNEL ?=

# ECS単発タスクを起動し、終了コードまで確認する。SSM転送は不要。
migrate-rds-up migrate-rds-status:
	bash scripts/ecs-migrate.sh "$(patsubst migrate-rds-%,%,$@)"

# マイグレーション専用イメージ。通常API用リポジトリとは分離する。
migration-ecr-push:
	$(MAKE) ecr-push ECR_REPOSITORY=backend-migration DOCKER_TARGET=migration

db-tunnel:
	bash scripts/db-tunnel.sh

# AWS_PROFILE/AWS_REGIONは実行者が指定。RDSへのネットワーク到達性が必要。
db-user: SHELL := /bin/bash
db-user:
	@set -euo pipefail; \
	rds_ca_file=$$(mktemp); \
	trap 'rm -f -- "$$rds_ca_file"' EXIT; \
	curl --fail --silent --show-error --location \
		--connect-timeout 10 --max-time 60 \
		https://truststore.pki.rds.amazonaws.com/ap-northeast-1/ap-northeast-1-bundle.pem \
		--output "$$rds_ca_file"; \
	go run ./cmd/db-user --ca-file "$$rds_ca_file" --tunnel "$(DB_TUNNEL)"

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

# ログイン・ビルド・pushを順に実行し、途中で失敗したら停止する。
ecr-push: SHELL := /bin/bash
ecr-push:
	@set -euo pipefail; \
	image_tag=$$(git rev-parse HEAD | cut -c1-6); \
	image_uri="$(ECR_REGISTRY)/$(ECR_REPOSITORY):$$image_tag"; \
	aws ecr get-login-password --profile "$(ECR_PROFILE)" --region "$(ECR_REGION)" \
		| docker login --username AWS --password-stdin "$(ECR_REGISTRY)"; \
	docker buildx build --target "$(DOCKER_TARGET)" --platform linux/amd64 --load -t "$$image_uri" .; \
	docker push "$$image_uri"
