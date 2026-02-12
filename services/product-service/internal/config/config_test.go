package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() func()
		want    *Config
		wantErr bool
	}{
		{
			name: "success with defaults",
			setup: func() func() {
				return func() {}
			},
			want: &Config{
				GRPCServerAddress:    ":9091",
				MetricsServerAddress: ":8091",
				DatabaseURL:          "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable",
				RedisURL:             "redis://localhost:6379",
			},
			wantErr: false,
		},
		{
			name: "success with env vars",
			setup: func() func() {
				_ = os.Setenv("GRPC_SERVER_ADDRESS", ":50051")
				_ = os.Setenv("METRICS_SERVER_ADDRESS", ":8080")
				_ = os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db?sslmode=require")
				_ = os.Setenv("REDIS_URL", "redis://redis.example.com:6380")

				return func() {
					_ = os.Unsetenv("GRPC_SERVER_ADDRESS")
					_ = os.Unsetenv("METRICS_SERVER_ADDRESS")
					_ = os.Unsetenv("DATABASE_URL")
					_ = os.Unsetenv("REDIS_URL")
				}
			},
			want: &Config{
				GRPCServerAddress:    ":50051",
				MetricsServerAddress: ":8080",
				DatabaseURL:          "postgres://user:pass@host:5432/db?sslmode=require",
				RedisURL:             "redis://redis.example.com:6380",
			},
			wantErr: false,
		},
		{
			name: "success with partial env vars",
			setup: func() func() {
				_ = os.Setenv("GRPC_SERVER_ADDRESS", ":3000")
				_ = os.Setenv("REDIS_URL", "redis://custom:6379")

				return func() {
					_ = os.Unsetenv("GRPC_SERVER_ADDRESS")
					_ = os.Unsetenv("REDIS_URL")
				}
			},
			want: &Config{
				GRPCServerAddress:    ":3000",
				MetricsServerAddress: ":8091",
				DatabaseURL:          "postgres://shinkansen:shinkansen_dev_password@localhost:5432/shinkansen?sslmode=disable",
				RedisURL:             "redis://custom:6379",
			},
			wantErr: false,
		},
		{
			name: "success with empty env vars",
			setup: func() func() {
				_ = os.Setenv("GRPC_SERVER_ADDRESS", "")
				_ = os.Setenv("DATABASE_URL", "")

				return func() {
					_ = os.Unsetenv("GRPC_SERVER_ADDRESS")
					_ = os.Unsetenv("DATABASE_URL")
				}
			},
			want: &Config{
				GRPCServerAddress:    "",
				MetricsServerAddress: ":8091",
				DatabaseURL:          "",
				RedisURL:             "redis://localhost:6379",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			got, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		setup         func() func()
		expectedValue string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			setup: func() func() {
				_ = os.Setenv("TEST_VAR", "value")
				return func() {
					_ = os.Unsetenv("TEST_VAR")
				}
			},
			expectedValue: "value",
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "fallback",
			setup: func() func() {
				return func() {}
			},
			expectedValue: "fallback",
		},
		{
			name:         "env var is empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setup: func() func() {
				_ = os.Setenv("EMPTY_VAR", "")
				return func() {
					_ = os.Unsetenv("EMPTY_VAR")
				}
			},
			expectedValue: "",
		},
		{
			name:         "env var with special characters",
			key:          "SPECIAL_VAR",
			defaultValue: "default",
			setup: func() func() {
				_ = os.Setenv("SPECIAL_VAR", "value@#$%^&*()")
				return func() {
					_ = os.Unsetenv("SPECIAL_VAR")
				}
			},
			expectedValue: "value@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			got := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, got)
		})
	}
}
