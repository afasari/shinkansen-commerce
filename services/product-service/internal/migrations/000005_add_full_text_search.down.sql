-- Migration: 000005_add_full_text_search.down.sql
-- Description: Rollback full-text search migration
-- Schema: catalog
-- Author: Shinkansen Commerce
-- Date: 2026-02-27

-- Step 1: Drop trigger
DROP TRIGGER IF EXISTS product_search_vector_update ON catalog.products;

-- Step 2: Drop function
DROP FUNCTION IF EXISTS catalog.update_product_search_vector();

-- Step 3: Drop GIN index
DROP INDEX IF EXISTS catalog.idx_products_search;

-- Step 4: Drop tsvector column
ALTER TABLE catalog.products 
DROP COLUMN IF EXISTS product_search_vector;
