package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPServerAddress           string
	ProductServiceGRPCAddress   string
	OrderServiceGRPCAddress     string
	UserServiceGRPCAddress      string
	PaymentServiceGRPCAddress   string
	InventoryServiceGRPCAddress string
	DeliveryServiceGRPCAddress  string
	JWTSecret                   string
}

func Load() (*Config, error) {
	httpAddr := getEnv("HTTP_SERVER_ADDRESS", ":8080")
	productAddr := getEnv("PRODUCT_SERVICE_GRPC_ADDRESS", "localhost:9091")
	orderAddr := getEnv("ORDER_SERVICE_GRPC_ADDRESS", "localhost:9092")
	userAddr := getEnv("USER_SERVICE_GRPC_ADDRESS", "localhost:9103")
	paymentAddr := getEnv("PAYMENT_SERVICE_GRPC_ADDRESS", "localhost:9104")
	inventoryAddr := getEnv("INVENTORY_SERVICE_GRPC_ADDRESS", "localhost:9105")
	deliveryAddr := getEnv("DELIVERY_SERVICE_GRPC_ADDRESS", "localhost:9106")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")

	if jwtSecret == "your-secret-key-change-in-production" {
		fmt.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	return &Config{
		HTTPServerAddress:           httpAddr,
		ProductServiceGRPCAddress:   productAddr,
		OrderServiceGRPCAddress:     orderAddr,
		UserServiceGRPCAddress:      userAddr,
		PaymentServiceGRPCAddress:   paymentAddr,
		InventoryServiceGRPCAddress: inventoryAddr,
		DeliveryServiceGRPCAddress:  deliveryAddr,
		JWTSecret:                   jwtSecret,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
