.PHONY: migrate-up migrate-down seed

MIGRATE_DATABASE_URL ?= postgres://app:app@postgres:5432/app?sslmode=disable

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
