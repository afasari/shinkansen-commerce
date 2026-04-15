# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a **spec-first polyglot microservices monorepo** for a Japanese e-commerce platform. The Protocol Buffers in `proto/` are the source of truth - all other code is generated from these specifications.

### Services

| Service | Language | Port | Purpose |
|---------|----------|------|---------|
| gateway | Go | 8080 | API Gateway, gRPC↔REST translation, auth |
| product-service | Go | 9091 | Product catalog, search, caching |
| order-service | Go | 9092 | Orders, cart, checkout |
| payment-service | Go | 9104 | Payments, Konbini, point system |
| user-service | Go | 9103 | Auth, profiles, addresses |
| inventory-service | Rust | 9105 | Stock management, allocation (performance-critical) |
| delivery-service | Go | 9106 | Same-day delivery, logistics |
| analytics-worker | Python | - | Batch analytics, AI insights |

### Communication Flow

- **External**: REST via Gateway (port 8080)
- **Internal**: gRPC between services
- **Events**: Kafka (not in docker-compose.yml - would need to be added)
- **Data**: PostgreSQL 15 + Redis 7 caching

## Essential Commands

```bash
# Infrastructure (Docker Compose)
make up              # Start postgres, redis, and all microservices (runs make gen first)
make down            # Stop all infrastructure

# Code Generation (CRITICAL: run after proto changes)
make gen             # Generate all: protobufs, sqlc, OpenAPI docs
make proto-gen       # Generate Go gRPC code from proto/
make proto-openapi-gen  # Generate OpenAPI/Swagger docs
make sqlc-gen        # Generate Go database code from .sql queries

# Building
make build           # Build all services to bin/
make build-<service> # Build specific service (e.g., build-product)

# Testing
make test            # Run all Go and Rust tests
cd services/<service> && go test ./...  # Test specific Go service
cd services/<service> && go test ./... -run TestFunctionName  # Run specific test
cd services/inventory-service && cargo test -- --test-threads=1  # Rust tests (single-threaded required)
make test-coverage   # Tests with coverage

# Linting
make lint            # golangci-lint for Go, clippy for Rust, ruff for Python

# Database
make db-migrate      # Run migrations for all services
make db-rollback     # Rollback last migration per service

# Python (analytics-worker)
make uv-sync         # Install Python dependencies via uv
make uv-run CMD="<command>"  # Run Python command
```

## Service Structure

### Go Services (gateway, product, order, payment, user, delivery)

```
services/<name>/
├── cmd/<name>/          # Main entry point
├── internal/
│   ├── handler/         # gRPC server implementation
│   ├── service/         # Business logic layer
│   ├── repository/      # Data access abstraction
│   ├── db/              # sqlc-generated code (DO NOT EDIT)
│   ├── queries/         # SQL queries for sqlc
│   ├── migrations/      # Database migrations
│   ├── cache/           # Redis caching layer
│   ├── config/          # Configuration loading
│   └── pkg/             # Internal utilities
├── sqlc.yaml            # sqlc configuration
├── Dockerfile
└── go.mod
```

**Pattern**: handler → service → repository → db (sqlc) → PostgreSQL

### Rust Service (inventory-service)

```
services/inventory-service/
├── src/
│   ├── main.rs          # Entry point
│   ├── service.rs       # gRPC service implementation
│   ├── repository.rs    # Data access with sqlx
│   ├── config.rs        # Configuration
│   ├── database.rs      # Database connection
│   └── health.rs        # Health check endpoint
├── migrations/          # Database migrations
├── Cargo.toml
└── build.rs             # Build script (prost codegen)
```

Uses `sqlx` with compile-time checked queries (macros).

### Python Service (analytics-worker)

```
services/analytics-worker/
├── analytics_worker/
│   └── cli.py           # CLI entry point
├── tests/
├── pyproject.toml       # uv-based dependency management
└── Dockerfile
```

Uses `uv` for fast Python package management.

## Code Generation Workflow

### Adding a New API Endpoint

1. **Edit proto file** in `proto/<service>/<service>_service.proto`
2. **Run**: `make proto-gen` → generates Go code in `gen/proto/go/`
3. **Implement** the handler in `services/<service>/internal/handler/`
4. **For Go services with DB**: Add SQL queries to `internal/queries/*.sql`, then `make sqlc-gen`

### Adding SQL Queries (Go Services)

1. Create `.sql` file in `services/<service>/internal/queries/`
2. Run `make sqlc-gen` → generates typed Go code in `internal/db/`
3. Use generated interface in repository layer

**IMPORTANT**: `make sqlc-gen` only processes **product-service** and **order-service**. Other services have sqlc steps commented out in the Makefile.

### Database Migrations

Each service owns a PostgreSQL schema (e.g., `catalog`, `orders`, `users`, `payments`, `inventory`, `delivery`) in a single `shinkansen` database.

**Go services**: Migrations in `internal/migrations/` with up/down pairs: `XXXXXXX_description.{up,down}.sql`
**Rust inventory-service**: Migrations in `migrations/` (root level) with single SQL files (not up/down pairs)

```bash
make db-migrate                          # All services
make db-rollback                         # Rollback last migration per service
cd services/<service> && migrate -path internal/migrations -database "$DATABASE_URL" up
```

## Go Workspace

This uses Go 1.24+ workspace mode (`go.work`). The workspace includes: all Go services (gateway, product-service, order-service, payment-service, user-service, delivery-service) plus services/shared/go. Services reference generated proto code via:

```go
import "github.com/afasari/shinkansen-commerce/gen/proto/go/<service>"
```

The workspace is defined in `go.work` at the repository root.

## Infrastructure

### Local Development

`docker-compose.yml` starts PostgreSQL 15, Redis 7, and all microservices. **Note**: Kafka, MinIO, Jaeger, Prometheus, and Grafana are NOT included in docker-compose.yml (despite being mentioned in README.md).

Database URLs default to `postgres://shinkansen:shinkansen_dev_password@postgres:5432/shinkansen?sslmode=disable`.

`make up` runs `make gen` as a prerequisite before starting docker compose.

### Service Ports

| Service | gRPC | HTTP/Metrics |
|---------|------|--------------|
| gateway | - | 8080 |
| product | 9091 | 8091 |
| order | 9092 | 8092 |
| user | 9103 | 8103 |
| payment | 9104 | 8104 |
| inventory | 9105 | 8105 |
| delivery | 9106 | 8106 |

### Observability

- **Grafana**: `http://localhost:3000` (admin/admin) — not in docker-compose.yml
- **Jaeger tracing**: `http://localhost:16686` — not in docker-compose.yml
- **Prometheus metrics**: Exposed on each service's METRICS_PORT (8091-8106)

## Japan-Specific Features

- **Konbini Payments**: Convenience store payment (7-Eleven, Lawson, FamilyMart, etc.) - see `proto/payment/konbini.proto`
- **Point System**: Multi-vendor point ecosystem - see `proto/payment/points.proto`
- **Same-day Delivery**: With geospatial queries via PostGIS - see delivery-service

## Shared Code

`services/shared/` contains language-agnostic utilities:
- `go/` - Shared Go utilities (minimal, mostly proto code is shared)
- `python/` - Shared Python code
- `rust/` - Shared Rust code

## Important Notes

- **Never edit generated code**: `gen/proto/go/*`, `services/*/internal/db/*.go` (sqlc output)
- **Proto is source of truth**: Always edit `.proto` files, regenerate code
- **sqlc requires valid SQL**: Queries in `internal/queries/` must be valid PostgreSQL
- **Rust build**: `inventory-service` has a `build.rs` that compiles protoc via prost

## Go Conventions

- **Import groups**: stdlib → external (alphabetical) → internal (alphabetical)
- **Error wrapping**: `fmt.Errorf("failed to X: %w", err)`
- **Structured logging**: `go.uber.org/zap`
- **Testing**: Table-driven with `t.Run()`, using `github.com/stretchr/testify/assert` + `require`
- **Redis mocking in tests**: `github.com/alicebob/miniredis/v2`
- **Config via env vars** with defaults using `getEnv()` in `internal/config/config.go`
- **UUID conversion**: Use `pgutil.ToPG()` / `pgutil.FromPG()` for PostgreSQL UUID columns
- **sqlc driver**: `pgx/v5`

## Rust Conventions (inventory-service)

- Tests must run single-threaded: `cargo test -- --test-threads=1`
- Uses `sqlx` with compile-time checked queries (macros)
- Migrations are single SQL files in `migrations/` (not up/down pairs like Go services)

## Python Conventions (analytics-worker)

```bash
cd services/analytics-worker && uv sync        # Install deps
cd services/analytics-worker && uv run pytest   # Run tests
cd services/analytics-worker && uv run ruff check . && uv run ruff format --check .  # Lint
cd services/analytics-worker && uv run ruff format .   # Format
```

## Verification Workflow

Before committing changes, run in order:
1. `make gen` — Regenerate all code after proto changes
2. `make build-all` — Ensure all services build
3. `make lint` — Run linters (Go: golangci-lint, Rust: clippy, Python: ruff)
4. `make test` — Run all tests

## CI/CD

`.github/workflows/ci-cd.yml`: Only the **lint** job is active. Test and build jobs are commented out. CI runs: `make proto-gen` → `make sqlc-gen` → `make lint` → proto format check. Go 1.24, golangci-lint v2.9.0.
