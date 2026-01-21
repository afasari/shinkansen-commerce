module github.com/shinkansen-commerce/shinkansen/services/order-service

go 1.24.0

toolchain go1.24.9

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.4
	github.com/redis/go-redis/v9 v9.3.1
	github.com/shinkansen-commerce/shinkansen/gen/proto/go v0.0.0
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.64.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260114163908-3f89685c29c3 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/shinkansen-commerce/shinkansen/gen/proto/go => ../../gen/proto/go
