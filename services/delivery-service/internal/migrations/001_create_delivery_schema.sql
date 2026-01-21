-- Create delivery schema
CREATE SCHEMA IF NOT EXISTS delivery;

-- Create delivery_zones table
CREATE TABLE IF NOT EXISTS delivery.delivery_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    postal_codes TEXT[] NOT NULL,
    prefectures TEXT[] NOT NULL,
    delivery_days INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create delivery_slots table
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

-- Create shipments table
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

-- Create delivery_reservations table
CREATE TABLE IF NOT EXISTS delivery.delivery_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES delivery.delivery_slots(id) ON DELETE CASCADE,
    order_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_delivery_slots_zone_id ON delivery.delivery_slots(delivery_zone_id);
CREATE INDEX IF NOT EXISTS idx_delivery_slots_date ON delivery.delivery_slots(date);
CREATE INDEX IF NOT EXISTS idx_delivery_slots_available ON delivery.delivery_slots(available) WHERE available > 0;
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON delivery.shipments(order_id);

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
