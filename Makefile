SHELL := /bin/sh
GOCACHE ?= /tmp/chronograph-go-cache
COMPOSE ?= docker compose

.DEFAULT_GOAL := help
.PHONY: help install generate fmt lint lint-go lint-frontend test test-backend test-frontend check-openapi build build-backend build-frontend check compose-config up down logs migrate prod-up prod-down

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install frontend dependencies from the lock file
	cd frontend && npm ci

generate: ## Regenerate OpenAPI server bindings, docs copy and frontend types
	cd backend && GOCACHE=$(GOCACHE) go generate ./internal/openapi
	cp backend/api/openapi.yaml docs/openapi.yaml
	cd frontend && npm run api:generate

fmt: ## Format Go and frontend sources
	cd backend && gofmt -w $$(find . -name '*.go' -not -path './internal/openapi/*')
	cd frontend && npm run format

lint: lint-go lint-frontend ## Run all linters

lint-go: ## Check Go formatting and static analysis
	@test -z "$$(gofmt -l $$(find backend -name '*.go'))" || { gofmt -l $$(find backend -name '*.go'); exit 1; }
	cd backend && GOCACHE=$(GOCACHE) go vet ./...
	cmp backend/api/openapi.yaml docs/openapi.yaml

check-openapi: ## Ensure generated frontend OpenAPI types are up to date
	cd frontend && tmp=$$(mktemp) && npx openapi-typescript ../backend/api/openapi.yaml -o $$tmp && cmp src/api-schema.d.ts $$tmp; result=$$?; rm -f $$tmp; exit $$result

lint-frontend: ## Run ESLint, Prettier and TypeScript checks
	cd frontend && npm run lint && npm run format:check && npm run typecheck && npm run audit

test: test-backend test-frontend ## Run backend and frontend test suites

test-backend: ## Run backend tests (integration tests use TEST_DATABASE_URL when set)
	cd backend && GOCACHE=$(GOCACHE) go test -race -cover ./...

test-frontend: ## Run frontend unit and UI tests
	cd frontend && npm test

build: build-backend build-frontend ## Build all applications

build-backend: ## Build backend binaries
	cd backend && GOCACHE=$(GOCACHE) go build ./cmd/server ./cmd/migrate

build-frontend: ## Build frontend assets
	cd frontend && npm run build

check: lint test check-openapi build compose-config ## Run the full local CI suite

compose-config: ## Validate development and production Compose files
	$(COMPOSE) config --quiet
	DB_PASSWORD=check API_WRITE_KEY=check $(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml config --quiet

up: ## Start the development stack
	$(COMPOSE) up --build -d

down: ## Stop the development stack
	$(COMPOSE) down

logs: ## Follow stack logs
	$(COMPOSE) logs -f

migrate: ## Apply pending database migrations
	$(COMPOSE) run --rm migrate

prod-up: ## Start the production Compose stack using .env
	$(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml up --build -d

prod-down: ## Stop the production Compose stack
	$(COMPOSE) -f docker-compose.yml -f docker-compose.prod.yml down
