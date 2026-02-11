# Development Setup

## Prerequisites

Install the following tools:

### Go 1.21+

```bash
# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# macOS
brew install go
```

### Docker & Docker Compose

```bash
# Linux
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# macOS
brew install --cask docker
```

### buf (Protocol Buffer tool)

```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

### sqlc (SQL code generation)

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Node.js 18+

```bash
# Using nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
source ~/.bashrc
nvm install 18

# Or using homebrew
brew install node@18
```

### uv (Python package manager)

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

## Project Structure

```
shinkansen-commerce/
├── proto/                   # Protocol Buffers (source of truth)
│   ├── product/             # Product service definitions
│   ├── order/               # Order service definitions
│   ├── user/                # User service definitions
│   ├── payment/             # Payment service definitions
│   ├── inventory/           # Inventory service definitions
│   ├── delivery/            # Delivery service definitions
│   └── shared/             # Shared types
│
├── services/               # Service implementations
│   ├── gateway/            # HTTP API gateway
│   ├── product-service/    # Product catalog (Go)
│   ├── order-service/      # Order management (Go)
│   ├── user-service/       # User auth (Go)
│   ├── payment-service/    # Payment processing (Go)
│   ├── inventory-service/  # Stock management (Rust)
│   ├── delivery-service/   # Delivery logistics (Go)
│   └── analytics-worker/   # Data processing (Python)
│
├── gen/                   # Generated code
│   └── proto/go/          # Go protobuf code
│
├── deploy/                # Deployment configs
│   ├── docker-compose.yml  # Local development
│   ├── k8s/              # Kubernetes manifests
│   └── terraform/        # Infrastructure as code
│
├── docs/                  # Documentation (VitePress)
├── scripts/               # Utility scripts
└── Makefile              # Build automation
```

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/afasari/shinkansen-commerce.git
cd shinkansen-commerce
```

### 2. Start Infrastructure

```bash
make up
```

This starts PostgreSQL, Redis, and Kafka.

### 3. Generate Code

```bash
make gen
```

Generates:
- Go protobuf code from proto files
- SQL code from sqlc
- OpenAPI specs from proto files

### 4. Build Services

```bash
make build-all
```

Builds all 7 services to `bin/` directory.

### 5. Run Services

```bash
# Option 1: Run all in Docker
make up

# Option 2: Run locally (requires infrastructure running)
./bin/gateway &
./bin/product-service &
./bin/order-service &
./bin/user-service &
./bin/payment-service &
./bin/inventory-service &
./bin/delivery-service &
```

### 6. Test

```bash
make test-integration
```

## Working with Protobufs

### Add New Message Type

```protobuf
// proto/product/product_messages.proto

message NewProductRequest {
  string name = 1;
  string description = 2;
  shinkansen.common.Money price = 3;
}
```

### Add New RPC Method

```protobuf
// proto/product/product_service.proto

service ProductService {
  rpc GetFeaturedProducts(GetFeaturedProductsRequest) returns (GetFeaturedProductsResponse) {
    option (google.api.http) = {
      get: "/v1/products/featured"
    };
  }
}
```

### Regenerate Code

```bash
make proto-gen
```

### Regenerate OpenAPI Docs

```bash
make proto-openapi-gen
```

## Working with Database Migrations

### Add New Migration

```sql
-- services/product-service/internal/migrations/001_new_table.sql

CREATE TABLE catalog.new_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);
```

### Run Migrations

```bash
make db-migrate
```

### Rollback Migration

```bash
make db-rollback
```

## Running Tests

### Unit Tests (Single Service)

```bash
cd services/product-service
go test ./... -v -cover
```

### All Unit Tests

```bash
make test
```

### Integration Tests

```bash
make test-integration
```

### Test with Coverage

```bash
make test-coverage
```

## Local Documentation

### Generate API Docs

```bash
make proto-openapi-gen
```

Access Swagger UI at `http://localhost:8080/swagger/`

### Run VitePress Dev Server

```bash
cd docs
npm install
npm run dev
```

Access documentation at `http://localhost:5173`

## Environment Variables

Service configurations use environment variables. Create `.env` file:

```bash
# Database
DATABASE_URL=postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-key-change-in-production

# Service Addresses
PRODUCT_SERVICE_GRPC_ADDRESS=localhost:9091
ORDER_SERVICE_GRPC_ADDRESS=localhost:9092
USER_SERVICE_GRPC_ADDRESS=localhost:9103
PAYMENT_SERVICE_GRPC_ADDRESS=localhost:9104
INVENTORY_SERVICE_GRPC_ADDRESS=localhost:9105
DELIVERY_SERVICE_GRPC_ADDRESS=localhost:9106
```

## Debugging

### View Service Logs

```bash
# All services
make logs

# Specific service
docker-compose logs product-service
```

### Connect to PostgreSQL

```bash
docker exec -it shinkansen-postgres psql -U shinkansen -d shinkansen
```

### Connect to Redis

```bash
docker exec -it shinkansen-redis redis-cli
```

### gRPC Debugging

```bash
# List all gRPC methods
grpcurl -plaintext localhost:9091 list shinkansen.product.ProductService

# Call gRPC method
grpcurl -plaintext -d '{"page":1,"limit":20}' localhost:9091 shinkansen.product.ProductService/ListProducts
```

## Common Issues

### Port Already in Use

```bash
# Find what's using the port
lsof -i :8080

# Kill the process
kill -9 <PID>
```

### Database Connection Failed

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Build Errors

```bash
# Clean and rebuild
make clean-all
make init-deps
make build-all
```

## IDE Setup

### VS Code

Recommended extensions:
- Go (Google)
- Protobuf (pbkit.pro)
- Docker
- VitePress
- Mermaid Preview

### GoLand (JetBrains)

1. Open project as Go module
2. Enable Go modules integration
3. Configure workspace

## Development Workflow

```mermaid
graph LR
    A[Edit Proto] --> B[Regenerate Code]
    B --> C[Write Tests]
    C --> D[Run Tests]
    D --> E{Pass?}
    E -->|Yes| F[Commit]
    E -->|No| A
```

## Next Steps

- [Testing Guide](/development/testing)
- [Code Generation](/development/code-generation)
- [API Reference](/api/overview)
- [Architecture](/architecture/overview)
