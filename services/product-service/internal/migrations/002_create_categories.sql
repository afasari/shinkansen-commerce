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

-- Index for parent_id to enable efficient tree queries
CREATE INDEX idx_categories_parent ON catalog.categories(parent_id);

-- Index for level to enable filtering by depth
CREATE INDEX idx_categories_level ON catalog.categories(level);

-- Foreign key constraint for self-referencing
ALTER TABLE catalog.categories ADD CONSTRAINT fk_categories_parent
    FOREIGN KEY (parent_id) REFERENCES catalog.categories(id) ON DELETE SET NULL;

-- Comments for documentation
COMMENT ON TABLE catalog.categories IS 'Product categories for hierarchical organization';
COMMENT ON COLUMN catalog.categories.id IS 'Unique category identifier';
COMMENT ON COLUMN catalog.categories.name IS 'Category name';
COMMENT ON COLUMN catalog.categories.parent_id IS 'Parent category identifier (for hierarchical structure)';
COMMENT ON COLUMN catalog.categories.level IS 'Category depth level (0 = root, 1 = second level, etc.)';
COMMENT ON COLUMN catalog.categories.created_at IS 'Creation timestamp';
COMMENT ON COLUMN catalog.categories.updated_at IS 'Last update timestamp';
