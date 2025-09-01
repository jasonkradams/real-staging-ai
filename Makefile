.PHONY: help test test-integration
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

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@echo "Not implemented yet."
