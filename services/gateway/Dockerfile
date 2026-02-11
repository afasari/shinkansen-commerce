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
      - name: Copy swagger files
        run: |
          mkdir -p /root/docs/api
          cp ./docs/api/swagger.yaml /root/docs/api/

      - name: Update main.go docs path
        run: |
          sed -i 's|http.FileServer(http.Dir("./docs/api"))|http.FileServer(http.Dir("/docs/api"))|g' ./cmd/gateway/main.go

  EXPOSE 8080

 EXPOSE 8080

 CMD ["./gateway"]
