module github.com/afasari/shinkansen-commerce/services/gateway

go 1.24.0

toolchain go1.24.9

require (
	github.com/afasari/shinkansen-commerce/gen/proto/go v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.17.2
	go.uber.org/zap v1.26.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260114163908-3f89685c29c3 // indirect
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240318140521-94a12d6c2237 // indirect
)

replace github.com/afasari/shinkansen-commerce/gen/proto/go => ../../gen/proto/go
