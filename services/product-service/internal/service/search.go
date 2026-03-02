package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	productpb "github.com/afasari/shinkansen-commerce/gen/proto/go/product"
	sharedpb "github.com/afasari/shinkansen-commerce/gen/proto/go/shared"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/cache"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/db"
	"github.com/afasari/shinkansen-commerce/services/product-service/internal/pkg/pgutil"
)

// SearchService handles advanced product search functionality
type SearchService struct {
	queries    db.Querier
	cache      cache.Cache
	redis      *redis.Client
	logger     *zap.Logger
}

// NewSearchService creates a new search service
func NewSearchService(
	queries db.Querier,
	cacheClient cache.Cache,
	redisClient *redis.Client,
	logger *zap.Logger,
) *SearchService {
	return &SearchService{
		queries: queries,
		cache:   cacheClient,
		redis:   redisClient,
		logger:  logger,
	}
}

// SearchRequest represents an advanced search request
type SearchRequest struct {
	Query              string
	CategoryID         string
	MinPrice           *int64
	MaxPrice           *int64
	InStockOnly        bool
	Attributes         map[string]string
	SortBy             string // "relevance", "price_asc", "price_desc", "name", "newest"
	SortOrder          string // "asc", "desc"
	FuzzyMatch         bool
	Page               int32
	Limit              int32
}

// SearchResult represents a search result with metadata
type SearchResult struct {
	Products     []*productpb.Product
	TotalCount   int64
	Page         int32
	Limit        int32
	Query        string
	Duration     time.Duration
	DidYouMean   []string
	Suggestions  []string
}

// Search performs an advanced product search
func (s *SearchService) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	start := time.Now()

	// Normalize query
	normalizedQuery := s.normalizeQuery(req.Query)
	fuzzyQuery := normalizedQuery

	// Track search query (async)
	go s.trackSearch(context.Background(), normalizedQuery, req.Page)

	// Check cache for search results
	cacheKey := s.searchCacheKey(req)
	var cached db.SearchProductsRow
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		s.logger.Debug("Search cache hit", zap.String("query", normalizedQuery))
		return s.buildSearchResult([]*db.SearchProductsRow{&cached}, req, time.Since(start)), nil
	}

	// Perform search
	var products []db.SearchProductsRow
	var err error

	if req.FuzzyMatch && len(normalizedQuery) >= 3 {
		// Use fuzzy search
		products, err = s.fuzzySearch(ctx, fuzzyQuery, req)
	} else {
		// Use exact search
		products, err = s.exactSearch(ctx, normalizedQuery, req)
	}

	if err != nil {
		s.logger.Error("Search failed", zap.String("query", normalizedQuery), zap.Error(err))
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// If no results and fuzzy matching is enabled, try fuzzy suggestions
	var didYouMean []string
	if len(products) == 0 && req.FuzzyMatch {
		didYouMean = s.getSpellingSuggestions(normalizedQuery)
	}

	// Get search suggestions based on query
	suggestions := s.getSearchSuggestions(normalizedQuery)

	// Cache results
	if len(products) > 0 {
		// Cache first product as representative
		if err := s.cache.Set(ctx, cacheKey, products[0], cache.ShortTTL); err != nil {
			s.logger.Warn("Failed to cache search results", zap.Error(err))
		}
	}

	return s.buildSearchResultFromRows(products, req, time.Since(start), didYouMean, suggestions), nil
}

// normalizeQuery normalizes a search query
func (s *SearchService) normalizeQuery(query string) string {
	// Convert to lowercase
	query = strings.ToLower(query)

	// Remove extra whitespace
	query = strings.Join(strings.Fields(query), " ")

	// Remove special characters (keep alphanumeric and Japanese)
	var result strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Hiragana, r) ||
			unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Han, r) || r == ' ' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// exactSearch performs exact text search
func (s *SearchService) exactSearch(ctx context.Context, query string, req *SearchRequest) ([]db.SearchProductsRow, error) {
	s.logger.Info("Performing exact search", zap.String("query", query))

	offset := (req.Page - 1) * req.Limit

	// Note: The SearchProductsFuzzy function doesn't support offset/limit in params
	// These should be handled by the SQL query itself
	interfaceResults, err := s.queries.SearchProductsFuzzy(ctx, db.SearchProductsFuzzyParams{
		SearchQuery:    &query,
		CategoryFilter: pgUUIDFromString(req.CategoryID),
		MinPrice:       req.MinPrice,
		MaxPrice:       req.MaxPrice,
		StockOnly:      &req.InStockOnly,
		FuzzyThreshold: nil, // Use default threshold
	})
	if err != nil {
		return nil, err
	}

	// Convert []interface{} to []db.SearchProductsRow
	results := make([]db.SearchProductsRow, 0, len(interfaceResults))
	for _, item := range interfaceResults {
		if row, ok := item.(db.SearchProductsRow); ok {
			results = append(results, row)
		}
	}

	// Apply pagination manually
	start := int(offset)
	end := int(offset + req.Limit)
	if start > len(results) {
		return []db.SearchProductsRow{}, nil
	}
	if end > len(results) {
		end = len(results)
	}
	return results[start:end], nil
}

// fuzzySearch performs fuzzy text search with typos tolerance
func (s *SearchService) fuzzySearch(ctx context.Context, query string, req *SearchRequest) ([]db.SearchProductsRow, error) {
	s.logger.Info("Performing fuzzy search", zap.String("query", query))

	// Generate trigrams for fuzzy matching
	trigrams := s.generateTrigrams(query)

	// Build search query with trigrams
	fuzzyQuery := s.buildFuzzyQuery(trigrams)

	offset := (req.Page - 1) * req.Limit

	// Note: The SearchProductsFuzzy function doesn't support offset/limit in params
	interfaceResults, err := s.queries.SearchProductsFuzzy(ctx, db.SearchProductsFuzzyParams{
		SearchQuery:    &fuzzyQuery,
		CategoryFilter: pgUUIDFromString(req.CategoryID),
		MinPrice:       req.MinPrice,
		MaxPrice:       req.MaxPrice,
		StockOnly:      &req.InStockOnly,
		FuzzyThreshold: nil, // Use default threshold
	})
	if err != nil {
		return nil, err
	}

	// Convert []interface{} to []db.SearchProductsRow
	results := make([]db.SearchProductsRow, 0, len(interfaceResults))
	for _, item := range interfaceResults {
		if row, ok := item.(db.SearchProductsRow); ok {
			results = append(results, row)
		}
	}

	// Apply pagination manually
	start := int(offset)
	end := int(offset + req.Limit)
	if start > len(results) {
		return []db.SearchProductsRow{}, nil
	}
	if end > len(results) {
		end = len(results)
	}
	return results[start:end], nil
}

// generateTrigrams generates trigrams from a query string
func (s *SearchService) generateTrigrams(query string) []string {
	if len(query) < 3 {
		return []string{query}
	}

	words := strings.Fields(query)
	var trigrams []string

	for _, word := range words {
		// For short words, use the word itself
		if len(word) <= 3 {
			trigrams = append(trigrams, word)
			continue
		}

		// Generate trigrams
		for i := 0; i <= len(word)-3; i++ {
			trigrams = append(trigrams, word[i:i+3])
		}
	}

	return trigrams
}

// buildFuzzyQuery builds a PostgreSQL query for fuzzy matching
func (s *SearchService) buildFuzzyQuery(trigrams []string) string {
	if len(trigrams) == 0 {
		return ""
	}

	// Join trigrams with & for AND matching
	return strings.Join(trigrams, " & ")
}

// getSpellingSuggestions returns spelling suggestions for a query
func (s *SearchService) getSpellingSuggestions(query string) []string {
	days := int32(5)
	suggestions, err := s.queries.GetTopSearchQueries(context.Background(), &days)
	if err != nil {
		return nil
	}

	var matches []string
	queryWords := strings.Fields(query)

	for _, suggestion := range suggestions {
		strSuggestion, ok := suggestion.(string)
		if !ok {
			continue
		}
		suggestionWords := strings.Fields(strSuggestion)
		for _, sw := range suggestionWords {
			for _, qw := range queryWords {
				if s.levenshteinDistance(sw, qw) <= 2 {
					matches = append(matches, strSuggestion)
					break
				}
			}
		}
	}

	return matches
}

// getSearchSuggestions returns search suggestions based on partial query
func (s *SearchService) getSearchSuggestions(query string) []string {
	if len(query) < 2 {
		return nil
	}

	days := int32(10)
	suggestions, err := s.queries.GetTopSearchQueries(context.Background(), &days)
	if err != nil {
		return nil
	}

	var matches []string
	lowerQuery := strings.ToLower(query)

	for _, suggestion := range suggestions {
		strSuggestion, ok := suggestion.(string)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(strSuggestion), lowerQuery) {
			matches = append(matches, strSuggestion)
		}
		if len(matches) >= 5 {
			break
		}
	}

	return matches
}

// trackSearch tracks a search query for analytics
func (s *SearchService) trackSearch(ctx context.Context, query string, page int32) {
	if query == "" {
		return
	}

	if err := s.queries.TrackSearch(ctx, db.TrackSearchParams{
		SearchQuery:  &query,
		ResultsCount: nil,
		UserID:       pgtype.UUID{},
	}); err != nil {
		s.logger.Warn("Failed to track search", zap.Error(err))
	}
}

// searchCacheKey generates a cache key for search results
func (s *SearchService) searchCacheKey(req *SearchRequest) string {
	return fmt.Sprintf("search:%s:%s:%d:%d",
		req.Query,
		req.CategoryID,
		req.Page,
		req.Limit,
	)
}

// buildSearchResult builds a search result from product rows
func (s *SearchService) buildSearchResultFromRows(
	rows []db.SearchProductsRow,
	req *SearchRequest,
	duration time.Duration,
	didYouMean, suggestions []string,
) *SearchResult {
	products := make([]*productpb.Product, len(rows))
	for i, row := range rows {
		products[i] = s.searchProductRowToProto(row)
	}

	return &SearchResult{
		Products:    products,
		TotalCount:  int64(len(products)),
		Page:        req.Page,
		Limit:       req.Limit,
		Query:       req.Query,
		Duration:    duration,
		DidYouMean:  didYouMean,
		Suggestions: suggestions,
	}
}

// buildSearchResult builds a search result (for compatibility)
func (s *SearchService) buildSearchResult(products []*db.SearchProductsRow, req *SearchRequest, duration time.Duration) *SearchResult {
	// Convert pointers to values
	rows := make([]db.SearchProductsRow, len(products))
	for i, p := range products {
		if p != nil {
			rows[i] = *p
		}
	}
	return s.buildSearchResultFromRows(rows, req, duration, nil, nil)
}

// WarmCache warms the cache with popular search queries
func (s *SearchService) WarmCache(ctx context.Context) error {
	s.logger.Info("Starting search cache warming")

	// Get top search queries
	days := int32(20)
	topQueries, err := s.queries.GetTopSearchQueries(ctx, &days)
	if err != nil {
		return fmt.Errorf("failed to get top queries: %w", err)
	}

	warmedCount := 0
	for _, query := range topQueries {
		strQuery, ok := query.(string)
		if !ok {
			continue
		}
		// Pre-load search results for each query
		req := &SearchRequest{
			Query:  strQuery,
			Page:   1,
			Limit:  20,
			SortBy: "relevance",
		}

		_, err := s.Search(ctx, req)
		if err != nil {
			s.logger.Warn("Failed to warm cache for query",
				zap.String("query", strQuery),
				zap.Error(err))
			continue
		}

		warmedCount++
	}

	s.logger.Info("Search cache warming completed",
		zap.Int("total_queries", len(topQueries)),
		zap.Int("warmed", warmedCount))

	return nil
}

// GetSearchAnalytics returns search analytics data
func (s *SearchService) GetSearchAnalytics(ctx context.Context, days int) (*SearchAnalytics, error) {
	daysAgo := int32(days)
	topQueries, err := s.queries.GetTopSearchQueries(ctx, &daysAgo)
	if err != nil {
		return nil, err
	}

	// Convert []interface{} to []string
	queryStrings := make([]string, 0, len(topQueries))
	for _, q := range topQueries {
		if str, ok := q.(string); ok {
			queryStrings = append(queryStrings, str)
		}
	}

	return &SearchAnalytics{
		TopQueries:      queryStrings,
		TotalSearches:   int64(len(queryStrings)),
		AverageQueryLen: s.calculateAverageQueryLength(queryStrings),
	}, nil
}

// SearchAnalytics represents search analytics
type SearchAnalytics struct {
	TopQueries      []string
	TotalSearches   int64
	AverageQueryLen float64
}

// calculateAverageQueryLength calculates the average query length
func (s *SearchService) calculateAverageQueryLength(queries []string) float64 {
	if len(queries) == 0 {
		return 0
	}

	totalLen := 0
	for _, q := range queries {
		totalLen += len(q)
	}

	return float64(totalLen) / float64(len(queries))
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (s *SearchService) levenshteinDistance(a, b string) int {
	lenA := len(a)
	lenB := len(b)

	// Create a 2D slice for dynamic programming
	dp := make([][]int, lenA+1)
	for i := range dp {
		dp[i] = make([]int, lenB+1)
	}

	// Initialize base cases
	for i := 0; i <= lenA; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		dp[0][j] = j
	}

	// Fill the DP table
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			dp[i][j] = min(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}

	return dp[lenA][lenB]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// searchProductRowToProto converts a SearchProductsRow to a protobuf Product
func (s *SearchService) searchProductRowToProto(p db.SearchProductsRow) *productpb.Product {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}

	activeVal := false
	if p.Active != nil {
		activeVal = *p.Active
	}

	stockQty := int32(0)
	if p.StockQuantity != nil {
		stockQty = *p.StockQuantity
	}

	return &productpb.Product{
		Id:          pgutil.FromPG(p.ID),
		Name:        p.Name,
		Description: desc,
		CategoryId:  pgutil.FromPG(p.CategoryID),
		Price: &sharedpb.Money{
			Units:    p.PriceUnits,
			Currency: p.PriceCurrency,
		},
		Sku:           p.Sku,
		Active:        activeVal,
		StockQuantity: stockQty,
		ImageUrls:     []string{},
	}
}

// pgUUIDFromString converts a string to pgtype.UUID
func pgUUIDFromString(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{}
	}
	return pgutil.ToPG(uuid.MustParse(s))
}
