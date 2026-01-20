package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/order"
	productpb "github.com/shinkansen-commerce/shinkansen/gen/proto/go/product"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/cache"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/config"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/db"
	"github.com/shinkansen-commerce/shinkansen/services/order-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	dbpool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbpool.Close()

	queries := db.NewQueries(dbpool)
	redisClient := cache.NewRedisClient(cfg.RedisURL)
	cacheClient := cache.NewRedisCache(redisClient)

	conn, err := grpc.NewClient(cfg.ProductServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial product service", zap.Error(err))
	}
	defer conn.Close()

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
		w.Write([]byte("OK"))
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
