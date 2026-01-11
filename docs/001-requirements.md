# Project Shinkansen - Requirements Specification

---

Document Version: 1.0
Status: Draft
Author: Batiar Afas Rahmamulia
Date: 2026-01-11

---

1. Executive Summary

Project Shinkansen is a high-concurrency, distributed e-commerce platform designed to simulate the architectural challenges of major Japanese tech companies (e.g., Mercari, Rakuten, LINE).

The project aims to demonstrate expertise in building scalable microservices, handling high-throughput "Flash Sales" with strong consistency, and integrating a polyglot tech stack (Go, Rust, Python) optimized for specific performance needs.

2. Functional Requirements

2.1 Actors

    Buyer: End customer purchasing goods.
    Seller: Merchant managing inventory and orders.
    Admin: System operator monitoring health and data.

2.2 Client Interfaces

    Web Buyer (clients/web-buyer):
        Product browsing, search (Elasticsearch), and filtering.
        Shopping cart management.
        Flash Sale countdown UI.
        Checkout flow supporting Credit Card (Stripe) and Konbini (Convenience Store) payments.
    Web Admin (clients/web-admin):
        Dashboard for creating products and managing inventory.
        Real-time sales monitoring (Analytics).
        System health status.

2.3 Core Business Workflows

A. The Flash Sale (High Concurrency)

    Buyer selects a limited-item product.
    Buyer clicks "Buy" exactly when the sale starts.
    System must handle 10,000+ concurrent requests.
    Constraint: "Fail Fast." Users must receive an immediate "Success" or "Out of Stock" response. No waiting queues.

B. The Checkout & Inventory Locking

    Order Initiation: Order Service creates an OrderPending record.
    Reservation: Inventory Service (Rust) atomically decrements stock. If stock is 0, the transaction fails immediately.
    Payment:
        If Credit Card: Charge immediately.
        If Konbini: Generate payment slip. Stock remains reserved.
    Fulfillment: If payment fails or expires (timeout), the reserved stock is automatically returned to the pool.

3. Non-Functional Requirements (NFR)

3.1 Performance & Scalability

    Throughput: The Checkout API must sustain 10,000 Requests Per Second (RPS) during peak load.
    Latency: API Gateway response time (p95) must be < 100ms.
    Concurrency: System must prevent race conditions (overselling) under extreme load without using heavy database locks.

3.2 Reliability & Availability

    Availability: 99.9% uptime target.
    Fault Tolerance: If the Analytics service goes down, the Checkout flow must remain unaffected (Isolation of concerns).
    Data Consistency: Strong consistency for Inventory/Orders (ACID). Eventual consistency for Analytics/Search.

3.3 Tech Stack & Architecture

    Pattern: Domain-Driven Microservices.
    Communication: gRPC for internal service-to-service calls (performance), REST for external client communication.
    Languages:
        Golang: Primary backend (Gateway, Order, Product). Focus on concurrency and maintainability.
        Rust: Inventory Service. Focus on memory safety and zero-cost abstractions for locking.
        Python: Analytics Worker. Focus on data processing speed and ease of integration.

4. High-Level Architecture (Logical)

The system is organized by Business Domain (not language).

4.1 Backend Services (/services)

    Gateway Service (Go): Entry point. Auth (JWT), Rate Limiting, Request Routing, Idempotency.
    Product Service (Go): Catalog CRUD, Search indexing.
    Inventory Service (Rust): High-performance stock reservation and locking logic.
    Order Service (Go): Orchestrator. Manages order state machine (Pending -> Paid -> Shipped).
    Payment Service (Go): Integration with Stripe and Mock Konbini provider.
    Analytics Worker (Python): Consumes Kafka events (OrderCreated, PaymentSuccess) to update dashboards.

4.2 Data & Infrastructure

    Databases: PostgreSQL (Relational data), Redis (Cache/Session), Elasticsearch (Search).
    Messaging: Apache Kafka (Event-driven communication).
    Observability: Prometheus + Grafana (Metrics), OpenTelemetry (Distributed Tracing).

5. Scope Limitations (Out of Scope)

To ensure the project remains feasible for a portfolio while delivering value:

    AI/ML Recommendations: Not included in MVP.
    Complex Tax Logic: Will use a flat/simple tax calculation.
    Real Logistics: API calls to shipping carriers (Yamato/Sagawa) will be mocked, but the data structures will be realistic.
    Mobile Apps: Native iOS/Android are out of scope, but the architecture supports mobile via the Gateway API.

6. Success Metrics (Portfolio Goals)

The project will be considered a success if it demonstrates:

    Clean Architecture: Clear separation of concerns in the codebase.
    Load Handling: Evidence of passing a stress test (e.g., using k6) simulating 10k RPS.
    Polyglot Integration: Seamless communication between Go, Rust, and Python via gRPC and Kafka.
