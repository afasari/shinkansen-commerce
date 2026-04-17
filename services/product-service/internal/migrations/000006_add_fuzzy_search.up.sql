-- Migration: 000006_add_fuzzy_search.up.sql
-- Description: Add fuzzy search with trigram similarity and search analytics
-- Schema: catalog
-- Author: Shinkansen Commerce
-- Date: 2026-02-27

-- Step 1: Enable pg_trgm extension for trigram similarity
-- This extension provides functions like similarity() and word_similarity()
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Step 2: Add trigram index for fuzzy name matching
-- GIN index with gin_trgm_ops enables efficient trigram similarity searches
CREATE INDEX IF NOT EXISTS idx_products_name_trgm 
ON catalog.products USING gin(name gin_trgm_ops);

-- Step 3: Add fuzzy search function
-- Combines exact full-text match with trigram similarity
CREATE OR REPLACE FUNCTION catalog.search_products_fuzzy(
    search_query text,
    category_filter uuid DEFAULT NULL,
    min_price bigint DEFAULT NULL,
    max_price bigint DEFAULT NULL,
    stock_only boolean DEFAULT NULL,
    fuzzy_threshold real DEFAULT 0.3
)
RETURNS TABLE (
    id uuid,
    name text,
    description text,
    rank real,
    similarity real
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        p.id,
        p.name,
        p.description,
        ts_rank(p.product_search_vector, plainto_tsquery(search_query)) AS rank,
        SIMILARITY(p.name, search_query) AS similarity
    FROM catalog.products p
    WHERE 
        p.deleted_at IS NULL
        AND p.active = true
        AND (category_filter IS NULL OR p.category_id = category_filter)
        AND (min_price IS NULL OR p.price_units >= min_price)
        AND (max_price IS NULL OR p.price_units <= max_price)
        AND (NOT stock_only OR p.stock_quantity > 0)
        AND (
            p.product_search_vector @@ plainto_tsquery(search_query)
            OR
            SIMILARITY(p.name, search_query) >= fuzzy_threshold
        )
    ORDER BY 
        CASE WHEN p.product_search_vector @@ plainto_tsquery(search_query) THEN 0 ELSE 1 END,
        rank DESC,
        similarity DESC,
        p.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- Step 4: Create search analytics table
-- Tracks search queries for business intelligence
CREATE TABLE IF NOT EXISTS catalog.search_analytics (
    id BIGSERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    results_count INTEGER NOT NULL,
    user_id UUID,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Step 5: Create indexes for search analytics
-- Query index for finding popular searches
CREATE INDEX IF NOT EXISTS idx_search_analytics_query 
ON catalog.search_analytics(query);

-- Created at index for time-based queries
CREATE INDEX IF NOT EXISTS idx_search_analytics_created 
ON catalog.search_analytics(created_at DESC);

-- Composite index for user search history
CREATE INDEX IF NOT EXISTS idx_search_analytics_user_created 
ON catalog.search_analytics(user_id, created_at DESC);

-- Step 6: Create function to track search queries
CREATE OR REPLACE FUNCTION catalog.track_search(
    search_query text,
    results_count integer,
    user_id uuid DEFAULT NULL
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO catalog.search_analytics (query, results_count, user_id)
    VALUES (search_query, results_count, user_id);
END;
$$ LANGUAGE plpgsql;

-- Step 7: Create function to get top search queries
-- Returns most searched queries in last N days
CREATE OR REPLACE FUNCTION catalog.get_top_search_queries(
    days_ago integer DEFAULT 7
)
RETURNS TABLE (
    query text,
    search_count bigint,
    unique_users bigint
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        sa.query,
        COUNT(*) as search_count,
        COUNT(DISTINCT sa.user_id) as unique_users
    FROM catalog.search_analytics sa
    WHERE sa.created_at >= NOW() - INTERVAL '1 day' * days_ago
    GROUP BY sa.query
    ORDER BY search_count DESC
    LIMIT 100;
END;
$$ LANGUAGE plpgsql;

-- Step 8: Add documentation comments
COMMENT ON EXTENSION pg_trgm IS 
    'PostgreSQL extension providing trigram matching for fuzzy text search. Used for typo correction in product search.';

COMMENT ON INDEX catalog.idx_products_name_trgm IS 
    'GIN trigram index for fuzzy name matching. Enables similarity() function for efficient typo correction.';

COMMENT ON FUNCTION catalog.search_products_fuzzy(text, uuid, bigint, bigint, boolean, real) IS 
    'Fuzzy search function combining exact full-text match with trigram similarity. Parameters: search_query, category_filter, min_price, max_price, stock_only, fuzzy_threshold. Returns ranked results with similarity score.';

COMMENT ON TABLE catalog.search_analytics IS 
    'Analytics table tracking search queries, results count, and user behavior. Used for business intelligence, product optimization, and search insights.';

COMMENT ON FUNCTION catalog.track_search(text, integer, uuid) IS 
    'Function to log search queries to analytics table. Tracks query text, results count, and optional user_id for search behavior analysis.';

COMMENT ON FUNCTION catalog.get_top_search_queries(integer) IS 
    'Function to retrieve most popular search queries. Returns top 100 queries from the last N days with search count and unique user count. Useful for understanding user intent and product demand.';
