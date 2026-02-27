-- Migration: 000006_add_fuzzy_search.down.sql
-- Description: Rollback fuzzy search and search analytics
-- Schema: catalog
-- Author: Shinkansen Commerce
-- Date: 2026-02-27

-- Step 1: Drop analytics functions
DROP FUNCTION IF EXISTS catalog.get_top_search_queries();

-- Step 2: Drop tracking function
DROP FUNCTION IF EXISTS catalog.track_search();

-- Step 3: Drop search analytics indexes
DROP INDEX IF EXISTS catalog.idx_search_analytics_user_created;
DROP INDEX IF EXISTS catalog.idx_search_analytics_created;
DROP INDEX IF EXISTS catalog.idx_search_analytics_query;

-- Step 4: Drop search analytics table
DROP TABLE IF EXISTS catalog.search_analytics;

-- Step 5: Drop fuzzy search function
DROP FUNCTION IF EXISTS catalog.search_products_fuzzy();

-- Step 6: Drop trigram index
DROP INDEX IF EXISTS catalog.idx_products_name_trgm;

-- Step 7: Drop pg_trgm extension (optional, usually keep)
-- COMMENT: We keep the extension as it might be used elsewhere
-- DROP EXTENSION IF EXISTS pg_trgm;
