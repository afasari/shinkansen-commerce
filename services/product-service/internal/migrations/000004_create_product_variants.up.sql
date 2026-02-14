-- Name: create_product_variants
-- Description: Create product_variants table for managing product options (sizes, colors, etc.)
-- Schema: catalog

CREATE TABLE IF NOT EXISTS catalog.product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    attributes JSONB,
    price_units BIGINT NOT NULL,
    price_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    sku VARCHAR(100),
    stock_quantity INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Foreign key constraint linking variants to products
ALTER TABLE catalog.product_variants ADD CONSTRAINT fk_variants_product
    FOREIGN KEY (product_id) REFERENCES catalog.products(id) ON DELETE CASCADE;

-- Index for efficient product variant lookups
CREATE INDEX idx_variants_product ON catalog.product_variants(product_id);

-- Index for SKU filtering (if not null)
CREATE INDEX idx_variants_sku ON catalog.product_variants(sku) WHERE sku IS NOT NULL;

-- Index for stock quantity filtering (in-stock variants)
CREATE INDEX idx_variants_stock ON catalog.product_variants(stock_quantity) WHERE stock_quantity > 0;

-- Comments for documentation
COMMENT ON TABLE catalog.product_variants IS 'Product variants (sizes, colors, etc.)';
COMMENT ON COLUMN catalog.product_variants.id IS 'Unique variant identifier';
COMMENT ON COLUMN catalog.product_variants.product_id IS 'Parent product identifier';
COMMENT ON COLUMN catalog.product_variants.name IS 'Variant name (e.g., "Small", "Blue")';
COMMENT ON COLUMN catalog.product_variants.attributes IS 'Variant attributes as JSON (e.g., {"size": "M", "color": "Blue"})';
COMMENT ON COLUMN catalog.product_variants.price_units IS 'Price in minor units (yen has no minor units)';
COMMENT ON COLUMN catalog.product_variants.price_currency IS 'Currency code (JPY)';
COMMENT ON COLUMN catalog.product_variants.sku IS 'Stock Keeping Unit for variant';
COMMENT ON COLUMN catalog.product_variants.stock_quantity IS 'Available stock quantity';
COMMENT ON COLUMN catalog.product_variants.created_at IS 'Creation timestamp';
COMMENT ON COLUMN catalog.product_variants.updated_at IS 'Last update timestamp';
