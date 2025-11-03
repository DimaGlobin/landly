.PHONY: help dev run down migrate migrate-down migration seed test test-backend test-frontend build clean

ROOT_DIR := $(shell pwd)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start all services in development mode
	docker-compose -f deploy/docker/docker-compose.yml up --build -d
	@echo "Services started:"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Backend API: http://localhost:8080"
	@echo "  - MinIO Console: http://localhost:9001"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"
	@echo "--- Tailing backend logs (press Ctrl+C to stop) ---"
	docker-compose -f deploy/docker/docker-compose.yml logs -f backend

run: ## Run tests in container, then start development services
	COMPOSE_PROFILES=test docker-compose -f deploy/docker/docker-compose.yml run --rm --no-deps --entrypoint "" backend-builder sh -c "go test ./..."
	COMPOSE_PROFILES=test docker-compose -f deploy/docker/docker-compose.yml run --rm --no-deps frontend-tester
	docker-compose -f deploy/docker/docker-compose.yml up --build -d
	@echo "Services started:"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Backend API: http://localhost:8080"
	@echo "  - MinIO Console: http://localhost:9001"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Redis: localhost:6379"

down: ## Stop all services
	docker-compose -f deploy/docker/docker-compose.yml down

logs: ## Show logs from all services
	docker-compose -f deploy/docker/docker-compose.yml logs -f

migrate: ## Run database migrations
	docker-compose -f deploy/docker/docker-compose.yml exec backend /app/migrate up

migrate-down: ## Rollback last migration
	docker-compose -f deploy/docker/docker-compose.yml exec backend /app/migrate down

migration: ## Create new migration (use: make migration name=add_users_table)
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migration name=add_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@cd apps/backend && goose -dir migrations create $(name) sql

seed: ## Load seed data into database
	docker-compose -f deploy/docker/docker-compose.yml exec backend go run cmd/seed/main.go

test: test-backend test-frontend ## Run all tests
	@$(MAKE) clean

test-backend: ## Run backend tests
	@set -euo pipefail; \
	trap 'cd $(ROOT_DIR) && $(MAKE) clean' EXIT; \
	cd $(ROOT_DIR)/apps/backend && go test -v -race -cover ./...

test-frontend: ## Run frontend tests
	@set -euo pipefail; \
	trap 'cd $(ROOT_DIR) && $(MAKE) clean' EXIT; \
	cd $(ROOT_DIR)/apps/frontend && npm test

test-integration: ## Run integration tests with docker-compose (verbose)
	@set -euo pipefail; \
	trap 'cd $(ROOT_DIR) && $(MAKE) clean' EXIT; \
	echo "Starting integration test infrastructure..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.test.yml up --build -d; \
	echo ""; \
	echo "Current containers:"; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.test.yml ps; \
	echo ""; \
	echo "Waiting for services to be ready..."; \
	sleep 5; \
	echo ""; \
	echo "Running backend integration test suite (verbose, no caching)..."; \
	cd $(ROOT_DIR)/apps/backend && \
		TEST_POSTGRES_DSN="postgres://landly:landly@localhost:5433/landly_test?sslmode=disable" \
		TEST_REDIS_ADDR="localhost:6380" \
		TEST_S3_ENDPOINT="localhost:9002" \
		TEST_S3_ACCESS_KEY="minioadmin" \
		TEST_S3_SECRET_KEY="minioadmin" \
		TEST_S3_USE_SSL="false" \
		go test -v -count=1 -tags=integration ./...; \
	echo ""; \
	echo "Stopping integration test infrastructure..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.test.yml down

test-e2e: ## Run end-to-end tests
	@set -euo pipefail; \
	trap 'cd $(ROOT_DIR) && $(MAKE) clean' EXIT; \
	echo "Starting full stack for e2e tests..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.yml up -d; \
	echo "Waiting for services to be ready..."; \
	sleep 10; \
	echo "Running e2e tests..."; \
	cd $(ROOT_DIR)/apps/frontend && npm run test:e2e; \
	echo "Stopping services..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.yml down

build: ## Build all services
	docker-compose -f deploy/docker/docker-compose.yml build

build-backend: ## Build backend only
	cd apps/backend && go build -o bin/api cmd/api/main.go
	cd apps/backend && go build -o bin/worker cmd/worker/main.go

build-frontend: ## Build frontend only
	cd apps/frontend && npm run build

clean: ## Clean containers, volumes, and build artifacts
	@set -euo pipefail; \
	echo "Stopping and removing runtime stack (if running)..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.yml down --remove-orphans || true; \
	echo "Stopping and removing test stack (if running)..."; \
	docker-compose -f $(ROOT_DIR)/deploy/docker/docker-compose.test.yml down --volumes --remove-orphans || true; \
	echo "Removing dangling volumes for landly project..."; \
	docker volume ls --filter name=landly -q | xargs -r docker volume rm; \
	echo "Removing backend build artifacts..."; \
	rm -rf apps/backend/bin; \
	echo "Removing frontend build artifacts..."; \
	rm -rf apps/frontend/.next apps/frontend/out; \
	echo "Cleanup complete."

lint: ## Run linters
	cd apps/backend && golangci-lint run
	cd apps/frontend && npm run lint

fmt: ## Format code
	cd apps/backend && go fmt ./...
	cd apps/frontend && npm run format

setup-local: ## Setup local development environment
	@echo "Setting up local development environment..."
	cp apps/backend/.env.example apps/backend/.env || true
	cp apps/frontend/.env.example apps/frontend/.env || true
	@echo "Installing backend dependencies..."
	cd apps/backend && go mod download
	@echo "Installing frontend dependencies..."
	cd apps/frontend && npm install
	@echo "Setup complete! Run 'make dev' to start services."

reset-db: ## Reset database (WARNING: destroys all data)
	docker-compose -f deploy/docker/docker-compose.yml down -v
	docker-compose -f deploy/docker/docker-compose.yml up -d postgres
	sleep 3
	make migrate

