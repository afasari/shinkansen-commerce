# Shinkansen Commerce

A spec-first polyglot microservices e-commerce platform, built with Go, Rust, Python, and Vue.js.

## Philosophy

**The protobuf specification is the source of truth. Code is generated from it.**

- Decoupled, type-safe microservices with gRPC internal communication
- Polyglot monorepo (Go, Rust, Python, TypeScript)
- Spec-first development (proto -> code generation via `make gen`)
- Role-based access control (customer vs admin)

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Go Services** | Go 1.24 |
| **Inventory Service** | Rust (tonic, sqlx) |
| **Analytics Worker** | Python 3.11+ (uv-managed) |
| **Frontend** | Vue 3 + TypeScript + Tailwind CSS |
| **API Gateway** | Go (REST -> gRPC proxy) |
| **Communication** | gRPC (internal), REST (external via gateway) |
| **Data Access** | sqlc (SQL -> typed Go code), sqlx (Rust) |
| **Database** | PostgreSQL 15 |
| **Cache** | Redis 7 |
| **Code Generation** | buf (proto), sqlc (SQL), protoc-gen-openapiv2 (OpenAPI) |
| **Containerization** | Docker Compose, Kubernetes (partial) |
| **CI/CD** | GitHub Actions (lint active, test/build commented out) |

## Architecture

```
                    ┌────────────────────────────────┐
                    │        Frontend (Vue 3)         │
                    │    :5173 (dev, proxies to :8080) │
                    └──────────────┬─────────────────┘
                                   │
                    ┌──────────────▼─────────────────┐
                    │     Gateway (Go) :8080          │
                    │  REST ↔ gRPC, JWT auth, RBAC    │
                    └──────────────┬─────────────────┘
                                   │ gRPC
          ┌────────────┬───────────┼───────────┬────────────┐
          ▼            ▼           ▼           ▼            ▼
   ┌────────────┐ ┌─────────┐ ┌────────┐ ┌─────────┐ ┌──────────┐
   │  Product   │ │  Order  │ │  User  │ │ Payment │ │ Delivery │
   │  (Go)      │ │  (Go)   │ │  (Go)  │ │  (Go)   │ │  (Go)    │
   │ :9091      │ │ :9092   │ │ :9103  │ │ :9104   │ │ :9106    │
   └────────────┘ └────┬────┘ └────────┘ └─────────┘ └──────────┘
                       │
              ┌────────▼────────┐
              │   Inventory     │
              │   (Rust) :9105  │
              └─────────────────┘
                       │
              ┌────────▼────────┐
              │   Analytics     │
              │   (Python)      │
              └─────────────────┘
```

### Service Ports

| Service | Language | gRPC Port | Metrics Port |
|---------|----------|-----------|--------------|
| gateway | Go | — | 8080 (HTTP) |
| product-service | Go | 9091 | 8091 |
| order-service | Go | 9092 | 8092 |
| user-service | Go | 9103 | 8103 |
| payment-service | Go | 9104 | 8104 |
| inventory-service | Rust | 9105 | 8105 |
| delivery-service | Go | 9106 | 8106 |
| analytics-worker | Python | — | — |
| frontend | Vue 3 + TS | — | 5173 (dev) |

## Monorepo Structure

```
shinkansen-commerce/
├── proto/                          # Protocol Buffers (source of truth)
├── gen/                            # Generated code (DO NOT EDIT)
│   ├── proto/go/                   # Generated Go gRPC code
│   └── proto/rust/                 # Generated Rust proto code
├── services/
│   ├── gateway/                    # REST↔gRPC gateway (Go)
│   ├── product-service/            # Product catalog (Go)
│   ├── order-service/              # Orders & cart (Go)
│   ├── user-service/               # Auth & users (Go)
│   ├── payment-service/            # Payments (Go)
│   ├── inventory-service/          # Stock management (Rust)
│   ├── delivery-service/           # Delivery & shipping (Go)
│   ├── analytics-worker/           # Analytics (Python)
│   └── frontend/                   # Customer & admin UI (Vue 3)
├── deploy/
│   └── k8s/base/                   # Kubernetes manifests (gateway + product only)
├── scripts/                        # Utility scripts
├── docs/                           # VitePress documentation site
├── docker-compose.yml              # Local development (PG + Redis + all services)
├── Makefile                        # Build automation
├── go.work                         # Go workspace
└── AGENTS.md                       # AI agent instructions
```

## Quick Start

### Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- [Rust](https://rustup.rs/) (for inventory-service)
- [Python 3.11+](https://www.python.org/) with [uv](https://docs.astral.sh/uv/) (for analytics-worker)
- [Docker & Docker Compose](https://www.docker.com/products/docker-desktop)
- [buf](https://docs.buf.build/installation)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [Node.js 18+](https://nodejs.org/) (for frontend)

### 1. Start Infrastructure

```bash
make up
```

This starts PostgreSQL 15, Redis 7, and all microservices via Docker Compose.

### 2. Generate Code

```bash
make gen
```

Generates Go gRPC code, Rust proto code, sqlc typed queries, and OpenAPI docs from protobuf specs.

### 3. Build

```bash
make build-all        # Build all Go services to bin/
cd services/frontend && npm install && npm run dev   # Start frontend dev server
```

### 4. Run Tests

```bash
make test             # All tests (Go + Rust)
make lint             # All linters
cd services/frontend && npm run build  # Verify frontend builds
```

### 5. Create an Admin User

Register via the frontend or API, then promote in the database:

```sql
UPDATE users.users SET role = 'admin' WHERE email = 'your@email.com';
```

## Available Commands

```bash
# Infrastructure
make up              # Start all services (docker compose up)
make down            # Stop all services
make logs            # View logs

# Code Generation
make gen             # Generate all code (proto + sqlc + openapi)
make proto-gen       # Generate Go protobuf code only
make sqlc-gen        # Generate sqlc code (product + order services)

# Build
make build-all       # Build all Go services
make build-inventory # Build Rust inventory service

# Test
make test            # Run all tests
make test-coverage   # Run tests with coverage
make test-integration # Run integration tests (requires docker)

# Lint
make lint            # Run all linters (Go + Rust + Python)

# Database
make db-migrate      # Run all migrations
make db-rollback     # Rollback last migration per service
```

## Frontend

The frontend (`services/frontend/`) is a Vue 3 + TypeScript application with:

- **Customer pages**: Home, product browsing, search, cart, 4-step checkout, order tracking
- **Account pages**: Profile, address management, order history
- **Admin pages**: Dashboard, product CRUD, order management, inventory, delivery slots, shipments, payments
- **Bilingual i18n**: English + Japanese
- **Role-aware**: Admin link and `/admin/*` routes only accessible when `role === "admin"`
- **Client-side cart**: localStorage-based (no backend cart API)

```bash
cd services/frontend
npm install
npm run dev          # Dev server at :5173, proxies /v1 to :8080
npm run build        # Production build
```

## Database

Each service owns a PostgreSQL schema (`catalog`, `orders`, `users`, `payments`, `inventory`, `delivery`) in a single `shinkansen` database.

Default local connection: `postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable`

## Testing

```bash
make test                           # Unit tests (Go + Rust)
make test-integration               # Integration tests (starts docker-compose)
cd services/frontend && npm run build  # Frontend typecheck + build
cd services/analytics-worker && uv run pytest  # Python tests
```

## Deployment

### Docker Compose (Local)

```bash
make up
```

### Kubernetes (Partial)

```bash
make k8s-apply       # Apply manifests for gateway + product-service
```

Note: Only gateway and product-service have K8s manifests. Other services need manifests added to `deploy/k8s/base/k8s.yaml`.

## Documentation

- [Architecture Overview](docs/architecture/overview.md)
- [High-Level Design](docs/architecture/high-level-design.md)
- [Low-Level Design](docs/architecture/low-level-design.md)
- [API Documentation](docs/api/)
- [Development Guide](docs/development/)
- [Runbooks](docs/runbooks/)

## CI

`.github/workflows/ci-cd.yml` runs on push/PR to `main` and `develop`. The active pipeline runs lint (Go, Rust, Python) and proto format checks. Test and build jobs are commented out.

## License

MIT License. Copyright 2026 Ba'tiar Afas Rahmamulia.
