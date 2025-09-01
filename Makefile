.PHONY: help test test-integration migrate-test migrate-up-all migrate-up migrate-down-all migrate-down seed-test docs
.DEFAULT_GOAL := help

TAB = $(shell printf '\t')

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	 sed -e 's/:.*## /$(TAB)/' | \
	 sort | \
	 awk -F '$(TAB)' '{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	@echo "Running unit tests..."
	@echo "--> Running api tests"
	cd apps/api && go test -v ./...
	@echo "--> Running worker tests"
	cd apps/worker && go test -v ./...

migrate-test: migrate-down-all ## Run database migrations on the test database
	@echo "Running database migrations on the test database..."
	$(MAKE) migrate-up-all

migrate-up-all: ## Apply all database migrations on the test database
	@echo "Applying all database migrations on the test database..."
	docker-compose -f docker-compose.test.yml run --rm migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable up

migrate-up: ## Apply one database migration on the test database
	@echo "Applying one database migration on the test database..."
	docker-compose -f docker-compose.test.yml run --rm migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable up $(N)

migrate-down-all: ## Rollback all database migrations on the test database.
	@echo "Rolling back all database migrations on the test database..."
	docker-compose -f docker-compose.test.yml run --rm migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable down -all

migrate-down: ## Rollback database migrations on the test database. Optional N=x to rollback x migrations.
	@echo "Rolling back database migrations on the test database..."
ifdef N
	docker-compose -f docker-compose.test.yml run --rm migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable down $(N)
else
	$(MAKE) migrate-down-all
endif

seed-test: ## Seed the test database with sample data
	@echo "Seeding the test database..."
	docker-compose -f docker-compose.test.yml run --rm -e PGPASSWORD=testpassword -v ./apps/api/testdata:/seed postgres-client -f /seed/seed.sql

test-integration: migrate-test ## Run integration tests
	@echo "Starting test database..."
	docker-compose -f docker-compose.test.yml up -d --remove-orphans postgres-test
	@echo "Running integration tests..."
	cd apps/api && go test -v -tags=integration -p 1 ./...
	@echo "Stopping test database..."
	docker-compose -f docker-compose.test.yml down

docs: ## Validate the OpenAPI specification
	@echo "Validating OpenAPI specification..."
	docker run --rm -v $(CURDIR)/web/api/v1:/spec python:3.9-slim-buster /bin/sh -c "pip install openapi-spec-validator && openapi-spec-validator /spec/oas3.yaml"
