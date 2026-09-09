.PHONY: migrate-up migrate-down

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
