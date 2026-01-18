package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPServerAddress string
	GRPCServerAddress string
	JWTSecret         string
}

func Load() (*Config, error) {
	httpAddr := getEnv("HTTP_SERVER_ADDRESS", ":8080")
	grpcAddr := getEnv("GRPC_SERVER_ADDRESS", "localhost:9090")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")

	if jwtSecret == "your-secret-key-change-in-production" {
		fmt.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET environment variable in production!")
	}

	return &Config{
		HTTPServerAddress: httpAddr,
		GRPCServerAddress: grpcAddr,
		JWTSecret:         jwtSecret,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
