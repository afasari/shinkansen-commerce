.PHONY: help up down logs proto-gen init-deps

help: ## Display this help
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Infrastructure ---
up: ## Start Infrastructure
    docker-compose -f deploy/docker-compose.yml up -d

down: ## Stop Infrastructure
    docker-compose -f deploy/docker-compose.yml down

logs: ## View Logs
    docker-compose -f deploy/docker-compose.yml logs -f

# --- Generation ---
proto-gen: ## Generate gRPC code
    @echo "Generating Go..."
    docker run --rm -v $(PWD)/proto:/proto -v $(PWD)/gen:/gen \
        grpc/go:1.20 \
        protoc -I/proto --go_out=/gen --go-grpc_out=/gen inventory.proto
    @echo "Generating Rust (via cargo build)..."
    cd services/inventory-service && cargo build
    @echo "✅ Generation Complete."

# --- Dependencies ---
init-deps: ## Download all dependencies (Go, Python)
    @echo "Installing Go dependencies..."
    cd services/gateway-service && go mod tidy
    cd services/order-service && go mod tidy
    cd services/product-service && go mod tidy
    @echo "Installing Python dependencies via uv..."
    cd services/analytics-worker && uv sync

