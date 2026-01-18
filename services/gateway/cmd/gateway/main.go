package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/config"
	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/handler"
	"github.com/shinkansen-commerce/shinkansen/services/gateway/internal/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.GRPCServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Fatal("Failed to dial gRPC server", zap.Error(err))
	}
	defer conn.Close()

	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if key == "Authorization" {
				return key, true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	)

	if err := handler.RegisterHandlers(ctx, gwmux, conn); err != nil {
		logger.Fatal("Failed to register handlers", zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.Handle("/", middleware.Chain(
		gwmux,
		middleware.CORS(),
		middleware.RequestID(),
		middleware.Logging(logger),
		middleware.Auth(cfg.JWTSecret),
	))

	srv := &http.Server{
		Addr:    cfg.HTTPServerAddress,
		Handler: mux,
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
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown gracefully", zap.Error(err))
	}
	logger.Info("HTTP gateway stopped")
}
