package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DefaultTTL = 5 * time.Minute
	ShortTTL   = 1 * time.Minute
	LongTTL    = 1 * time.Hour
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		return err
	}

	if err := json.Unmarshal(val, dest); err != nil {
		return err
	}

	return nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func ProductCacheKey(productID string) string {
	return fmt.Sprintf("product:%s", productID)
}

func ProductListCacheKey(categoryID string, activeOnly bool) string {
	return fmt.Sprintf("products:category:%s:active:%v", categoryID, activeOnly)
}

func ProductSearchCacheKey(query string) string {
	return fmt.Sprintf("products:search:%s", query)
}

func NewRedisClient(redisURL string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
}
