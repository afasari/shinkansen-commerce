-- Create payments schema
CREATE SCHEMA IF NOT EXISTS payments;

-- Create payments table
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments.payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments.payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments.payments(transaction_id);

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
