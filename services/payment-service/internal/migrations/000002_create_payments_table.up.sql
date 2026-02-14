-- Name: create_payments_table
-- Description: Create payments table
-- Schema: payments

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
CREATE INDEX idx_payments_order_id ON payments.payments(order_id);
CREATE INDEX idx_payments_status ON payments.payments(status);
CREATE INDEX idx_payments_transaction_id ON payments.payments(transaction_id);

-- Comments
COMMENT ON TABLE payments.payments IS 'Payment transactions';
COMMENT ON COLUMN payments.payments.id IS 'Unique payment identifier';
COMMENT ON COLUMN payments.payments.order_id IS 'Order ID';
COMMENT ON COLUMN payments.payments.method IS 'Payment method (credit_card, konbini, points)';
COMMENT ON COLUMN payments.payments.amount_minor IS 'Payment amount in minor units';
COMMENT ON COLUMN payments.payments.currency IS 'Currency code (JPY)';
COMMENT ON COLUMN payments.payments.status IS 'Payment status';
COMMENT ON COLUMN payments.payments.transaction_id IS 'Transaction ID from payment gateway';
COMMENT ON COLUMN payments.payments.payment_data IS 'Payment metadata as JSONB';
COMMENT ON COLUMN payments.payments.created_at IS 'Creation timestamp';
COMMENT ON COLUMN payments.payments.updated_at IS 'Last update timestamp';
