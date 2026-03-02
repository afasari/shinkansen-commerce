package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/afasari/shinkansen-commerce/services/product-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/pkg/pgutil"
)

// CacheWarmer handles cache warming operations
type CacheWarmer struct {
	queries   db.Querier
	cache     cache.Cache
	redis     *redis.Client
	logger    *zap.Logger
	searchSvc *SearchService
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(
	queries db.Querier,
	cacheClient cache.Cache,
	redisClient *redis.Client,
	searchService *SearchService,
	logger *zap.Logger,
) *CacheWarmer {
	return &CacheWarmer{
		queries:   queries,
		cache:     cacheClient,
		redis:     redisClient,
		searchSvc: searchService,
		logger:    logger,
	}
}

// WarmProductCache warms the cache with popular products
func (w *CacheWarmer) WarmProductCache(ctx context.Context, limit int) error {
	w.logger.Info("Starting product cache warming", zap.Int("limit", limit))

	start := time.Now()

	// Get popular products (most viewed, most sold, etc.)
	products, err := w.queries.ListProducts(ctx, db.ListProductsParams{
		CategoryID: pgtype.UUID{},
		ActiveOnly: boolPtr(true),
		Offset:     int32Ptr(0),
		Limit:      int32Ptr(int32(limit)),
	})
	if err != nil {
		return fmt.Errorf("failed to list products: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent operations

	warmedCount := 0
	var mu sync.Mutex

	for _, product := range products {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire

		go func(p db.ListProductsRow) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release

			cacheKey := cache.ProductCacheKey(pgutil.FromPG(p.ID))
			if err := w.cache.Set(ctx, cacheKey, p, cache.LongTTL); err != nil {
				w.logger.Warn("Failed to warm cache for product",
					zap.String("product_id", pgutil.FromPG(p.ID)),
					zap.Error(err))
				return
			}

			mu.Lock()
			warmedCount++
			mu.Unlock()
		}(product)
	}

	wg.Wait()

	duration := time.Since(start)
	w.logger.Info("Product cache warming completed",
		zap.Int("total", len(products)),
		zap.Int("warmed", warmedCount),
		zap.Duration("duration", duration))

	return nil
}

// WarmCategoryCache warms the cache with products by category
func (w *CacheWarmer) WarmCategoryCache(ctx context.Context, categoryID string) error {
	w.logger.Info("Starting category cache warming", zap.String("category_id", categoryID))

	products, err := w.queries.ListProducts(ctx, db.ListProductsParams{
		CategoryID: pgutil.ToPG(uuid.MustParse(categoryID)),
		ActiveOnly: boolPtr(true),
		Offset:     int32Ptr(0),
		Limit:      int32Ptr(100),
	})
	if err != nil {
		return fmt.Errorf("failed to list products by category: %w", err)
	}

	for _, product := range products {
		cacheKey := cache.ProductCacheKey(pgutil.FromPG(product.ID))
		if err := w.cache.Set(ctx, cacheKey, product, cache.LongTTL); err != nil {
			w.logger.Warn("Failed to warm cache for product",
				zap.String("product_id", pgutil.FromPG(product.ID)),
				zap.Error(err))
		}
	}

	w.logger.Info("Category cache warming completed",
		zap.String("category_id", categoryID),
		zap.Int("count", len(products)))

	return nil
}

// WarmSearchCache warms the cache with popular search queries
func (w *CacheWarmer) WarmSearchCache(ctx context.Context) error {
	return w.searchSvc.WarmCache(ctx)
}

// WarmAll performs full cache warming
func (w *CacheWarmer) WarmAll(ctx context.Context) error {
	w.logger.Info("Starting full cache warming")

	start := time.Now()

	// Warm popular products
	if err := w.WarmProductCache(ctx, 100); err != nil {
		w.logger.Error("Failed to warm product cache", zap.Error(err))
	}

	// Warm search cache
	if err := w.WarmSearchCache(ctx); err != nil {
		w.logger.Error("Failed to warm search cache", zap.Error(err))
	}

	duration := time.Since(start)
	w.logger.Info("Full cache warming completed", zap.Duration("duration", duration))

	return nil
}

// StartPeriodicWarming starts periodic cache warming in the background
func (w *CacheWarmer) StartPeriodicWarming(ctx context.Context, interval time.Duration) {
	w.logger.Info("Starting periodic cache warming", zap.Duration("interval", interval))

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := w.WarmAll(ctx); err != nil {
					w.logger.Error("Periodic cache warming failed", zap.Error(err))
				}
			}
		}
	}()
}

// WarmOnDemand warms cache for specific products
func (w *CacheWarmer) WarmOnDemand(ctx context.Context, productIDs []string) error {
	w.logger.Info("Starting on-demand cache warming", zap.Int("count", len(productIDs)))

	for _, productID := range productIDs {
		productUUID := uuid.MustParse(productID)
		product, err := w.queries.GetProduct(ctx, pgutil.ToPG(productUUID))
		if err != nil {
			w.logger.Warn("Failed to get product for cache warming",
				zap.String("product_id", productID),
				zap.Error(err))
			continue
		}

		cacheKey := cache.ProductCacheKey(productID)
		if err := w.cache.Set(ctx, cacheKey, product, cache.DefaultTTL); err != nil {
			w.logger.Warn("Failed to warm cache for product",
				zap.String("product_id", productID),
				zap.Error(err))
		}
	}

	return nil
}

// InvalidateProduct invalidates cache for a specific product
func (w *CacheWarmer) InvalidateProduct(ctx context.Context, productID string) error {
	cacheKey := cache.ProductCacheKey(productID)
	if err := w.cache.Delete(ctx, cacheKey); err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	// Also invalidate from Redis if present
	if w.redis != nil {
		pattern := fmt.Sprintf("product:*:%s:*", productID)
		iter := w.redis.Scan(ctx, 0, pattern, 100).Iterator()
		for iter.Next(ctx) {
			w.redis.Del(ctx, iter.Val())
		}
	}

	w.logger.Debug("Product cache invalidated", zap.String("product_id", productID))
	return nil
}

// InvalidateCategory invalidates all product caches in a category
func (w *CacheWarmer) InvalidateCategory(ctx context.Context, categoryID string) error {
	pattern := fmt.Sprintf("product:category:%s:*", categoryID)
	if w.redis != nil {
		iter := w.redis.Scan(ctx, 0, pattern, 100).Iterator()
		keys := []string{}
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if len(keys) > 0 {
			w.redis.Del(ctx, keys...)
		}
	}

	w.logger.Debug("Category cache invalidated", zap.String("category_id", categoryID))
	return nil
}

// GetCacheStats returns cache statistics
func (w *CacheWarmer) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	var stats CacheStats

	// Count product cache keys
	if w.redis != nil {
		iter := w.redis.Scan(ctx, 0, "product:*", 100).Iterator()
		for iter.Next(ctx) {
			stats.ProductKeys++
		}

		// Count search cache keys
		iter = w.redis.Scan(ctx, 0, "search:*", 100).Iterator()
		for iter.Next(ctx) {
			stats.SearchKeys++
		}
	}

	return &stats, nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	ProductKeys int64
	SearchKeys  int64
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}
