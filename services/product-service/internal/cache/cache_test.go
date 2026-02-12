package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedis(t *testing.T) (*RedisCache, func()) {
	s := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cache := NewRedisCache(client)

	cleanup := func() {
		_ = client.Close()
		s.Close()
	}

	return cache, cleanup
}

func TestRedisCache_Get(t *testing.T) {
	ctx := context.Background()
	cache, cleanup := setupRedis(t)
	defer cleanup()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name      string
		setup     func() error
		key       string
		dest      interface{}
		wantErr   bool
		errMsg    string
		wantValue interface{}
	}{
		{
			name: "success - cache hit",
			setup: func() error {
				data := TestStruct{Name: "test", Value: 42}
				return cache.Set(ctx, "test-key", data, time.Minute)
			},
			key:       "test-key",
			dest:      &TestStruct{},
			wantErr:   false,
			wantValue: &TestStruct{Name: "test", Value: 42},
		},
		{
			name: "cache miss",
			setup: func() error {
				return nil
			},
			key:     "nonexistent-key",
			dest:    &TestStruct{},
			wantErr: true,
			errMsg:  "cache miss",
		},
		{
			name: "invalid JSON in cache",
			setup: func() error {
				_ = cache.client.Set(ctx, "invalid-key", "invalid json", time.Minute)
				return nil
			},
			key:     "invalid-key",
			dest:    &TestStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			err := cache.Get(ctx, tt.key, tt.dest)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, tt.dest)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	ctx := context.Background()
	cache, cleanup := setupRedis(t)
	defer cleanup()

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name    string
		key     string
		value   interface{}
		ttl     time.Duration
		wantErr bool
	}{
		{
			name:  "success - set string",
			key:   "string-key",
			value: "test value",
			ttl:   time.Minute,
		},
		{
			name:  "success - set struct",
			key:   "struct-key",
			value: TestStruct{Name: "test", Value: 42},
			ttl:   time.Minute,
		},
		{
			name:  "success - set number",
			key:   "number-key",
			value: 123,
			ttl:   time.Minute,
		},
		{
			name:  "success - set with short TTL",
			key:   "short-ttl-key",
			value: "test",
			ttl:   ShortTTL,
		},
		{
			name:  "success - set with long TTL",
			key:   "long-ttl-key",
			value: "test",
			ttl:   LongTTL,
		},
		{
			name:  "success - set with default TTL",
			key:   "default-ttl-key",
			value: "test",
			ttl:   DefaultTTL,
		},
		{
			name:  "success - set with zero TTL",
			key:   "zero-ttl-key",
			value: "test",
			ttl:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, tt.key, tt.value, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var dest interface{}
				if _, ok := tt.value.(string); ok {
					var s string
					dest = &s
				} else if _, ok := tt.value.(TestStruct); ok {
					dest = &TestStruct{}
				} else if _, ok := tt.value.(int); ok {
					var i int
					dest = &i
				}

				err := cache.Get(ctx, tt.key, dest)
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache, cleanup := setupRedis(t)
	defer cleanup()

	tests := []struct {
		name    string
		setup   func() error
		key     string
		wantErr bool
	}{
		{
			name: "success - delete existing key",
			setup: func() error {
				return cache.Set(ctx, "delete-me", "value", time.Minute)
			},
			key:     "delete-me",
			wantErr: false,
		},
		{
			name: "success - delete non-existing key",
			setup: func() error {
				return nil
			},
			key:     "nonexistent-key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			err := cache.Delete(ctx, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var dest string
				err := cache.Get(ctx, tt.key, &dest)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cache miss")
			}
		})
	}
}

func TestRedisCache_Integration(t *testing.T) {
	ctx := context.Background()
	cache, cleanup := setupRedis(t)
	defer cleanup()

	t.Run("set and get workflow", func(t *testing.T) {
		type Product struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}

		product := Product{ID: "prod-1", Name: "Test Product", Price: 99.99}
		key := "product:prod-1"

		err := cache.Set(ctx, key, product, DefaultTTL)
		require.NoError(t, err)

		var retrieved Product
		err = cache.Get(ctx, key, &retrieved)
		require.NoError(t, err)
		assert.Equal(t, product.ID, retrieved.ID)
		assert.Equal(t, product.Name, retrieved.Name)
		assert.Equal(t, product.Price, retrieved.Price)
	})

	t.Run("delete prevents get", func(t *testing.T) {
		err := cache.Set(ctx, "temp-key", "temp-value", DefaultTTL)
		require.NoError(t, err)

		err = cache.Delete(ctx, "temp-key")
		require.NoError(t, err)

		var result string
		err = cache.Get(ctx, "temp-key", &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache miss")
	})

	t.Run("overwrite existing key", func(t *testing.T) {
		key := "overwrite-key"

		err := cache.Set(ctx, key, "first-value", DefaultTTL)
		require.NoError(t, err)

		var first string
		err = cache.Get(ctx, key, &first)
		require.NoError(t, err)
		assert.Equal(t, "first-value", first)

		err = cache.Set(ctx, key, "second-value", DefaultTTL)
		require.NoError(t, err)

		var second string
		err = cache.Get(ctx, key, &second)
		require.NoError(t, err)
		assert.Equal(t, "second-value", second)
	})
}

func TestProductCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		want      string
	}{
		{
			name:      "valid UUID",
			productID: "123e4567-e89b-12d3-a456-426614174000",
			want:      "product:123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:      "simple string",
			productID: "prod-123",
			want:      "product:prod-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProductCacheKey(tt.productID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProductListCacheKey(t *testing.T) {
	tests := []struct {
		name       string
		categoryID string
		activeOnly bool
		want       string
	}{
		{
			name:       "with category and active",
			categoryID: "cat-123",
			activeOnly: true,
			want:       "products:category:cat-123:active:true",
		},
		{
			name:       "with category and inactive",
			categoryID: "cat-456",
			activeOnly: false,
			want:       "products:category:cat-456:active:false",
		},
		{
			name:       "empty category",
			categoryID: "",
			activeOnly: true,
			want:       "products:category::active:true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProductListCacheKey(tt.categoryID, tt.activeOnly)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewRedisClient(t *testing.T) {
	tests := []struct {
		name     string
		redisURL string
	}{
		{
			name:     "valid localhost URL",
			redisURL: "localhost:6379",
		},
		{
			name:     "valid remote URL",
			redisURL: "redis.example.com:6379",
		},
		{
			name:     "valid URL with IP",
			redisURL: "192.168.1.100:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewRedisClient(tt.redisURL)
			assert.NotNil(t, client)
		})
	}
}

func TestProductSearchCacheKey(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "simple query",
			query: "test",
			want:  "products:search:test",
		},
		{
			name:  "query with spaces",
			query: "test product",
			want:  "products:search:test product",
		},
		{
			name:  "query with special chars",
			query: "test-product!",
			want:  "products:search:test-product!",
		},
		{
			name:  "empty query",
			query: "",
			want:  "products:search:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProductSearchCacheKey(tt.query)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRedisCache_Get_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cache, cleanup := setupRedis(t)
	defer cleanup()

	cancel()

	var dest string
	err := cache.Get(ctx, "test-key", &dest)
	assert.Error(t, err)
}

func TestRedisCache_Set_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cache, cleanup := setupRedis(t)
	defer cleanup()

	cancel()

	err := cache.Set(ctx, "test-key", "value", time.Minute)
	assert.Error(t, err)
}

func TestRedisCache_Delete_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cache, cleanup := setupRedis(t)
	defer cleanup()

	cancel()

	err := cache.Delete(ctx, "test-key")
	assert.Error(t, err)
}
