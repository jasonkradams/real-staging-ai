SHELL := /bin/bash
GO_PKGS := ./...
MIGRATIONS_DIR := infra/migrations

.PHONY: tidy build run clean test test-integration migrate-up migrate-down sqlc gen lint

install-tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/segmentio/golines@latest

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$PGDATABASE?sslmode=disable" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$$PGDATABASE?sslmode=disable" down 1

sqlc:
	sqlc generate

build:
	go build -o bin/api ./apps/api
	go build -o bin/worker ./apps/worker

run:
	docker compose up --build

clean:
	rm -rf bin
	go clean -testcache

lint:
	golines -w .
	go vet $(GO_PKGS)

test:
	go test -race -v $(GO_PKGS)

# Spins pg/redis/minio/otel-collector, runs migrations, then `go test -tags=integration`
test-integration:
	docker compose -f infra/docker-compose.test.yml up -d --build
	sleep 5
	make migrate-up
	go test -v -tags=integration ./apps/api/...
	docker compose -f infra/docker-compose.test.yml down -v
