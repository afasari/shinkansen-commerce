-- Name: create_products
-- Description: Create products table with indexes for performance
-- Schema: catalog

CREATE TABLE IF NOT EXISTS catalog.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID,
    sku VARCHAR(100) UNIQUE NOT NULL,
    price_units BIGINT NOT NULL,
    price_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    active BOOLEAN DEFAULT true,
    stock_quantity INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Index for category filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON catalog.products(category_id);

CREATE INDEX IF NOT EXISTS idx_products_sku ON catalog.products(sku);

CREATE INDEX IF NOT EXISTS idx_products_active ON catalog.products(active) WHERE active = true;

CREATE INDEX IF NOT EXISTS idx_products_created ON catalog.products(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_products_active_category ON catalog.products(active, category_id) WHERE active = true;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_products_category') THEN
        ALTER TABLE catalog.products ADD CONSTRAINT fk_products_category
            FOREIGN KEY (category_id) REFERENCES catalog.categories(id) ON DELETE SET NULL;
    END IF;
END
$$;

-- Comments for documentation
COMMENT ON TABLE catalog.products IS 'Product catalog items';
COMMENT ON COLUMN catalog.products.id IS 'Unique product identifier';
COMMENT ON COLUMN catalog.products.name IS 'Product name';
COMMENT ON COLUMN catalog.products.description IS 'Product description (HTML)';
COMMENT ON COLUMN catalog.products.category_id IS 'Category identifier';
COMMENT ON COLUMN catalog.products.sku IS 'Stock Keeping Unit';
COMMENT ON COLUMN catalog.products.price_units IS 'Price in minor units (yen has no minor units)';
COMMENT ON COLUMN catalog.products.price_currency IS 'Currency code (JPY)';
COMMENT ON COLUMN catalog.products.active IS 'Is product active for sale';
COMMENT ON COLUMN catalog.products.stock_quantity IS 'Available stock quantity';
COMMENT ON COLUMN catalog.products.created_at IS 'Creation timestamp';
COMMENT ON COLUMN catalog.products.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN catalog.products.deleted_at IS 'Soft delete timestamp';
