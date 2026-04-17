.PHONY: help up down logs ps proto-gen proto-openapi-gen proto-lint proto-format sqlc-gen docs-gen docs-gen-api gen init-deps init-go-deps init-python-deps uv-install uv-sync uv-add uv-add-dev uv-run build-all build build-gateway build-product build-order build-user build-payment build-inventory build-delivery load-test benchmark-cache build-python test test-coverage test-integration test-python lint lint-python format-python db-migrate db-seed db-rollback db-rollback-all db-reset docker-build docker-push k8s-apply k8s-delete k8s-logs clean clean-all proto-watch install-git-hooks

DATABASE_URL ?= postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable
PSQL = docker exec -i shinkansen-postgres psql -U shinkansen -d shinkansen

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Infrastructure ---
up: gen ## Start Infrastructure (Docker Compose)
	docker compose up -d

down: ## Stop Infrastructure
	docker compose down

logs: ## View Infrastructure Logs
	docker compose logs -f

ps: ## Show running containers
	docker compose ps

# --- Generation ---
PROTOC_DIR := $(HOME)/.local
PROTOC_INCLUDE := $(PROTOC_DIR)/include
PROTOC_BIN := $(PROTOC_DIR)/bin

proto-gen: ## Generate gRPC code from protobufs
	@echo "🔄 Generating Go protobuf code..."
	@if ! command -v protoc >/dev/null 2>&1 && [ ! -f $(PROTOC_BIN)/protoc ]; then \
		echo "Installing protoc..." && \
		mkdir -p $(PROTOC_DIR)/bin $(PROTOC_INCLUDE) && \
		curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/v30.2/protoc-30.2-linux-x86_64.zip -o /tmp/protoc.zip && \
		unzip -oq /tmp/protoc.zip -d $(PROTOC_DIR) 'bin/*' 'include/*' && \
		rm /tmp/protoc.zip; \
	fi
	@mkdir -p $(PROTOC_INCLUDE)/google/api
	@if [ ! -f $(PROTOC_INCLUDE)/google/api/annotations.proto ]; then \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o $(PROTOC_INCLUDE)/google/api/annotations.proto && \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o $(PROTOC_INCLUDE)/google/api/http.proto; \
	fi
	@mkdir -p gen/proto/go
	@if [ ! -f gen/proto/go/go.mod ]; then \
		echo 'module github.com/afasari/shinkansen-commerce/gen/proto/go' > gen/proto/go/go.mod && \
		echo 'go 1.21' >> gen/proto/go/go.mod; \
	fi
	protoc --proto_path=$(PROTOC_INCLUDE) --proto_path=proto --go_out=gen/proto/go --go_opt=paths=source_relative --go-grpc_out=gen/proto/go --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false proto/**/*.proto || true
	@echo "✅ Go protobuf code generated"

proto-gen-rust: ## Generate Rust protobuf code
	@echo "🦀 Generating Rust protobuf code..."
	@mkdir -p gen/proto/rust
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "Installing buf..." && \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	fi
	cd gen/proto/rust && cargo build --release
	@echo "✅ Rust protobuf code generated"

proto-openapi-gen: ## Generate OpenAPI docs from protobufs
	@echo "📝 Generating OpenAPI docs..."
	@if ! command -v protoc >/dev/null 2>&1 && [ ! -f $(PROTOC_BIN)/protoc ]; then \
		echo "Installing protoc..." && \
		mkdir -p $(PROTOC_DIR)/bin $(PROTOC_INCLUDE) && \
		curl -sSL https://github.com/protocolbuffers/protobuf/releases/download/v30.2/protoc-30.2-linux-x86_64.zip -o /tmp/protoc.zip && \
		unzip -oq /tmp/protoc.zip -d $(PROTOC_DIR) 'bin/*' 'include/*' && \
		rm /tmp/protoc.zip; \
	fi
	@mkdir -p $(PROTOC_INCLUDE)/google/api
	@if [ ! -f $(PROTOC_INCLUDE)/google/api/annotations.proto ]; then \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o $(PROTOC_INCLUDE)/google/api/annotations.proto && \
		curl -sSL https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o $(PROTOC_INCLUDE)/google/api/http.proto; \
	fi
	@mkdir -p services/gateway/docs/api
	protoc --proto_path=$(PROTOC_INCLUDE) --proto_path=proto --openapiv2_out=services/gateway/docs/api --openapiv2_opt=json_names_for_fields=false,allow_merge=true,merge_file_name=swagger,output_format=yaml proto/**/*.proto || true
	@echo "✅ OpenAPI docs generated"

proto-lint: ## Lint protobuf files
	buf lint proto

proto-format: ## Format protobuf files
	buf format -w proto

proto-watch: ## Watch proto files and auto-regenerate code (EXPERIMENTAL)
	@echo "👀 Proto file watcher (experimental)..."
	@bash scripts/automation/watch-proto.sh || (echo "⚠️  Watch mode not available - use manual workflow" && echo "   Edit proto files, then run: make gen")

install-git-hooks: ## Install git hooks for auto code generation
	@echo "🔧 Installing git hooks..."
	@cp scripts/automation/pre-commit.sh .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"
	@echo "   Code will be auto-generated when proto files are committed"
	@echo "   (Hook installed as direct file, not symlink for reliability)"

sqlc-gen: ## Generate SQL code for Go services
	@echo "🔄 Generating SQL code for Product Service..."
	cd services/product-service && sqlc generate
	@echo "✅ Product Service SQL generated"
	@echo "🔄 Generating SQL code for Order Service..."
	cd services/order-service && sqlc generate
	@echo "✅ Order Service SQL generated"
	# @echo "🔄 Generating SQL code for Payment Service..."
	# cd services/payment-service && sqlc generate
	# @echo "✅ Payment Service SQL generated"
	# @echo "🔄 Generating SQL code for User Service..."
	# cd services/user-service && sqlc generate
	# @echo "✅ User Service SQL generated"
	# @echo "🔄 Generating SQL code for Delivery Service..."
	# cd services/delivery-service && sqlc generate
	# @echo "✅ Delivery Service SQL generated"

# Documentation
docs-gen: ## Generate Rust documentation
		@echo "📖️  Generating Rust documentation..."
		@cd services/inventory-service && cargo doc --open --no-deps

docs-gen-api: ## Generate API documentation from proto files
		@echo "📝 Generating API documentation from proto files..."
		@bash scripts/automation/generate-api-docs.sh

gen: proto-gen proto-gen-rust proto-openapi-gen sqlc-gen docs-gen docs-gen-api ## Generate all code (protobuf + sqlc + openapi + docs)

# --- Dependencies ---
init-deps: ## Download all dependencies
	@echo "📦 Installing Go dependencies..."
	cd services/gateway && go mod tidy && go fmt ./...
	cd services/product-service && go mod tidy && go fmt ./..
	cd services/order-service && go mod tidy && go fmt ./..
	cd services/payment-service && go mod tidy && go fmt ./..
	cd services/user-service && go mod tidy && go fmt ./..
	cd services/delivery-service && go mod tidy && go fmt ./..
	@echo "📝 Formatting inventory-service..."
	@cd services/inventory-service && cargo fmt --all
	@echo "✅ All dependencies installed"

uv-install: ## Install uv package manager
	@echo "📦 Installing uv..."
	@command -v uv --version || (curl -LsSf https://astral.sh/uv/install.sh | sh && echo "✅ uv installed")

uv-sync: uv-install ## Sync Python dependencies using uv
	@echo "📦 Syncing Python dependencies with uv..."
	cd services/analytics-worker && uv sync

uv-add: uv-install ## Add Python package (usage: make uv-add PACKAGE=<name>)
	@if [ -z "$(PACKAGE)" ]; then echo "Usage: make uv-add PACKAGE=<package-name>"; exit 1; fi
	cd services/analytics-worker && uv add $(PACKAGE)

uv-add-dev: uv-install ## Add Python dev package (usage: make uv-add-dev PACKAGE=<name>)
	@if [ -z "$(PACKAGE)" ]; then echo "Usage: make uv-add-dev PACKAGE=<package-name>"; exit 1; fi
	cd services/analytics-worker && uv add --dev $(PACKAGE)

uv-run: uv-install ## Run Python command (usage: make uv-run CMD=<command>)
	@if [ -z "$(CMD)" ]; then echo "Usage: make uv-run CMD=<command>"; exit 1; fi
	cd services/analytics-worker && uv run $(CMD)

# --- Build ---
build-all: ## Build all Go services
	@echo "🔨 Building all services..."
	@mkdir -p bin
	@echo "🔨 Building Gateway..."
	cd services/gateway && go build -o ../../bin/gateway ./cmd/gateway
	@echo "🔨 Building Product Service..."
	cd services/product-service && go build -o ../../bin/product-service ./cmd/product-service
	@echo "🔨 Building Order Service..."
	cd services/order-service && go build -o ../../bin/order-service ./cmd/order-service
	@echo "🔨 Building Payment Service..."
	cd services/payment-service && go build -o ../../bin/payment-service ./cmd/payment-service
	@echo "🔨 Building User Service..."
	cd services/user-service && go build -o ../../bin/user-service ./cmd/user-service
	@echo "🔨 Building Delivery Service..."
	cd services/delivery-service && go build -o ../../bin/delivery-service ./cmd/delivery-service
	@echo "🔨 Building Inventory Service..."
	@echo "✅ All services built to bin/"

build: build-all ## Build all services (alias)

build-gateway: ## Build Gateway only
	@mkdir -p bin
	cd services/gateway && go build -o ../../bin/gateway ./cmd/gateway
	@echo "✅ Gateway built to bin/gateway"

build-product: ## Build Product Service only
	@mkdir -p bin
	cd services/product-service && go build -o ../../bin/product-service ./cmd/product-service
	@echo "✅ Product Service built to bin/product-service"

build-order: ## Build Order Service only
	@mkdir -p bin
	cd services/order-service && go build -o ../../bin/order-service ./cmd/order-service
	@echo "✅ Order Service built to bin/order-service"

build-user: ## Build User Service only
	@mkdir -p bin
	cd services/user-service && go build -o ../../bin/user-service ./cmd/user-service
	@echo "✅ User Service built to bin/user-service"

build-payment: ## Build Payment Service only
	@mkdir -p bin
	cd services/payment-service && go build -o ../../bin/payment-service ./cmd/payment-service
	@echo "✅ Payment Service built to bin/payment-service"

build-inventory: ## Build inventory-service (Rust release)
	@echo "🏗️  Building inventory-service (Rust release)..."
	@cd services/inventory-service && cargo build --release
	@mkdir -p bin
	@cp services/inventory-service/target/release/shinkansen-inventory-service bin/inventory-service

build-delivery: ## Build Delivery Service only
	@mkdir -p bin
	cd services/delivery-service && go build -o ../../bin/delivery-service ./cmd/delivery-service
	@echo "✅ Delivery Service built to bin/delivery-service"

load-test: ## Run load test for product service (10K concurrent reads)
	@echo "🚀 Running load test (10K concurrent reads)..."
	cd services/product-service && go run cmd/load-test/main.go

benchmark-cache: ## Run cache performance benchmark
	@echo "📊 Running cache benchmark..."
	cd services/product-service && go run cmd/load-test/main.go benchmark

build-python: uv-sync ## Build/Package Python analytics worker
	@echo "🐍 Building Python analytics worker..."
	cd services/analytics-worker && uv pip compile pyproject.toml -o requirements.txt

# --- Test ---
test: ## Run all tests
	@echo "🧪 Running Gateway tests..."
	cd services/gateway && go test ./...
	@echo "🧪 Running Product Service tests..."
	cd services/product-service && go test ./...
	@echo "🧪 Running Order Service tests..."
	cd services/order-service && go test ./...
	@echo "🧪 Running Payment Service tests..."
	cd services/payment-service && go test ./...
	@echo "🧪 Running User Service tests..."
	cd services/user-service && go test ./...
	@echo "🧪 Running Delivery Service tests..."
	cd services/delivery-service && go test ./...
	@echo "🧪 Running inventory-service tests..."
	@cd services/inventory-service &&	cargo test -- --test-threads=1

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	cd services/gateway && go test -coverprofile=coverage.out ./...
	cd services/product-service && go test -coverprofile=coverage.out ./...
	cd services/order-service && go test -coverprofile=coverage.out ./...

test-integration: ## Run integration tests (requires infrastructure)
	@echo "🔧 Running integration tests..."
	@./scripts/run-integration-tests.sh

test-python: uv-install ## Run Python tests
	@echo "🧪 Running Python tests..."
	cd services/analytics-worker && uv run pytest

# --- Lint ---
lint: ## Run linters
	@echo "🔍 Linting Go code..."
	cd services/gateway && golangci-lint run
	cd services/product-service && golangci-lint run
	cd services/order-service && golangci-lint run
	cd services/payment-service && golangci-lint run
	cd services/user-service && golangci-lint run
	cd services/delivery-service && golangci-lint run
	@echo "🔍 Linting Rust code..."
	cd services/inventory-service && cargo clippy
	@echo "🔍 Linting Python code..."
	cd services/analytics-worker && uv run --dev ruff check .
	cd services/analytics-worker && uv run --dev ruff format --check .

lint-python: uv-install ## Lint Python code only
	@echo "🔍 Linting Python code..."
	cd services/analytics-worker && uv run --dev ruff check .
	cd services/analytics-worker && uv run --dev ruff format --check .

format-python: uv-install ## Format Python code
	@echo "✨ Formatting Python code..."
	cd services/analytics-worker && uv run --dev ruff format .

# --- Database ---
MIGRATE_DB = $(DATABASE_URL)
MIGRATE_PRODUCT_DB = $(subst ?sslmode,?x-migrations-table=product_migrations&sslmode,$(DATABASE_URL))
MIGRATE_ORDER_DB = $(subst ?sslmode,?x-migrations-table=order_migrations&sslmode,$(DATABASE_URL))
MIGRATE_PAYMENT_DB = $(subst ?sslmode,?x-migrations-table=payment_migrations&sslmode,$(DATABASE_URL))
MIGRATE_USER_DB = $(subst ?sslmode,?x-migrations-table=user_migrations&sslmode,$(DATABASE_URL))
MIGRATE_DELIVERY_DB = $(subst ?sslmode,?x-migrations-table=delivery_migrations&sslmode,$(DATABASE_URL))

db-migrate: ## Run database migrations
	@echo "🗄️  Running database migrations..."
	cd services/product-service && migrate -path internal/migrations -database "$(MIGRATE_PRODUCT_DB)" up
	cd services/order-service && migrate -path internal/migrations -database "$(MIGRATE_ORDER_DB)" up
	cd services/payment-service && migrate -path internal/migrations -database "$(MIGRATE_PAYMENT_DB)" up
	cd services/user-service && migrate -path internal/migrations -database "$(MIGRATE_USER_DB)" up
	cd services/delivery-service && migrate -path internal/migrations -database "$(MIGRATE_DELIVERY_DB)" up
	cat services/inventory-service/migrations/001_create_inventory_schema.sql | $(PSQL)
	@echo "✅ Migrations complete"

db-seed: ## Seed database with test data
	@echo "🌱 Seeding database..."
	cat scripts/seed-data.sql | $(PSQL)
	@echo "✅ Database seeded"

db-rollback: ## Rollback last migration per service
	@echo "⏪ Rolling back last migration per service..."
	cd services/product-service && migrate -path internal/migrations -database "$(MIGRATE_PRODUCT_DB)" down 1
	cd services/order-service && migrate -path internal/migrations -database "$(MIGRATE_ORDER_DB)" down 1
	cd services/payment-service && migrate -path internal/migrations -database "$(MIGRATE_PAYMENT_DB)" down 1
	cd services/user-service && migrate -path internal/migrations -database "$(MIGRATE_USER_DB)" down 1
	cd services/delivery-service && migrate -path internal/migrations -database "$(MIGRATE_DELIVERY_DB)" down 1
	@echo "✅ Rollback complete"

db-rollback-all: ## Rollback ALL migrations (drops all service tables)
	@echo "⏪ Rolling back ALL migrations..."
	cd services/product-service && migrate -path internal/migrations -database "$(MIGRATE_PRODUCT_DB)" down -all
	cd services/order-service && migrate -path internal/migrations -database "$(MIGRATE_ORDER_DB)" down -all
	cd services/payment-service && migrate -path internal/migrations -database "$(MIGRATE_PAYMENT_DB)" down -all
	cd services/user-service && migrate -path internal/migrations -database "$(MIGRATE_USER_DB)" down -all
	cd services/delivery-service && migrate -path internal/migrations -database "$(MIGRATE_DELIVERY_DB)" down -all
	$(PSQL) -c "DROP SCHEMA IF EXISTS inventory CASCADE; CREATE SCHEMA inventory;"
	@echo "✅ All migrations rolled back"

db-reset: ## Drop all data and re-migrate + seed
	@echo "💣 Resetting database..."
	$(PSQL) -c "DROP SCHEMA IF EXISTS catalog, orders, users, payments, inventory, delivery CASCADE;"
	$(PSQL) -c "DROP TABLE IF EXISTS product_migrations, order_migrations, payment_migrations, user_migrations, delivery_migrations, schema_migrations;"
	$(PSQL) -c "CREATE SCHEMA catalog; CREATE SCHEMA orders; CREATE SCHEMA users; CREATE SCHEMA payments; CREATE SCHEMA inventory; CREATE SCHEMA delivery;"
	@echo "🗑️  Dropped all schemas, re-running migrations..."
	$(MAKE) db-migrate
	$(MAKE) db-seed

# --- Docker ---
docker-build: ## Build Docker images
	@echo "🐳 Building Docker images..."
	docker build -t shinkansen/gateway:latest -f services/gateway/Dockerfile .
	docker build -t shinkansen/product-service:latest -f services/product-service/Dockerfile .
	docker build -t shinkansen/order-service:latest -f services/order-service/Dockerfile .
	docker build -t shinkansen/payment-service:latest -f services/payment-service/Dockerfile .
	docker build -t shinkansen/user-service:latest -f services/user-service/Dockerfile .
	docker build -t shinkansen/delivery-service:latest -f services/delivery-service/Dockerfile .
	docker build -t shinkansen/inventory-service:latest -f services/inventory-service/Dockerfile .
	docker build -t shinkansen/analytics-worker:latest -f services/analytics-worker/Dockerfile .
	@echo "✅ Docker images built"

docker-push: docker-build ## Push Docker images
	docker push shinkansen/gateway:latest
	docker push shinkansen/product-service:latest
	docker push shinkansen/order-service:latest
	docker push shinkansen/payment-service:latest
	docker push shinkansen/user-service:latest
	docker push shinkansen/delivery-service:latest
	docker push shinkansen/inventory-service:latest
	docker push shinkansen/analytics-worker:latest

# --- Kubernetes ---
k8s-apply: ## Apply Kubernetes manifests
	kubectl apply -f deploy/k8s/base

k8s-delete: ## Delete Kubernetes resources
	kubectl delete -f deploy/k8s/base

k8s-logs: ## View Kubernetes logs
	kubectl logs -f -l app=shinkansen

# --- Clean ---
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f services/*/coverage.out
	@echo "✅ Cleaned build artifacts"

clean-all: clean ## Clean everything including generated code
	rm -rf gen/
	rm -f services/*/internal/db/*.go
	@echo "✅ Cleaned all generated code"
