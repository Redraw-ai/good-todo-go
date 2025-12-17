# =============================================================================
# 環境変数 (.envから読み込み)
# =============================================================================
POSTGRES_DB_USER ?= $(shell grep POSTGRES_DB_USER .env 2>/dev/null | cut -d '=' -f2)
POSTGRES_DB_PASSWORD ?= $(shell grep POSTGRES_DB_PASSWORD .env 2>/dev/null | cut -d '=' -f2)
POSTGRES_DB_NAME ?= $(shell grep POSTGRES_DB_NAME .env 2>/dev/null | cut -d '=' -f2)
POSTGRES_DB_PORT ?= $(shell grep POSTGRES_DB_PORT .env 2>/dev/null | cut -d '=' -f2)
POSTGRES_DB_HOST ?= $(shell grep POSTGRES_DB_HOST .env 2>/dev/null | cut -d '=' -f2)

# Atlas設定変数のエクスポート
export POSTGRES_DB_USER
export POSTGRES_DB_PASSWORD
export POSTGRES_DB_NAME
export POSTGRES_DB_PORT
export POSTGRES_DB_HOST

# =============================================================================
# Atlas マイグレーション設定
# =============================================================================
ATLAS_CONFIG := file://atlas.hcl
ATLAS_ENV ?= local
MIGRATE_TO := ent://internal/ent/schema

# =============================================================================
# コード生成
# =============================================================================
generate_ent:
	go generate ./internal/ent
	go fmt ./internal/ent/...
.PHONY: generate_ent

mockgen:
	go generate ./internal/...
.PHONY: mockgen

# =============================================================================
# コード品質チェック
# =============================================================================
fmt:
	go fmt ./...
.PHONY: fmt

lint: fmt
	staticcheck ./...
.PHONY: lint

vet: fmt
	go vet ./...
.PHONY: vet

all-checks: fmt vet lint
.PHONY: all-checks

# =============================================================================
# テスト
# =============================================================================
test_unit:
	go test -v ./internal/usecase/... ./internal/pkg/... ./internal/domain/...
.PHONY: test_unit

test_integration:
	go test -v ./internal/integration_test/...
.PHONY: test_integration

test_repository:
	go test -v ./internal/infrastructure/repository/...
.PHONY: test_repository

test:
	go test -v ./...
.PHONY: test

test_coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
.PHONY: test_coverage

# =============================================================================
# データベースマイグレーション (Atlas)
# =============================================================================
migrate_diff:
	@echo "Enter migration name:"
	@read migration_name && \
	atlas migrate diff $$migration_name \
		--config "$(ATLAS_CONFIG)" \
		--env "$(ATLAS_ENV)" \
		--to "$(MIGRATE_TO)"
.PHONY: migrate_diff

migrate_apply:
	atlas migrate apply \
		--config "$(ATLAS_CONFIG)" \
		--env "$(ATLAS_ENV)"
.PHONY: migrate_apply

migrate_status:
	atlas migrate status \
		--config "$(ATLAS_CONFIG)" \
		--env "$(ATLAS_ENV)"
.PHONY: migrate_status

migrate_down:
	@if [ -z "$(n)" ]; then \
		echo "Usage: make migrate_down n=<number>"; \
		exit 1; \
	fi; \
	atlas migrate down $(n) \
		--config "$(ATLAS_CONFIG)" \
		--env "$(ATLAS_ENV)"
	@$(MAKE) migrate_status
.PHONY: migrate_down

migrate_hash:
	atlas migrate hash \
		--config "$(ATLAS_CONFIG)" \
		--env "$(ATLAS_ENV)"
.PHONY: migrate_hash

# =============================================================================
# OpenAPI Lint & Generate
# =============================================================================
oapi-lint:
	@echo "Linting OpenAPI specifications..."
	@npx @redocly/cli lint --config .redocly.yaml
.PHONY: oapi-lint

oapi-gen: oapi-lint
	@echo "Bundling and generating OpenAPI code..."
	npx @redocly/cli bundle ./openapi/openapi-public.yaml -o ./openapi-public.yaml
	npx @redocly/cli bundle ./openapi/openapi-admin.yaml -o ./openapi-admin.yaml
	oapi-codegen -config ./openapi/config-public.yaml ./openapi-public.yaml
	oapi-codegen -config ./openapi/config-admin.yaml ./openapi-admin.yaml
	@echo "OpenAPI code generation completed"
.PHONY: oapi-gen

# =============================================================================
# Docker Compose
# =============================================================================
run:
	docker compose up -d
.PHONY: run

stop:
	docker compose down
.PHONY: stop

# =============================================================================
# データベース初期化
# =============================================================================
init-db:
	@echo "Stopping and removing database containers..."
	@docker compose stop db atlas-dev-db || true
	@docker compose rm -f db atlas-dev-db || true
	@echo "Removing database volumes..."
	@docker volume rm backend_db_data backend_atlas_dev_data 2>/dev/null || true
	@echo "Starting fresh database containers..."
	@docker compose up -d db atlas-dev-db
	@echo "Waiting for database to be ready..."
	@sleep 5
	@docker compose exec -T db pg_isready -U $(POSTGRES_DB_USER) -d $(POSTGRES_DB_NAME) || (echo "Database not ready" && exit 1)
	@echo "Running migrations..."
	@$(MAKE) migrate_apply
	@echo "Database initialized successfully!"
	@$(MAKE) migrate_status
.PHONY: init-db

# =============================================================================
# 開発用
# =============================================================================
dev:
	go run ./cmd/api/main.go
.PHONY: dev
