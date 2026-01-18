package config

import (
	"os"
)

type Config struct {
	GRPCServerAddress    string
	MetricsServerAddress string
	DatabaseURL          string
	RedisURL             string
}

func Load() (*Config, error) {
	grpcAddr := getEnv("GRPC_SERVER_ADDRESS", ":9091")
	metricsAddr := getEnv("METRICS_SERVER_ADDRESS", ":8091")
	dbURL := getEnv("DATABASE_URL", "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

	return &Config{
		GRPCServerAddress:    grpcAddr,
		MetricsServerAddress: metricsAddr,
		DatabaseURL:          dbURL,
		RedisURL:             redisURL,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
