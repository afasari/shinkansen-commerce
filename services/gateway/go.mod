module github.com/afasari/shinkansen-commerce/services/gateway

go 1.24.0

toolchain go1.24.9

require (
	github.com/afasari/shinkansen-commerce/gen/proto/go v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/google/uuid v1.6.0
	github.com/redis/go-redis/v9 v9.17.2
	go.uber.org/zap v1.26.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require github.com/swaggo/swag v1.16.6

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/afasari/shinkansen-commerce/gen/proto/go => ../../gen/proto/go
