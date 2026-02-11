FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./services/gateway/cmd/gateway

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

 COPY --from=builder /app/gateway .
 COPY --from=builder /app/docs/api/swagger.yaml /root/docs/api/swagger.yaml
 COPY --from=builder /app/docs/api/index.html /root/docs/api/index.html

 EXPOSE 8080

 CMD ["./gateway"]
