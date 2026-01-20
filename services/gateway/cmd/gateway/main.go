package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/config"
	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/handler"
	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	logger *zap.Logger
	cfg    *config.Config
)

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.NewClient(cfg.GRPCServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial gRPC server", zap.Error(err))
	}
	defer conn.Close()

	mux := http.NewServeMux()

	if err := handler.RegisterHandlers(ctx, mux, conn); err != nil {
		logger.Fatal("Failed to register handlers", zap.Error(err))
	}

	chain := middleware.Chain(mux,
		middleware.CORS(),
		middleware.RequestID(),
		middleware.Logging(logger),
		middleware.Auth(cfg.JWTSecret),
	)

	srv := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: chain,
	}

	go func() {
		logger.Info("Starting HTTP gateway",
			zap.String("address", cfg.HTTPServerAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP gateway failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down HTTP gateway...")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown gracefully", zap.Error(err))
	}
	logger.Info("HTTP gateway stopped")
}
