.PHONY: help test test-integration migrate-test migrate-up-all migrate-up migrate-down-all migrate-down seed-test docs postman sqlc-generate generate lint lint-fix
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
	cd apps/api && APP_ENV=../../config go test -timeout 30s ./...
	@echo "--> Running worker tests"
	cd apps/worker && APP_ENV=../../config go test -timeout 60s -v ./internal/repository ./internal/events ./...
	@echo "--> Running web tests"
	cd apps/web && npm run test

test-cover: ## Run unit tests with coverage
	@echo "Running unit tests with coverage..."
	@echo "--> Running api tests"
	cd apps/api && APP_ENV=../../config go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "--> Running worker tests"
	cd apps/worker && APP_ENV=../../config go test -coverprofile=coverage.out ./internal/repository ./internal/events ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "--> Running web tests"
	cd apps/web && npm run test:coverage

migrate-test: migrate-down-all ## Run database migrations on the test database
	@echo "Running database migrations on the test database..."
	$(MAKE) migrate-up-all

migrate-up-all: ## Apply all database migrations on the test database
	@echo "Applying all database migrations on the test database..."
	docker compose -f docker-compose.test.yml run --remove-orphans --rm -T migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable up

migrate-up: ## Apply one database migration on the test database
	@echo "Applying one database migration on the test database..."
	docker compose -f docker-compose.test.yml run --remove-orphans --rm -T migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable up $(N)

migrate-down-all: ## Rollback all database migrations on the test database.
	@echo "Rolling back all database migrations on the test database..."
	docker compose -f docker-compose.test.yml run --remove-orphans --rm -T migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable down -all

migrate-down: ## Rollback database migrations on the test database. Optional N=x to rollback x migrations.
	@echo "Rolling back database migrations on the test database..."
ifdef N
	docker compose -f docker-compose.test.yml run --remove-orphans --rm -T migrate -path . -database postgres://testuser:testpassword@postgres-test:5432/testdb?sslmode=disable down $(N)
else
	$(MAKE) migrate-down-all
endif

migrate: ## Run database migrations on the development database
	@echo "Running database migrations on the development database..."
	docker-compose -f docker-compose.yml run --rm -T migrate -path . -database postgres://postgres:postgres@postgres:5432/virtualstaging?sslmode=disable up

migrate-down-dev: ## Rollback database migrations on the development database
	@echo "Running database migrations on the development database..."
	docker-compose -f docker-compose.yml run --rm -T migrate -path . -database postgres://postgres:postgres@postgres:5432/virtualstaging?sslmode=disable down -all

seed-test: ## Seed the test database with sample data
	@echo "Seeding the test database..."
	docker compose -f docker-compose.test.yml run --rm -T -e PGPASSWORD=testpassword -v ./apps/api/tests/integration/testdata:/seed postgres-client -f /seed/seed.sql

test-integration: migrate-test ## Run integration tests
	@echo "Starting test infrastructure..."
	docker-compose -f docker-compose.test.yml up -d --remove-orphans postgres-test redis-test localstack
	@echo "Running integration tests..."
	cd apps/api && CONFIG_DIR=../../config APP_ENV=test PGHOST=localhost PGPORT=5433 PGUSER=testuser PGPASSWORD=testpassword PGDATABASE=testdb PGSSLMODE=disable REDIS_ADDR=localhost:6379 go test -tags=integration -p 1 ./...
	cd apps/worker && CONFIG_DIR=../../config APP_ENV=test PGHOST=localhost PGPORT=5433 PGUSER=testuser PGPASSWORD=testpassword PGDATABASE=testdb PGSSLMODE=disable REDIS_ADDR=localhost:6379 go test -tags=integration -p 1 ./...
	@echo "Stopping test infrastructure..."
	docker compose -f docker-compose.test.yml down

docs: ## Validate the OpenAPI specification
	@echo "Validating OpenAPI specification..."
	docker run --rm -v $(CURDIR)/apps/api/web/api/v1:/spec python:3.13-slim /bin/sh -c "pip install openapi-spec-validator && openapi-spec-validator /spec/oas3.yaml"

postman: ## Generate a Postman collection from the OpenAPI specification
	@echo "Generating Postman collection..."
	@npx openapi-to-postmanv2 -s apps/api/web/api/v1/oas3.yaml -o postman_collection.json

sqlc-generate: ## Generate Go code from SQL queries using sqlc
	@echo "Generating sqlc code..."
	cd apps/api && ~/go/bin/sqlc generate

generate: ## Generate all code (mocks, sqlc, etc.)
	$(MAKE) clean-mock
	$(MAKE) generate-api
	$(MAKE) generate-worker
	$(MAKE) tidy

generate-api:
	@echo "Generating all code..."
	$(MAKE) sqlc-generate
	@echo "Generating mocks..."
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		cd apps/api && go install github.com/matryer/moq@v0.5.3; \
	fi
	cd apps/api && go generate ./...

generate-worker:
	@echo "Generating all code..."
	$(MAKE) sqlc-generate
	@echo "Generating mocks..."
	@if ! command -v mockgen >/dev/null 2>&1; then \
		echo "Installing mockgen..."; \
		cd apps/worker && go install github.com/matryer/moq@v0.5.3; \
	fi
	cd apps/worker && go generate ./...

lint: ## Run golangci-lint on all Go modules
	@echo "Running golangci-lint..."
	@echo "--> Linting api module"
	cd apps/api && docker run --rm -v $(CURDIR):/app -w /app/apps/api golangci/golangci-lint:v2.4.0-alpine golangci-lint run
	@echo "--> Linting worker module"
	cd apps/worker && docker run --rm -v $(CURDIR):/app -w /app/apps/worker golangci/golangci-lint:v2.4.0-alpine golangci-lint run
	@echo "--> Linting web server"
	cd apps/web && npm run lint

lint-fix: ## Run golangci-lint with --fix on all Go modules
	@echo "Running golangci-lint with --fix..."
	@echo "--> Linting and fixing api module"
	cd apps/api && docker run --rm -v $(CURDIR):/app -w /app/apps/api golangci/golangci-lint:v2.4.0-alpine golangci-lint run --fix
	@echo "--> Linting and fixing worker module"
	cd apps/worker && docker run --rm -v $(CURDIR):/app -w /app/apps/worker golangci/golangci-lint:v2.4.0-alpine golangci-lint run --fix
	@echo "--> Linting and fixing web server"
	cd apps/web && npm run lint:fix

up: migrate ## Run the api server
	@echo Starting Application...
	docker compose -f docker-compose.yml up --build -d --remove-orphans api worker
	$(MAKE) up-web

down: ## Stop the api server
	@echo Stopping Application...
	docker compose -f docker-compose.yml stop

up-web: ## Run the web server
	@echo Running web server...
	cd apps/web && npm run dev

down-web: ## Stop the web server
	@echo Stopping web server...
	docker compose -f docker-compose.yml stop web

clean: ## Remove unused and unnecessary files
	@echo "Removing unused and unnecessary files..."
	cd apps/api && go clean -cache -testcache -modcache
	cd apps/worker && go clean -cache -testcache -modcache
	rm -rf apps/api/bin apps/api/pkg apps/worker/bin apps/worker/pkg &
	rm -rf apps/web/.next &
	find . -type f -name "cover*.out" -exec rm -rf {} + &
	find . -type f -name "cover*.html" -exec rm -rf {} + &
	find . -type f -name .localstack -exec rm -rf {} + &
	find . -type d -name coverage -exec rm -rf {} + &

clean-mock: ## Remove all mock files
	find . -type f -name "*_mock.go" -exec rm -rf {} + &

clean-all: clean ## Remove all mock files and clean databases/storage
	$(MAKE) migrate-down-dev
	@echo "Cleaning MinIO buckets..."
	docker compose exec minio sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin && mc rm --recursive --force local/virtual-staging/uploads/ || true"
	@echo "Removing node_modules..."
	find . -type d -name node_modules -exec rm -rf {} + &
	$(MAKE) tidy

token: ## Generate a Auth0 Token
	@go run -C apps/api ./cmd/token/main.go | jq -r .access_token

tidy: ## Run go mod tidy for each app
	cd apps/api && go mod tidy
	cd apps/worker && go mod tidy

reconcile-images: ## Run storage reconciliation CLI (use DRY_RUN=1 for dry-run)
	@echo "Running storage reconciliation..."
	docker compose exec api /bin/sh -c "/reconcile --dry-run=$(or $(DRY_RUN),true) --batch-size=$(or $(BATCH_SIZE),100) --concurrency=$(or $(CONCURRENCY),5)"
