FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work ./

COPY services/gateway/go.mod services/gateway/go.sum ./services/gateway/
COPY services/product-service/go.mod services/product-service/go.sum ./services/product-service/
COPY services/order-service/go.mod services/order-service/go.sum ./services/order-service/
COPY services/payment-service/go.mod services/payment-service/go.sum ./services/payment-service/
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY services/delivery-service/go.mod services/delivery-service/go.sum ./services/delivery-service/
COPY gen/proto/go/go.mod gen/proto/go/go.sum ./gen/proto/go/

RUN go mod download

COPY gen/proto/go/ gen/proto/go/
COPY services/gateway/ services/gateway/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./services/gateway/cmd/gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

RUN mkdir -p /docs/api

COPY --from=builder /app/services/gateway/docs/api/swagger.yaml /docs/api/swagger.yaml
COPY --from=builder /app/services/gateway/docs/api/index.html /docs/api/index.html

COPY --from=builder /app/gateway /root/gateway

EXPOSE 8080

CMD ["./gateway"]
