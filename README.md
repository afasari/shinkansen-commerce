# 🚄 Shinkansen Commerce

A high-performance, spec-first polyglot monorepo e-commerce platform designed for the Japanese market, built with best practices used by companies like Rakuten and PayPay.

## 🎯 Philosophy

**The Specification (.proto) is the source of truth. Code is just a byproduct.**

This project demonstrates:
- Decoupled, type-safe microservices architecture
- Polyglot monorepo (Go, Rust, Python)
- Production-grade infrastructure (Kubernetes, Docker)
- Japanese e-commerce features (Konbini payments, Point systems)

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| **Core Services** | Go 1.21 |
| **Performance Services** | Rust |
| **Analytics/AI** | Python 3.11 |
| **API Gateway** | Go (grpc-gateway) |
| **Communication** | gRPC (Internal), REST (External) |
| **Data Access** | sqlc (SQL → Code generation) |
| **Database** | PostgreSQL 15 |
| **Cache** | Redis 7 |
| **Message Queue** | Kafka 3.5 |
| **Object Storage** | MinIO |
| **Observability** | Prometheus, Grafana, Jaeger |
| **Container Orchestration** | Kubernetes |
| **CI/CD** | GitHub Actions |
| **Infrastructure** | Terraform, Docker Compose |

## 📁 Monorepo Structure

```
shinkansen-commerce/
├── proto/                          # Protocol Buffers (Source of Truth)
│   ├── shared/                      # Shared types
│   ├── product/                     # Product service definitions
│   ├── order/                       # Order service definitions
│   ├── payment/                     # Payment service definitions
│   ├── konbini/                     # Konbini payments
│   ├── points/                      # Point system
│   ├── inventory/                   # Inventory service (Rust)
│   ├── user/                        # User service
│   ├── delivery/                    # Delivery service
│   └── buf.yaml                     # Buf configuration
│
├── services/                       # Service Implementations
│   ├── gateway/                     # Go - API Gateway
│   ├── product-service/             # Go - Product management
│   ├── order-service/               # Go - Order processing
│   ├── payment-service/             # Go - Payment processing
│   ├── inventory-service/           # Rust - High-performance inventory
│   ├── user-service/                # Go - User management
│   ├── delivery-service/            # Go - Delivery optimization
│   ├── analytics-worker/            # Python - Analytics & AI
│   └── shared/                     # Shared utilities
│
├── deploy/                         # Infrastructure
│   ├── k8s/                        # Kubernetes manifests
│   │   ├── base/                   # Base resources
│   │   └── overlays/               # Environment-specific
│   ├── docker-compose.yml           # Local development
│   └── terraform/                  # IaC
│
├── scripts/                        # Utility scripts
├── docs/                          # Documentation
├── Makefile                       # Build automation
├── go.work                        # Go workspace
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Docker & Docker Compose](https://www.docker.com/products/docker-desktop)
- [buf](https://docs.buf.build/installation)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [Node.js 18+](https://nodejs.org/) (for some tools)

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/shinkansen-commerce.git
cd shinkansen-commerce
```

### 2. Start Infrastructure

```bash
make up
```

This starts:
- PostgreSQL 15
- Redis 7
- Kafka 3.5
- MinIO
- Jaeger (tracing)
- Prometheus (metrics)
- Grafana (dashboards)

### 3. Generate Code

```bash
make gen
```

This generates:
- Go gRPC code from protobufs
- SQL queries from sqlc

### 4. Download Dependencies

```bash
make init-deps
```

### 5. Build Services

```bash
make build
```

### 6. Run Services

```bash
# Run individual services
./bin/gateway
./bin/product-service

# Or run all services in background
make run-all
```

### 7. Test

```bash
make test
```

## 📚 Available Commands

```bash
# Infrastructure
make up              # Start infrastructure (Docker Compose)
make down            # Stop infrastructure
make logs            # View logs

# Code Generation
make proto-gen       # Generate protobuf code
make sqlc-gen        # Generate SQL code
make gen             # Generate all code

# Dependencies
make init-deps       # Download all dependencies

# Build
make build           # Build all services
make build-gateway   # Build gateway only
make build-product   # Build product service only

# Test
make test            # Run all tests
make test-coverage   # Run tests with coverage

# Lint
make lint            # Run all linters

# Database
make db-migrate      # Run database migrations
make db-rollback     # Rollback migrations

# Docker
make docker-build     # Build Docker images
make docker-push     # Push Docker images

# Kubernetes
make k8s-apply       # Apply Kubernetes manifests
make k8s-delete      # Delete Kubernetes resources

# Clean
make clean           # Clean build artifacts
make clean-all       # Clean everything including generated code
```

## 🗺 Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                      API Gateway (Go)                     │
│              - Authentication & Authorization               │
│              - Rate Limiting & Circuit Breakers           │
│              - gRPC ↔ REST Translation                     │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐   ┌──────────────┐
│   Product    │    │    Order     │   │   Payment    │
│   Service    │    │   Service    │   │   Service    │
│     (Go)     │    │    (Go)      │   │    (Go)      │
├──────────────┤    ├──────────────┤   ├──────────────┤
│• Products    │    │• Orders      │   │• Payments    │
│• Categories  │    │• Cart        │   │• Konbini    │
│• Search      │    │• Checkout    │   │• Points      │
└──────────────┘    └──────────────┘   └──────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐   ┌──────────────┐
│   Inventory  │    │     User     │   │  Delivery    │
│   Service    │    │   Service    │   │   Service    │
│    (Rust)    │    │    (Go)      │   │    (Go)      │
├──────────────┤    ├──────────────┤   ├──────────────┤
│• Stock Mgmt  │    │• Auth        │   │• Same-day    │
│• Allocation  │    │• Profile     │   │• Tracking    │
│• Reservation │    │• Addresses   │   │• Logistics   │
└──────────────┘    └──────────────┘   └──────────────┘
                            │
                            ▼
                  ┌──────────────────┐
                  │  Analytics      │
                  │   Service      │
                  │   (Python)     │
                  ├──────────────────┤
                  │• Reporting     │
                  │• AI Insights   │
                  │• Batch Jobs    │
                  └──────────────────┘
```

### Data Flow

1. **Product Browsing** (Read-Heavy)
   ```
   Client → Gateway → Product Service → Redis Cache → PostgreSQL
   ```

2. **Order Placement** (Write-Heavy, ACID)
   ```
   Client → Gateway → Order Service
     ├── Lock Inventory (Inventory Service - Rust)
     ├── Create Order
     ├── Process Payment (Payment Service)
     ├── Deduct Points (User Service)
     └── Publish Event (Kafka)
   ```

3. **Konbini Payment Flow**
   ```
   Client → Gateway → Payment Service
     ├── Generate Payment Slip (PDF)
     ├── Send to User Email/Show in UI
     ├── Wait for Webhook from Payment Provider
     ├── Validate & Update Order Status
     └── Publish Payment Completed Event
   ```

## 🇯🇵 Japan-Specific Features

### Konbini Payments
- 7-Eleven (セブン-イレブン)
- Lawson (ローソン)
- FamilyMart (ファミリーマート)
- Ministop (ミニストップ)
- Seicomart (セイコーマート)

### Point System
- Multi-vendor point ecosystem
- Point redemption at checkout
- Point expiration management
- Cross-vendor point sharing

### Same-day Delivery
- Geospatial queries (PostGIS)
- Delivery slot management
- Real-time inventory check
- Tracking integration

## 📊 Observability

### Metrics
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- Port: `http://localhost:3000` (admin/admin)

### Tracing
- **Jaeger**: Distributed tracing
- Port: `http://localhost:16686`

### Logs
- Structured JSON logging with request IDs
- Centralized log aggregation

## 🧪 Testing

```bash
# Unit tests
make test

# Tests with coverage
make test-coverage

# Integration tests (requires running infrastructure)
make test-integration
```

## 🚢 Deployment

### Docker Compose (Local Development)
```bash
make up
make build
make docker-build
```

### Kubernetes (Production)
```bash
# Apply base manifests
make k8s-apply

# For specific environment
kubectl apply -k deploy/k8s/overlays/production
```

### Terraform (Infrastructure)
```bash
cd deploy/terraform
terraform init
terraform plan
terraform apply
```

## 🚀 Quick Start

**Get the platform running in 5 minutes!**

```bash
# Start all services (PostgreSQL, Redis, 7 microservices, Gateway)
make up

# Run integration tests
make test-integration

# Stop services
make down
```

📖 **See [QUICKSTART.md](QUICKSTART.md) for detailed setup instructions**

## 📖 Documentation

- [Quick Start Guide](QUICKSTART.md) - Get started in 5 minutes
- [Architecture Overview](docs/architecture/overview.md)
- [High-Level Design](docs/architecture/hld.md)
- [Low-Level Design](docs/architecture/lld.md)
- [API Documentation](docs/api/)
- [Deployment Guide](docs/deployment/)
- [Development Guide](docs/development/)
- [Runbooks](docs/runbooks/)

## 🤝 Contributing

This is a portfolio project demonstrating:
- System architecture skills
- Polyglot programming
- DevOps & infrastructure as code
- Japanese e-commerce domain knowledge
- Production-grade practices

## 📝 License

This project is licensed under the MIT License.

## 👨‍💻 Portfolio

Built as a demonstration of:
- **Senior Backend Engineer** skills
- **Japan-focused** e-commerce domain expertise
- **Polyglot** development (Go, Rust, Python)
- **Microservices** architecture
- **Kubernetes** orchestration
- **CI/CD** automation

## 🙏 Acknowledgments

Inspired by:
- [Saleor](https://saleor.io/)
- [Magento](https://magento.com/)
- [Rakuten](https://global.rakuten.com/)
- [PayPay](https://paypay.ne.jp/)
- [Buf](https://buf.build/)
- [sqlc](https://sqlc.dev/)
