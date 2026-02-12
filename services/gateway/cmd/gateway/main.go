package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/afasari/shinkansen-commerce/services/gateway/internal/config"
	"github.com/afasari/shinkansen-commerce/services/gateway/internal/handler"
	"github.com/afasari/shinkansen-commerce/services/gateway/internal/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	logger *zap.Logger
	cfg    *config.Config
)

func main() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	productConn, err := grpc.NewClient(cfg.ProductServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial product service", zap.Error(err))
	}
	defer func() { _ = productConn.Close() }()

	orderConn, err := grpc.NewClient(cfg.OrderServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial order service", zap.Error(err))
	}
	defer func() { _ = orderConn.Close() }()

	userConn, err := grpc.NewClient(cfg.UserServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial user service", zap.Error(err))
	}
	defer func() { _ = userConn.Close() }()

	paymentConn, err := grpc.NewClient(cfg.PaymentServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial payment service", zap.Error(err))
	}
	defer func() { _ = paymentConn.Close() }()

	inventoryConn, err := grpc.NewClient(cfg.InventoryServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial inventory service", zap.Error(err))
	}
	defer func() { _ = inventoryConn.Close() }()

	deliveryConn, err := grpc.NewClient(cfg.DeliveryServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to dial delivery service", zap.Error(err))
	}
	defer func() { _ = deliveryConn.Close() }()

	mux := http.NewServeMux()

	if err := handler.RegisterHandlers(ctx, mux, productConn, orderConn, userConn, paymentConn, inventoryConn, deliveryConn); err != nil {
		logger.Fatal("Failed to register handlers", zap.Error(err))
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(http.Dir("/docs/api")))
	mux.Handle("/swagger/", swaggerHandler)
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})

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
		logger.Info("Starting HTTP gateway", zap.String("address", cfg.HTTPServerAddress))
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
