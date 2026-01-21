-- Name: create_addresses_table
-- Description: Create addresses table
-- Schema: users

CREATE TABLE IF NOT EXISTS users.addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    postal_code VARCHAR(8) NOT NULL,
    prefecture VARCHAR(50) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_user_id ON users.addresses(user_id);
CREATE INDEX idx_addresses_default ON users.addresses(user_id, is_default) WHERE is_default = true;

COMMENT ON TABLE users.addresses IS 'User delivery addresses';
COMMENT ON COLUMN users.addresses.user_id IS 'User who owns this address';
COMMENT ON COLUMN users.addresses.is_default IS 'Whether this is the default address';
COMMENT ON COLUMN users.addresses.created_at IS 'Creation timestamp';
COMMENT ON COLUMN users.addresses.updated_at IS 'Last update timestamp';
