-- Name: create_categories
-- Description: Create categories table for product categorization
-- Schema: catalog

CREATE TABLE IF NOT EXISTS catalog.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID,
    level INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_categories_parent ON catalog.categories(parent_id);

CREATE INDEX IF NOT EXISTS idx_categories_level ON catalog.categories(level);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_categories_parent') THEN
        ALTER TABLE catalog.categories ADD CONSTRAINT fk_categories_parent
            FOREIGN KEY (parent_id) REFERENCES catalog.categories(id) ON DELETE SET NULL;
    END IF;
END
$$;

-- Comments for documentation
COMMENT ON TABLE catalog.categories IS 'Product categories for hierarchical organization';
COMMENT ON COLUMN catalog.categories.id IS 'Unique category identifier';
COMMENT ON COLUMN catalog.categories.name IS 'Category name';
COMMENT ON COLUMN catalog.categories.parent_id IS 'Parent category identifier (for hierarchical structure)';
COMMENT ON COLUMN catalog.categories.level IS 'Category depth level (0 = root, 1 = second level, etc.)';
COMMENT ON COLUMN catalog.categories.created_at IS 'Creation timestamp';
COMMENT ON COLUMN catalog.categories.updated_at IS 'Last update timestamp';
