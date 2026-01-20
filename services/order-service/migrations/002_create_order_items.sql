-- Name: create_order_items
-- Description: Create order_items table with indexes
-- Schema: orders

CREATE TABLE IF NOT EXISTS orders.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INT4 NOT NULL,
    price_units BIGINT NOT NULL,
    price_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_order_product UNIQUE(order_id, product_id)
);

-- Index for order lookups
CREATE INDEX idx_order_items_order_id ON orders.order_items(order_id);

-- Index for product lookups
CREATE INDEX idx_order_items_product_id ON orders.order_items(product_id);

-- Comments for documentation
COMMENT ON TABLE orders.order_items IS 'Order items/products';
COMMENT ON COLUMN orders.order_items.id IS 'Unique item identifier';
COMMENT ON COLUMN orders.order_items.order_id IS 'Parent order ID';
COMMENT ON COLUMN orders.order_items.product_id IS 'Product ID';
COMMENT ON COLUMN orders.order_items.quantity IS 'Quantity ordered';
COMMENT ON COLUMN orders.order_items.price_units IS 'Price at time of order';
COMMENT ON COLUMN orders.order_items.price_currency IS 'Currency at time of order';
COMMENT ON COLUMN orders.order_items.created_at IS 'Creation timestamp';
