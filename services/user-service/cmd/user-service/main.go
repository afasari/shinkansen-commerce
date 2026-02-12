package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	userv1 "github.com/afasari/shinkansen-commerce/gen/proto/go/user"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/config"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/user-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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
	defer func() { _ = logger.Sync() }()

	ctx := context.Background()
	dbpool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() { _ = dbpool.Close() }()

	queries := db.NewQueries(dbpool)
	redisClient := cache.NewRedisClient(cfg.RedisURL)
	cacheClient := cache.NewRedisCache(redisClient)
	userService := service.NewUserService(queries, cacheClient, logger, cfg)

	server := grpc.NewServer()
	userv1.RegisterUserServiceServer(server, userService)
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
