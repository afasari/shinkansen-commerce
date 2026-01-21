-- Name: create_order_items
-- Description: Create order_items table with indexes
-- Schema: orders

CREATE TABLE IF NOT EXISTS orders.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    variant_id VARCHAR(50),
    product_name VARCHAR(255) NOT NULL,
    quantity INT4 NOT NULL,
    unit_price_units BIGINT NOT NULL,
    unit_price_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    total_price_units BIGINT NOT NULL,
    total_price_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
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
COMMENT ON COLUMN orders.order_items.variant_id IS 'Product variant ID';
COMMENT ON COLUMN orders.order_items.product_name IS 'Product name at time of order';
COMMENT ON COLUMN orders.order_items.quantity IS 'Quantity ordered';
COMMENT ON COLUMN orders.order_items.unit_price_units IS 'Unit price in minor units at time of order';
COMMENT ON COLUMN orders.order_items.unit_price_currency IS 'Unit price currency (JPY)';
COMMENT ON COLUMN orders.order_items.total_price_units IS 'Total price (unit_price * quantity)';
COMMENT ON COLUMN orders.order_items.total_price_currency IS 'Total price currency (JPY)';
COMMENT ON COLUMN orders.order_items.created_at IS 'Creation timestamp';
