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
COPY services/product-service/ services/product-service/

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./services/product-service/cmd/product-service

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/product-service .

EXPOSE 9091

CMD ["./product-service"]
