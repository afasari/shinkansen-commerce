# Build stage
FROM rust:1.86-slim AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    libpq-dev \
    protobuf-compiler \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY services/inventory-service/Cargo.toml ./services/inventory-service/
COPY services/inventory-service/Cargo.lock ./services/inventory-service/
COPY services/inventory-service/build.rs ./services/inventory-service/

COPY services/inventory-service/src ./services/inventory-service/src/
COPY services/inventory-service/migrations ./services/inventory-service/migrations/

COPY gen/proto/rust ./gen/proto/rust/
COPY proto ./proto/

WORKDIR /app/services/inventory-service
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    libpq5 \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/services/inventory-service/target/release/shinkansen-inventory-service .

EXPOSE 9105 8105

HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8105/health"]

CMD ["./shinkansen-inventory-service"]
