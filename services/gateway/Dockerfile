FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gen/ gen/
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./services/gateway/cmd/gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root

# Create docs directory structure
RUN mkdir -p /docs/api

# Copy swagger files
COPY --from=builder /app/services/gateway/docs/api/swagger.yaml /docs/api/swagger.yaml
COPY --from=builder /app/services/gateway/docs/api/index.html /docs/api/index.html

COPY --from=builder /app/gateway /root/gateway

EXPOSE 8080

CMD ["./gateway"]
