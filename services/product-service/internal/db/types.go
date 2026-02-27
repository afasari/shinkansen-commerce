package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// SearchProductsFuzzyRow represents a single row returned from catalog.search_products_fuzzy()
type SearchProductsFuzzyRow struct {
	// Product identifier
	ID pgtype.UUID
	// Product name
	Name string
	// Product description
	Description *string
	// Full-text search relevance rank (ts_rank)
	Rank float32
	// Trigram similarity score (0.0 to 1.0)
	Similarity float32
}

// GetTopSearchQueriesRow represents a single row returned from catalog.get_top_search_queries()
type GetTopSearchQueriesRow struct {
	// Search query text
	Query string
	// Number of times searched
	SearchCount int64
	// Number of unique users who searched
	UniqueUsers int64
}

// SearchAnalytics represents a record in catalog.search_analytics table
type SearchAnalytics struct {
	ID           int64
	Query        string
	ResultsCount int32
	UserID       *pgtype.UUID
	CreatedAt    time.Time
}
