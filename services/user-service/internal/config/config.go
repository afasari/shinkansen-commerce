package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCServerAddress    string
	MetricsServerAddress string
	DatabaseURL          string
	RedisURL             string
	JWTSecret            string
	AccessTokenDuration  int
	RefreshTokenDuration int
}

func Load() (*Config, error) {
	return &Config{
		GRPCServerAddress:    getEnv("GRPC_SERVER_ADDRESS", ":9103"),
		MetricsServerAddress: getEnv("METRICS_SERVER_ADDRESS", ":8103"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:            getEnv("JWT_SECRET", "your-jwt-secret-change-in-production"),
		AccessTokenDuration:  getEnvInt("ACCESS_TOKEN_DURATION", 3600),
		RefreshTokenDuration: getEnvInt("REFRESH_TOKEN_DURATION", 86400),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
