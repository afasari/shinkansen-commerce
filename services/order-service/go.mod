module github.com/shinkansen-commerce/shinkansen/services/order-service

go 1.21

require (
    github.com/jackc/pgx/v5 v5.5.4
    github.com/redis/go-redis/v9 v9.3.1
    go.uber.org/zap v1.26.0
    google.golang.org/grpc v1.59.0
    google.golang.org/protobuf v1.31.0
)

require (
    github.com/cespare/xxhash/v2 v2.2.0
    github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
    github.com/jackc/pgpassfile v1.0.0
    github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a
    go.uber.org/multierr v1.11.0
    golang.org/x/crypto v0.16.0
    golang.org/x/text v0.14.0
)
