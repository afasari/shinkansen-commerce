-- Migration: 000005_add_full_text_search.up.sql
-- Description: Add full-text search with tsvector and GIN index for product search
-- Schema: catalog
-- Author: Shinkansen Commerce
-- Date: 2026-02-27

-- Step 1: Add tsvector column for full-text search
ALTER TABLE catalog.products 
ADD COLUMN IF NOT EXISTS product_search_vector tsvector;

-- Step 2: Create function to update search vector on INSERT/UPDATE
CREATE OR REPLACE FUNCTION catalog.update_product_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    -- Concatenate name and description, then convert to tsvector
    -- Use 'english' configuration for tokenization and stop word removal
    NEW.product_search_vector := to_tsvector('english', 
        COALESCE(NEW.name, '') || ' ' || COALESCE(NEW.description, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 3: Drop trigger if exists (for idempotency)
DROP TRIGGER IF EXISTS product_search_vector_update ON catalog.products;

-- Step 4: Create trigger to auto-update search vector
CREATE TRIGGER product_search_vector_update
BEFORE INSERT OR UPDATE ON catalog.products
FOR EACH ROW
EXECUTE FUNCTION catalog.update_product_search_vector();

-- Step 5: Create GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_products_search 
ON catalog.products USING gin(product_search_vector);

-- Step 6: Backfill existing products (critical for existing data)
UPDATE catalog.products 
SET product_search_vector = to_tsvector('english', 
    COALESCE(name, '') || ' ' || COALESCE(description, '')
)
WHERE product_search_vector IS NULL;

-- Step 7: Add documentation comments
COMMENT ON COLUMN catalog.products.product_search_vector IS 
    'Full-text search vector for product search. Automatically maintained by trigger product_search_vector_update. Uses PostgreSQL tsvector for efficient full-text search.';

COMMENT ON INDEX catalog.idx_products_search IS 
    'GIN index for fast full-text search on products using product_search_vector column. GIN indexes are optimized for tsvector lookups.';

COMMENT ON FUNCTION catalog.update_product_search_vector() IS 
    'Trigger function that updates product_search_vector tsvector on INSERT/UPDATE operations. Concatenates name and description fields into a searchable tsvector.';

COMMENT ON TRIGGER product_search_vector_update ON catalog.products IS 
    'Trigger that automatically updates product_search_vector before INSERT or UPDATE operations. Ensures search index stays in sync with product data.';
