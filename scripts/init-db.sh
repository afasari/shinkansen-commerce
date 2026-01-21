#!/bin/sh

set -e

echo "🗄️  Initializing Shinkansen Commerce Database..."

# Wait for PostgreSQL to be ready
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q'; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done

echo "✅ PostgreSQL is ready"

# Create extensions
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";"

echo "✅ Extensions created"

# Run Product Service migrations
echo "📦 Running Product Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- Product Service Schema
CREATE SCHEMA IF NOT EXISTS catalog;

-- Categories
CREATE TABLE IF NOT EXISTS catalog.categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    parent_id UUID REFERENCES catalog.categories(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Products
CREATE TABLE IF NOT EXISTS catalog.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    price INT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'JPY',
    active BOOLEAN NOT NULL DEFAULT true,
    category_id UUID REFERENCES catalog.categories(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Product Variants
CREATE TABLE IF NOT EXISTS catalog.product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price INT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'JPY',
    sku TEXT UNIQUE,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_category_id ON catalog.products(category_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON catalog.product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_sku ON catalog.product_variants(sku);
EOSQL

echo "✅ Product Service migrations complete"

# Run Order Service migrations
echo "📦 Running Order Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- Order Service Schema
CREATE SCHEMA IF NOT EXISTS orders;

-- Orders
CREATE TABLE IF NOT EXISTS orders.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'ORDER_STATUS_PENDING',
    total_price INT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'JPY',
    shipping_address_id UUID,
    points_applied INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Order Items
CREATE TABLE IF NOT EXISTS orders.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    product_name TEXT NOT NULL,
    variant_id UUID,
    variant_name TEXT,
    quantity INT NOT NULL DEFAULT 1,
    unit_price INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'JPY',
    total_price INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders.orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders.orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON orders.order_items(order_id);
EOSQL

echo "✅ Order Service migrations complete"

# Run User Service migrations
echo "📦 Running User Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- User Service Schema
CREATE SCHEMA IF NOT EXISTS users;

-- Users
CREATE TABLE IF NOT EXISTS users.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Addresses
CREATE TABLE IF NOT EXISTS users.addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    phone TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    prefecture TEXT NOT NULL,
    city TEXT NOT NULL,
    address_line1 TEXT NOT NULL,
    address_line2 TEXT,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Function to set only one default address per user
CREATE OR REPLACE FUNCTION users.set_default_address()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_default = true THEN
        UPDATE users.addresses
        SET is_default = false
        WHERE user_id = NEW.user_id AND id != NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to enforce single default address
DROP TRIGGER IF EXISTS trigger_set_default_address ON users.addresses;
CREATE TRIGGER trigger_set_default_address
BEFORE INSERT OR UPDATE ON users.addresses
FOR EACH ROW
WHEN (NEW.is_default = true)
EXECUTE FUNCTION users.set_default_address();

-- Indexes
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON users.addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_addresses_postal_code ON users.addresses(postal_code);
CREATE INDEX IF NOT EXISTS idx_addresses_is_default ON users.addresses(user_id, is_default) WHERE is_default = true;
EOSQL

echo "✅ User Service migrations complete"

# Run Payment Service migrations
echo "📦 Running Payment Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- Payment Service Schema
CREATE SCHEMA IF NOT EXISTS payments;

-- Payments
CREATE TABLE IF NOT EXISTS payments.payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    method TEXT NOT NULL,
    amount_minor INT NOT NULL,
    currency TEXT NOT NULL DEFAULT 'JPY',
    status TEXT NOT NULL DEFAULT 'PAYMENT_STATUS_PENDING',
    transaction_id TEXT,
    payment_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments.payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments.payments(transaction_id);
EOSQL

echo "✅ Payment Service migrations complete"

# Run Inventory Service migrations
echo "📦 Running Inventory Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- Inventory Service Schema
CREATE SCHEMA IF NOT EXISTS inventory;

-- Stock Items
CREATE TABLE IF NOT EXISTS inventory.stock_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    variant_id UUID,
    warehouse_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 0,
    reserved_quantity INT NOT NULL DEFAULT 0,
    available_quantity INT GENERATED ALWAYS AS (quantity - reserved_quantity) STORED,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_product_warehouse UNIQUE (product_id, variant_id, warehouse_id)
);

-- Stock Movements
CREATE TABLE IF NOT EXISTS inventory.stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stock_item_id UUID NOT NULL REFERENCES inventory.stock_items(id) ON DELETE CASCADE,
    movement_type TEXT NOT NULL,
    quantity INT NOT NULL,
    reference TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Stock Reservations
CREATE TABLE IF NOT EXISTS inventory.stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    stock_item_id UUID NOT NULL REFERENCES inventory.stock_items(id) ON DELETE CASCADE,
    quantity INT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_order_stock UNIQUE (order_id, stock_item_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_stock_items_product_id ON inventory.stock_items(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_items_warehouse_id ON inventory.stock_items(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_stock_item_id ON inventory.stock_movements(stock_item_id);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_order_id ON inventory.stock_reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_stock_reservations_expires_at ON inventory.stock_reservations(expires_at);
EOSQL

echo "✅ Inventory Service migrations complete"

# Run Delivery Service migrations
echo "📦 Running Delivery Service migrations..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOSQL'
-- Delivery Service Schema
CREATE SCHEMA IF NOT EXISTS delivery;

-- Delivery Zones
CREATE TABLE IF NOT EXISTS delivery.delivery_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    postal_codes TEXT[] NOT NULL,
    prefectures TEXT[] NOT NULL,
    delivery_days INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Delivery Slots
CREATE TABLE IF NOT EXISTS delivery.delivery_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_zone_id UUID NOT NULL REFERENCES delivery.delivery_zones(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    capacity INT NOT NULL DEFAULT 10,
    reserved INT NOT NULL DEFAULT 0,
    available INT GENERATED ALWAYS AS (capacity - reserved) STORED,
    date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Shipments
CREATE TABLE IF NOT EXISTS delivery.shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE,
    tracking_number TEXT,
    status TEXT NOT NULL DEFAULT 'SHIPMENT_STATUS_PREPARING',
    estimated_delivery_at TIMESTAMP,
    actual_delivery_at TIMESTAMP,
    carrier TEXT DEFAULT 'Shinkansen Logistics',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Delivery Reservations
CREATE TABLE IF NOT EXISTS delivery.delivery_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES delivery.delivery_slots(id) ON DELETE CASCADE,
    order_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_delivery_slots_zone_id ON delivery.delivery_slots(delivery_zone_id);
CREATE INDEX IF NOT EXISTS idx_delivery_slots_date ON delivery.delivery_slots(date);
CREATE INDEX IF NOT EXISTS idx_delivery_slots_available ON delivery.delivery_slots(available) WHERE available > 0;
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON delivery.shipments(order_id);
EOSQL

echo "✅ Delivery Service migrations complete"

echo "🎉 All migrations completed successfully!"
