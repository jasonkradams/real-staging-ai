.PHONY: help test test-integration migrate-test migrate-up-all migrate-up migrate-down-all migrate-down seed-test docs postman sqlc-generate generate lint lint-fix
.DEFAULT_GOAL := help

TAB = $(shell printf '\t')

#########
# Linting
#########

GOLANGCI_LINT_VERSION := 2.5.0
GOLANGCI_LINT_IMAGE := golangci/golangci-lint:v$(GOLANGCI_LINT_VERSION)-alpine

HAS_GOLANGCI_LINT := $(shell command -v golangci-lint >/dev/null 2>&1 && echo 1 || echo 0)
LOCAL_LINT_MATCH := $(shell [ "$(HAS_GOLANGCI_LINT)" = "1" ] && \
	[ "$$(golangci-lint version 2>/dev/null | head -n1 | grep -oE 'version [0-9]+\.[0-9]+\.[0-9]+')" = "version $(GOLANGCI_LINT_VERSION)" ] && echo 1 || echo 0)


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

coverage: ## Generate coverage report excluding mocks
	@echo "Generating coverage report (excluding mocks)..."
	@echo "--> Running api tests with coverage"
	cd apps/api && APP_ENV=../../config go test -timeout 30s -coverprofile=coverage.tmp.out ./... && \
		cat coverage.tmp.out | grep -v _mock.go > coverage.out && \
		rm coverage.tmp.out
	@echo "--> Running worker tests with coverage"
	cd apps/worker && APP_ENV=../../config go test -timeout 60s -coverprofile=coverage.tmp.out ./... && \
		cat coverage.tmp.out | grep -v _mock.go > coverage.out && \
		rm coverage.tmp.out
	@echo ""
	@echo "Coverage reports generated:"
	@echo "  - apps/api/coverage.out"
	@echo "  - apps/worker/coverage.out"

coverage-html: coverage ## Generate HTML coverage report (excluding mocks)
	@echo "Generating HTML coverage reports..."
	cd apps/api && go tool cover -html=coverage.out -o coverage.html
	cd apps/worker && go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "HTML reports generated:"
	@echo "  - apps/api/coverage.html"
	@echo "  - apps/worker/coverage.html"
	@echo ""
	@echo "Open in browser:"
	@echo "  open apps/api/coverage.html"
	@echo "  open apps/worker/coverage.html"

coverage-summary: coverage ## Show coverage summary
	@echo ""
	@echo "=== API Coverage Summary ==="
	cd apps/api && go tool cover -func=coverage.out | tail -1
	@echo ""
	@echo "=== Worker Coverage Summary ==="
	cd apps/worker && go tool cover -func=coverage.out | tail -1

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
	docker compose -f docker-compose.yml run --rm -T migrate -path . -database postgres://postgres:postgres@postgres:5432/realstaging?sslmode=disable up

migrate-down-dev: ## Rollback database migrations on the development database
	@echo "Running database migrations on the development database..."
	docker compose -f docker-compose.yml run --rm -T migrate -path . -database postgres://postgres:postgres@postgres:5432/realstaging?sslmode=disable down -all

seed-test: ## Seed the test database with sample data
	@echo "Seeding the test database..."
	docker compose -f docker-compose.test.yml run --rm -T -e PGPASSWORD=testpassword -v ./apps/api/tests/integration/testdata:/seed postgres-client -f /seed/seed.sql

test-integration: migrate-test ## Run integration tests
	@echo "Starting test infrastructure..."
	docker compose -f docker-compose.test.yml up -d --remove-orphans postgres-test redis-test localstack
	@echo "Running integration tests..."
	cd apps/api && CONFIG_DIR=../../config APP_ENV=test PGHOST=localhost PGPORT=5433 PGUSER=testuser PGPASSWORD=testpassword PGDATABASE=testdb PGSSLMODE=disable REDIS_ADDR=localhost:6379 go test -tags=integration -p 1 ./...
	cd apps/worker && CONFIG_DIR=../../config APP_ENV=test PGHOST=localhost PGPORT=5433 PGUSER=testuser PGPASSWORD=testpassword PGDATABASE=testdb PGSSLMODE=disable REDIS_ADDR=localhost:6379 go test -tags=integration -p 1 ./...
	@echo "Stopping test infrastructure..."
	docker compose -f docker-compose.test.yml down

docs-validate: ## Validate the OpenAPI specification
	@echo "Validating OpenAPI specification..."
	docker run --rm -v $(CURDIR)/apps/api/web/api/v1:/spec python:3.13-slim /bin/sh -c "pip install openapi-spec-validator && openapi-spec-validator /spec/oas3.yaml"

docs-install: ## Install MkDocs and dependencies for documentation site
	@echo "Installing documentation dependencies..."
	cd apps/docs && pip3 install -r requirements.txt

docs-serve: ## Serve documentation site locally at localhost:8000
	@echo "Starting documentation server..."
	cd apps/docs && mkdocs serve -a 0.0.0.0:8000

docs-build: ## Build static documentation site
	@echo "Building documentation site..."
	cd apps/docs && mkdocs build

docs-up: ## Run docs in Docker container
	@echo "Starting docs container..."
	docker compose up -d docs

docs-down: ## Stop docs container
	@echo "Stopping docs container..."
	docker compose stop docs

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

lint: ## Run golangci-lint locally or in Docker if not available or mismatched
	@echo "Running golangci-lint ($(GOLANGCI_LINT_VERSION))..."
	@if [ "$(LOCAL_LINT_MATCH)" = "1" ]; then \
		echo "--> Using local golangci-lint"; \
		echo "--> Running api lint"; \
		cd apps/api && golangci-lint run; \
		echo "--> Running worker lint"; \
		cd ../../apps/worker && golangci-lint run; \
	else \
		echo "--> Using Dockerized golangci-lint"; \
		echo "--> Running api lint"; \
		cd apps/api && docker run --rm -v $(CURDIR):/app -w /app/apps/api $(GOLANGCI_LINT_IMAGE) golangci-lint run; \
		echo "--> Running worker lint"; \
		cd ../../apps/worker && docker run --rm -v $(CURDIR):/app -w /app/apps/worker $(GOLANGCI_LINT_IMAGE) golangci-lint run; \
	fi
	@echo "--> Linting web server"
	@echo "--> Running web lint"
	cd apps/web && npm run lint

lint-fix: ## Run golangci-lint with --fix locally or in Docker if not available or mismatched
	@echo "Running golangci-lint with --fix ($(GOLANGCI_LINT_VERSION))..."
	@if [ "$(LOCAL_LINT_MATCH)" = "1" ]; then \
		echo "--> Using local golangci-lint"; \
		echo "--> Running api lint with --fix"; \
		cd apps/api && golangci-lint run --fix; \
		echo "--> Running worker lint with --fix"; \
		cd ../../apps/worker && golangci-lint run --fix; \
	else \
		echo "--> Using Dockerized golangci-lint"; \
		echo "--> Running api lint with --fix"; \
		cd apps/api && docker run --rm -v $(CURDIR):/app -w /app/apps/api $(GOLANGCI_LINT_IMAGE) golangci-lint run --fix; \
		echo "--> Running worker lint with --fix"; \
		cd ../../apps/worker && docker run --rm -v $(CURDIR):/app -w /app/apps/worker $(GOLANGCI_LINT_IMAGE) golangci-lint run --fix; \
	fi
	@echo "--> Linting and fixing web server"
	@echo "--> Running web lint with --fix"
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
	cd apps/api && go clean -cache -testcache -modcache &
	cd apps/worker && go clean -cache -testcache -modcache &
	rm -rf apps/api/bin apps/api/pkg apps/worker/bin apps/worker/pkg &
	rm -rf apps/web/.next &
	rm -rf apps/docs/site &
	find . -type f -name "cover*.out" -exec rm -rf {} + &
	find . -type f -name "cover*.html" -exec rm -rf {} + &
	find . -type f -name .localstack -exec rm -rf {} + &
	find . -type d -name coverage -exec rm -rf {} + &

clean-mock: ## Remove all mock files
	find . -type f -name "*_mock.go" -exec rm -rf {} + &

clean-all: clean ## Remove all mock files and clean databases/storage
	$(MAKE) migrate-down-dev
	@echo "Cleaning MinIO buckets..."
	docker compose exec minio sh -c "mc alias set local http://localhost:9000 minioadmin minioadmin && mc rm --recursive --force local/real-staging/uploads/ || true"
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
	docker compose exec api /bin/sh -c "/app/reconcile --dry-run=$(or $(DRY_RUN),true) --batch-size=$(or $(BATCH_SIZE),100) --concurrency=$(or $(CONCURRENCY),5)"
