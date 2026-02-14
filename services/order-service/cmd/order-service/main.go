package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	orderpb "github.com/afasari/shinkansen-commerce/gen/proto/go/order"
	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/config"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/order-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbpool.Close()

	queries := db.New(dbpool)
	redisClient := cache.NewRedisClient(cfg.RedisURL)
	cacheClient := cache.NewRedisCache(redisClient)

	conn, err := grpc.NewClient(cfg.ProductServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial product service", zap.Error(err))
	}
	defer func() { _ = conn.Close() }()

	productClient := productpb.NewProductServiceClient(conn)

	orderService := service.NewOrderService(queries, productClient, cacheClient, logger)

	server := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(server, orderService)
	reflection.Register(server)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("Starting gRPC server", zap.String("address", cfg.GRPCServerAddress))
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	go func() {
		logger.Info("Starting metrics server", zap.String("address", cfg.MetricsServerAddress))
		if err := http.ListenAndServe(cfg.MetricsServerAddress, metricsMux); err != nil {
			logger.Error("Metrics server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gRPC server...")
	server.GracefulStop()
	logger.Info("gRPC server stopped")
}
