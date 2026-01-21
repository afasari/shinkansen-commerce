-- Name: create_orders
-- Description: Create orders table with indexes
-- Schema: orders

CREATE SCHEMA IF NOT EXISTS orders;

CREATE TABLE IF NOT EXISTS orders.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    status INT4 NOT NULL DEFAULT 0,
    subtotal_units BIGINT NOT NULL DEFAULT 0,
    subtotal_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    tax_units BIGINT NOT NULL DEFAULT 0,
    tax_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    discount_units BIGINT NOT NULL DEFAULT 0,
    discount_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    total_units BIGINT NOT NULL DEFAULT 0,
    total_currency VARCHAR(3) NOT NULL DEFAULT 'JPY',
    points_applied INT4 NOT NULL DEFAULT 0,
    shipping_address JSONB NOT NULL,
    payment_method INT4 NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for user lookups
CREATE INDEX idx_orders_user_id ON orders.orders(user_id);

-- Index for order number lookups
CREATE INDEX idx_orders_order_number ON orders.orders(order_number);

-- Index for status filtering
CREATE INDEX idx_orders_status ON orders.orders(status);

-- Index for created_at sorting
CREATE INDEX idx_orders_created_at ON orders.orders(created_at DESC);

-- Composite index for user + status
CREATE INDEX idx_orders_user_status ON orders.orders(user_id, status);

-- Comments for documentation
COMMENT ON TABLE orders.orders IS 'Customer orders';
COMMENT ON COLUMN orders.orders.id IS 'Unique order identifier';
COMMENT ON COLUMN orders.orders.order_number IS 'Human-readable order number';
COMMENT ON COLUMN orders.orders.user_id IS 'Customer user ID';
COMMENT ON COLUMN orders.orders.status IS 'Order status: 0=pending, 1=confirmed, 2=shipped, 3=delivered, 4=cancelled';
COMMENT ON COLUMN orders.orders.subtotal_units IS 'Subtotal price in minor units (yen has no minor units)';
COMMENT ON COLUMN orders.orders.subtotal_currency IS 'Subtotal currency code (JPY)';
COMMENT ON COLUMN orders.orders.tax_units IS 'Tax amount in minor units';
COMMENT ON COLUMN orders.orders.tax_currency IS 'Tax currency code (JPY)';
COMMENT ON COLUMN orders.orders.discount_units IS 'Discount amount in minor units';
COMMENT ON COLUMN orders.orders.discount_currency IS 'Discount currency code (JPY)';
COMMENT ON COLUMN orders.orders.total_units IS 'Total price in minor units';
COMMENT ON COLUMN orders.orders.total_currency IS 'Total currency code (JPY)';
COMMENT ON COLUMN orders.orders.points_applied IS 'Loyalty points applied to order';
COMMENT ON COLUMN orders.orders.shipping_address IS 'Shipping address as JSONB';
COMMENT ON COLUMN orders.orders.payment_method IS 'Payment method: 0=credit_card, 1=konbini, 2=points';
COMMENT ON COLUMN orders.orders.created_at IS 'Creation timestamp';
COMMENT ON COLUMN orders.orders.updated_at IS 'Last update timestamp';
