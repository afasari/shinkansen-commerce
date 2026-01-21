-- Create inventory schema
CREATE SCHEMA IF NOT EXISTS inventory;

-- Create stock_items table
CREATE TABLE IF NOT EXISTS inventory.stock_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    variant_id UUID,
    warehouse_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    available_quantity INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_product_warehouse UNIQUE (product_id, variant_id, warehouse_id)
);

-- Create stock_movements table
CREATE TABLE IF NOT EXISTS inventory.stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_item_id UUID NOT NULL REFERENCES inventory.stock_items(id) ON DELETE CASCADE,
    movement_type TEXT NOT NULL,
    quantity INT NOT NULL,
    reference TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create stock_reservations table
CREATE TABLE IF NOT EXISTS inventory.stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    stock_item_id UUID NOT NULL REFERENCES inventory.stock_items(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_order_stock UNIQUE (order_id, stock_item_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_stock_items_product_id ON inventory.stock_items(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_items_warehouse_id ON inventory.stock_items(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_stock_item_id ON inventory.stock_movements(stock_item_id);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_order_id ON inventory.stock_reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_expires_at ON inventory.stock_reservations(expires_at);

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create trigger to update available_quantity
CREATE OR REPLACE FUNCTION inventory.update_available_quantity()
RETURNS TRIGGER AS $$
BEGIN
    NEW.available_quantity := NEW.quantity - NEW.reserved_quantity;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_available_quantity
BEFORE INSERT OR UPDATE ON inventory.stock_items
FOR EACH ROW EXECUTE FUNCTION inventory.update_available_quantity();
