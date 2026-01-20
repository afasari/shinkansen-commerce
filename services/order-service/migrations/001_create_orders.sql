-- Name: create_orders
-- Description: Create orders table with indexes
-- Schema: orders

CREATE TABLE IF NOT EXISTS orders.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_units BIGINT NOT NULL DEFAULT 0,
    total_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    shipping_address JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for user lookups
CREATE INDEX idx_orders_user_id ON orders.orders(user_id);

-- Index for status filtering
CREATE INDEX idx_orders_status ON orders.orders(status);

-- Index for created_at sorting
CREATE INDEX idx_orders_created_at ON orders.orders(created_at DESC);

-- Composite index for user + status
CREATE INDEX idx_orders_user_status ON orders.orders(user_id, status);

-- Comments for documentation
COMMENT ON TABLE orders.orders IS 'Customer orders';
COMMENT ON COLUMN orders.orders.id IS 'Unique order identifier';
COMMENT ON COLUMN orders.orders.user_id IS 'Customer user ID';
COMMENT ON COLUMN orders.orders.status IS 'Order status (pending, confirmed, shipped, delivered, cancelled)';
COMMENT ON COLUMN orders.orders.total_units IS 'Total price in minor units (yen has no minor units)';
COMMENT ON COLUMN orders.orders.total_currency IS 'Currency code (JPY)';
COMMENT ON COLUMN orders.orders.shipping_address IS 'Shipping address as JSONB';
COMMENT ON COLUMN orders.orders.created_at IS 'Creation timestamp';
COMMENT ON COLUMN orders.orders.updated_at IS 'Last update timestamp';
