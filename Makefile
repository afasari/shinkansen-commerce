.PHONY: help proto build run test clean

help: ## Display this help screen
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

proto: ## Generate gRPC code from proto files (Go/Rust/Python)
    @echo "Generating protobufs..."
    # ./scripts/generate-proto.sh

build: ## Build all services
    @echo "Building Docker images..."
    # docker-compose build

run: ## Run all services locally
    docker-compose up -d

test: ## Run unit tests
    @echo "Running tests..."
    # go test ./...
    # cd services/inventory-service && cargo test
