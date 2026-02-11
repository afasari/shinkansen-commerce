# Architecture Overview

## High-Level Architecture

The platform follows a microservices architecture with clear service boundaries based on domain-driven design principles.

```mermaid
graph TB
    subgraph "Client Layer"
        A[Web App]
        B[Mobile App]
        C[Admin Dashboard]
    end
    
    subgraph "API Gateway"
        D[Gateway<br/>:8080]
    end
    
    subgraph "Product Domain"
        E[Product Service<br/>:9091]
    end
    
    subgraph "Order Domain"
        F[Order Service<br/>:9092]
    end
    
    subgraph "User Domain"
        G[User Service<br/>:9103]
    end
    
    subgraph "Payment Domain"
        H[Payment Service<br/>:9104]
    end
    
    subgraph "Inventory Domain"
        I[Inventory Service<br/>:9105]
    end
    
    subgraph "Delivery Domain"
        J[Delivery Service<br/>:9106]
    end
    
    subgraph "Data Layer"
        K[PostgreSQL<br/>:5432]
        L[Redis<br/>:6379]
        M[Kafka<br/>:9092]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E
    D --> F
    D --> G
    D --> H
    D --> I
    D --> J
    E --> K
    E --> L
    F --> K
    F --> L
    F --> M
    G --> K
    G --> L
    H --> K
    H --> M
    I --> K
    I --> L
    J --> K
```

## Service Boundaries

### Product Domain
**Responsibility**: Product catalog management

**Capabilities**:
- Product CRUD operations
- Category management
- Product variants (size, color)
- Product search
- Stock display (read-only)

**Data**: `catalog` schema

### Order Domain
**Responsibility**: Order lifecycle management

**Capabilities**:
- Order creation
- Order status transitions
- Order history
- Point redemption
- Delivery slot reservation

**Data**: `orders` schema

### User Domain
**Responsibility**: User identity and profile

**Capabilities**:
- User registration/login
- JWT token generation
- Profile management
- Address management

**Data**: `users` schema

### Payment Domain
**Responsibility**: Payment processing

**Capabilities**:
- Multiple payment methods
- Konbini payment slips
- Point integration
- Refunds

**Data**: `payments` schema

### Inventory Domain
**Responsibility**: Stock management

**Capabilities**:
- Stock level tracking
- Stock reservation (optimistic locking)
- Stock movement logging
- Inventory updates

**Data**: `inventory` schema
**Note**: Built in Rust for performance

### Delivery Domain
**Responsibility**: Delivery logistics

**Capabilities**:
- Delivery zones (PostGIS)
- Delivery slot management
- Shipment tracking
- Same-day delivery

**Data**: `delivery` schema

## Communication Patterns

### Synchronous: gRPC
Used for:
- Request/response operations
- Cross-service queries
- Real-time validation

**Example**:
```
Gateway → Order Service (create order)
Order Service → Inventory Service (reserve stock)
Inventory Service → PostgreSQL (update stock)
```

### Asynchronous: Kafka
Used for:
- Event publishing
- Decoupled processing
- Analytics integration

**Example**:
```
Order Service → Kafka (order.created event)
Analytics Worker ← Kafka (consume event)
Analytics Worker → PostgreSQL (update analytics)
```

## Data Flow Examples

### Order Creation Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant O as Order Service
    participant I as Inventory Service
    participant P as Payment Service
    participant DB as PostgreSQL
    participant R as Redis
    
    C->>G: POST /v1/orders
    G->>O: CreateOrder (gRPC)
    O->>I: ReserveStock (gRPC)
    I->>DB: UPDATE stock (lock)
    I-->>O: ReserveStockResponse
    O->>DB: INSERT order
    O->>DB: INSERT order_items
    DB-->>O: Order created
    O->>P: ValidatePaymentMethod (gRPC)
    P-->>O: ValidationResponse
    O->>R: Cache order (TTL: 5min)
    O->>M: Publish order.created
    O-->>G: CreateOrderResponse
    G-->>C: 201 Created
```

### Product Browsing (Cached)

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant P as Product Service
    participant R as Redis
    participant DB as PostgreSQL
    
    C->>G: GET /v1/products
    G->>P: ListProducts (gRPC)
    P->>R: GET cache:products:page:1
    alt Cache Hit
        R-->>P: Cached data
    else Cache Miss
        P->>DB: SELECT products
        DB-->>P: Product data
        P->>R: SET cache:products:page:1 (TTL: 5min)
    end
    P-->>G: ListProductsResponse
    G-->>C: 200 OK
```

## Caching Strategy

| Data | Cache Location | TTL | Invalidation |
|------|---------------|-----|--------------|
| Product listings | Redis | 5 min | Product update/delete |
| Product details | Redis | 30 min | Product update |
| User sessions | Redis | 24 hr | Logout/timeout |
| Order details | Redis | 5 min | Order status update |

## Scalability Considerations

### Horizontal Scaling
- Stateless services (Gateway, Product, Order)
- Session store in Redis
- Load balancer: Kubernetes Service

### Database Scaling
- Read replicas for high-traffic services
- Connection pooling
- Query optimization

### Cache Scaling
- Redis Cluster for production
- Partitioning by domain
- Eviction policies

## Security

### Authentication
- JWT tokens (1 hr access, 24 hr refresh)
- HTTP Bearer token header
- Token verification in Gateway middleware

### Authorization
- Role-based access control (RBAC)
- Service-to-service: mutual TLS (future)
- Admin endpoints: additional role checks

### Network Security
- Internal gRPC: TLS (production)
- Gateway: HTTPS with Let's Encrypt
- API Gateway: Only public endpoint
