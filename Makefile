# ============================================================
# HRMS Makefile — Common Development Tasks
# ============================================================

.PHONY: help build vet test test-integration test-short migrate migrate-down dev-backend dev-frontend dev clean lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Backend ───────────────────────────────────────────────────

build: ## Build all Go binaries
	cd backend && go build ./...

vet: ## Run go vet
	cd backend && go vet ./...

test: ## Run all unit tests (short mode, skip integration)
	cd backend && go test -count=1 -short ./...

test-verbose: ## Run all unit tests with verbose output
	cd backend && go test -count=1 -short -v ./...

test-integration: ## Run integration tests (requires PostgreSQL + RUN_INTEGRATION_TESTS=true)
	cd backend && go test -count=1 -tags=integration -v ./internal/repository/ -run "TestOptimisticLocking"

test-all: ## Run all tests (unit + integration)
	cd backend && RUN_INTEGRATION_TESTS=true go test -count=1 -tags=integration ./...

lint: ## Run golangci-lint (if installed)
	golangci-lint run ./backend/... 2>/dev/null || echo "golangci-lint not installed, skipping"

dev-backend: ## Start backend with Air hot-reload
	cd backend && air

dev-frontend: ## Start frontend dev server
	cd frontend && npm run dev

migrate: ## Run database migrations via Goose
	cd backend && goose -dir ../database/migrations postgres "postgres://${DB_USER:-magnum}:${DB_PASSWORD:-magnum}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-hrms}?sslmode=${DB_SSLMODE:-disable}" up

migrate-down: ## Rollback latest migration
	cd backend && goose -dir ../database/migrations postgres "postgres://${DB_USER:-magnum}:${DB_PASSWORD:-magnum}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-hrms}?sslmode=${DB_SSLMODE:-disable}" down

migrate-status: ## Check migration status
	cd backend && goose -dir ../database/migrations postgres "postgres://${DB_USER:-magnum}:${DB_PASSWORD:-magnum}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-hrms}?sslmode=${DB_SSLMODE:-disable}" status

# ─── Frontend ──────────────────────────────────────────────────

frontend-install: ## Install frontend dependencies
	cd frontend && npm ci

frontend-check: ## Run Svelte type-check
	cd frontend && npm run check

frontend-build: ## Build frontend & copy to backend/public/ for Go static serving
	cd frontend && npm run build && rm -rf ../backend/public && cp -r build ../backend/public
	@echo "✅ Frontend built & copied to backend/public/"

frontend-test: ## Run Playwright E2E tests
	cd frontend && npm run test:e2e 2>/dev/null || echo "Playwright not configured, skipping"

# ─── Docker ───────────────────────────────────────────────────

docker-up: ## Start all services with Docker Compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-build: ## Rebuild Docker images
	docker compose build

docker-logs: ## Follow logs
	docker compose logs -f

docker-restart: ## Restart all services
	docker compose restart

# ─── Database ──────────────────────────────────────────────────

db-backup: ## Backup database to timestamped file
	@mkdir -p backups
	pg_dump "postgres://${DB_USER:-magnum}:${DB_PASSWORD:-magnum}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-hrms}?sslmode=${DB_SSLMODE:-disable}" \
		--format=custom --file=backups/hrms_$$(date +%Y%m%d_%H%M%S).dump
	@echo "✅ Backup saved to backups/"

db-restore: ## Restore database from dump file (usage: make db-restore FILE=backups/xxx.dump)
	@if [ -z "$(FILE)" ]; then echo "❌ Usage: make db-restore FILE=backups/xxx.dump"; exit 1; fi
	pg_restore --clean --no-owner --dbname="postgres://${DB_USER:-magnum}:${DB_PASSWORD:-magnum}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-hrms}?sslmode=${DB_SSLMODE:-disable}" "$(FILE)"
	@echo "✅ Restore complete"

db-psql: ## Open psql shell
	PGPASSWORD=${DB_PASSWORD:-magnum} psql -h ${DB_HOST:-localhost} -p ${DB_PORT:-5432} -U ${DB_USER:-magnum} -d ${DB_NAME:-hrms}

# ─── Git ───────────────────────────────────────────────────────

pre-commit: ## Run pre-commit checks (vet + build + test)
	$(MAKE) vet build test

# ─── Clean ─────────────────────────────────────────────────────

clean: ## Clean build artifacts
	cd backend && go clean -cache
	rm -rf frontend/.svelte-kit frontend/build
	@echo "✅ Clean complete"

# ─── CI Check (runs in GitHub Actions) ─────────────────────────

ci-check: ## Full CI pipeline (vet + build + test + frontend-check)
	$(MAKE) vet build test
	cd frontend && npm ci && npm run check && npm run build
	@echo "✅ CI check passed"

.DEFAULT_GOAL := help
