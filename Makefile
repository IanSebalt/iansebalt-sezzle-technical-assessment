# Single entry point for both services. Run `make` or `make help` to list targets.

GO      ?= $(shell command -v go 2>/dev/null || echo /usr/local/go/bin/go)
DOCKER  ?= $(shell command -v docker 2>/dev/null || echo $(HOME)/.docker/bin/docker)
NPM     ?= npm

BACKEND  := backend
FRONTEND := frontend

# Overridable so a busy port never blocks a local run: `make run-backend PORT=9090`
PORT ?= 9080

.DEFAULT_GOAL := help

.PHONY: help
help: ## List available targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk -F':.*?## ' '{printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: test-backend test-frontend ## Run every test suite

.PHONY: test-backend
test-backend: ## Run the Go test suite
	cd $(BACKEND) && $(GO) test ./...

.PHONY: test-frontend
test-frontend: ## Run the Vitest suite
	cd $(FRONTEND) && $(NPM) test

.PHONY: coverage
coverage: coverage-backend coverage-frontend ## Produce coverage for both layers

.PHONY: coverage-backend
coverage-backend: ## Go coverage summary per package
	cd $(BACKEND) && $(GO) test ./... -coverprofile=coverage.out
	cd $(BACKEND) && $(GO) tool cover -func=coverage.out

.PHONY: coverage-frontend
coverage-frontend: ## Vitest v8 coverage summary
	cd $(FRONTEND) && $(NPM) run test:coverage

.PHONY: run-backend
run-backend: ## Start the API (override with PORT=9090)
	cd $(BACKEND) && PORT=$(PORT) $(GO) run ./cmd/server

.PHONY: run-frontend
run-frontend: ## Start the Vite dev server
	cd $(FRONTEND) && $(NPM) run dev

.PHONY: fmt
fmt: ## Format Go sources
	cd $(BACKEND) && $(GO) fmt ./...

.PHONY: lint
lint: ## Vet Go sources and typecheck/lint the frontend
	cd $(BACKEND) && $(GO) vet ./...
	cd $(FRONTEND) && $(NPM) run typecheck && $(NPM) run lint

.PHONY: up
up: ## Build and start the full stack with Docker Compose
	$(DOCKER) compose up --build

.PHONY: down
down: ## Stop the Docker Compose stack
	$(DOCKER) compose down
