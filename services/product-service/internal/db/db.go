package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxRetries     = 10
	initialBackoff = 1 * time.Second
	maxBackoff     = 10 * time.Second
)

func New(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	backoff := initialBackoff
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := pool.Ping(context.Background()); err == nil {
			return pool, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(backoff)
			if backoff < maxBackoff {
				backoff *= 2
			}
		}
	}

	return nil, fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, lastErr)
}
