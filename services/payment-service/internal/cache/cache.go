package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultTTL = 15 * time.Minute
)

func PaymentCacheKey(paymentID string) string {
	return "payment:" + paymentID
}

func PaymentsByOrderCacheKey(orderID string) string {
	return "order:payments:" + orderID
}

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisClient(url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		opts = &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}
	return redis.NewClient(opts)
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}
