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
	return &Config{
		GRPCServerAddress:    getEnv("GRPC_SERVER_ADDRESS", ":9104"),
		MetricsServerAddress: getEnv("METRICS_SERVER_ADDRESS", ":8104"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
