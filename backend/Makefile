.PHONY: migrate-up migrate-down seed ecr-push

MIGRATE_DATABASE_URL ?= postgres://app:app@postgres:5432/app?sslmode=disable

ECR_PROFILE ?= x-clone-terraform-stg
ECR_REGION ?= ap-northeast-1
ECR_REGISTRY ?= 089244387218.dkr.ecr.ap-northeast-1.amazonaws.com
ECR_REPOSITORY ?= backend

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
	docker buildx build --platform linux/amd64 --load -t "$$image_uri" .; \
	docker push "$$image_uri"
