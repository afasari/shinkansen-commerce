.PHONY: help up down logs proto-gen sqlc-gen init-deps build test lint clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Infrastructure ---
up: ## Start Infrastructure (Docker Compose)
	docker-compose -f deploy/docker-compose.yml up -d

down: ## Stop Infrastructure
	docker-compose -f deploy/docker-compose.yml down

logs: ## View Infrastructure Logs
	docker-compose -f deploy/docker-compose.yml logs -f

ps: ## Show running containers
	docker-compose -f deploy/docker-compose.yml ps

# --- Generation ---
proto-gen: ## Generate gRPC code from protobufs
	@echo "🔄 Generating Go protobuf code..."
	@mkdir -p gen/proto/go
	buf generate --template proto/buf.gen.yaml
	@echo "✅ Go protobuf code generated"

proto-lint: ## Lint protobuf files
	buf lint proto

proto-format: ## Format protobuf files
	buf format -w proto

sqlc-gen: ## Generate SQL code for Go services
	@echo "🔄 Generating SQL code for Product Service..."
	cd services/product-service && sqlc generate
	@echo "✅ Product Service SQL generated"
	@echo "🔄 Generating SQL code for Order Service..."
	cd services/order-service && sqlc generate
	@echo "✅ Order Service SQL generated"
	@echo "🔄 Generating SQL code for Payment Service..."
	cd services/payment-service && sqlc generate
	@echo "✅ Payment Service SQL generated"
	@echo "🔄 Generating SQL code for User Service..."
	cd services/user-service && sqlc generate
	@echo "✅ User Service SQL generated"
	@echo "🔄 Generating SQL code for Delivery Service..."
	cd services/delivery-service && sqlc generate
	@echo "✅ Delivery Service SQL generated"

gen: proto-gen sqlc-gen ## Generate all code (protobuf + sqlc)

# --- Dependencies ---
init-deps: ## Download all dependencies
	@echo "📦 Installing Go dependencies..."
	cd services/gateway && go mod tidy
	cd services/product-service && go mod tidy
	cd services/order-service && go mod tidy
	cd services/payment-service && go mod tidy
	cd services/user-service && go mod tidy
	cd services/delivery-service && go mod tidy
	cd services/shared/go && go mod tidy
	@echo "📦 Installing Python dependencies..."
	cd services/analytics-worker && pip install -r requirements.txt
	@echo "✅ All dependencies installed"

# --- Build ---
build: ## Build all services
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
	@echo "🔨 Building Inventory Service (Rust)..."
	cd services/inventory-service && cargo build --release
	@echo "✅ All services built"

build-gateway: ## Build Gateway only
	cd services/gateway && go build -o ../../bin/gateway ./cmd/gateway

build-product: ## Build Product Service only
	cd services/product-service && go build -o ../../bin/product-service ./cmd/product-service

build-order: ## Build Order Service only
	cd services/order-service && go build -o ../../bin/order-service ./cmd/order-service

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
	@echo "🧪 Running Shared tests..."
	cd services/shared/go && go test ./...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	cd services/gateway && go test -coverprofile=coverage.out ./...
	cd services/product-service && go test -coverprofile=coverage.out ./...
	cd services/order-service && go test -coverprofile=coverage.out ./...

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
	cd services/analytics-worker && ruff check .

# --- Database ---
db-migrate: ## Run database migrations
	@echo "🗄️  Running database migrations..."
	cd services/product-service && go run cmd/migrate/main.go
	cd services/order-service && go run cmd/migrate/main.go
	cd services/payment-service && go run cmd/migrate/main.go
	cd services/user-service && go run cmd/migrate/main.go
	cd services/delivery-service && go run cmd/migrate/main.go

db-rollback: ## Rollback database migrations
	@echo "⏪ Rolling back database migrations..."
	cd services/product-service && go run cmd/migrate/main.go down
	cd services/order-service && go run cmd/migrate/main.go down
	cd services/payment-service && go run cmd/migrate/main.go down
	cd services/user-service && go run cmd/migrate/main.go down
	cd services/delivery-service && go run cmd/migrate/main.go down

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
	@echo "✅ Docker images built"

docker-push: docker-build ## Push Docker images
	docker push shinkansen/gateway:latest
	docker push shinkansen/product-service:latest
	docker push shinkansen/order-service:latest
	docker push shinkansen/payment-service:latest
	docker push shinkansen/user-service:latest
	docker push shinkansen/delivery-service:latest
	docker push shinkansen/inventory-service:latest

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
	cd services/inventory-service && cargo clean
	@echo "✅ Cleaned build artifacts"

clean-all: clean ## Clean everything including generated code
	rm -rf gen/
	rm -f services/*/internal/db/*.go
	@echo "✅ Cleaned all generated code"
