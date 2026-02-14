-- Name: create_addresses_table
-- Description: Create addresses table
-- Schema: users

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

-- Create indexes
CREATE INDEX idx_addresses_user_id ON users.addresses(user_id);
CREATE INDEX idx_addresses_postal_code ON users.addresses(postal_code);
CREATE INDEX idx_addresses_is_default ON users.addresses(user_id, is_default) WHERE is_default = true;

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

-- Comments
COMMENT ON TABLE users.addresses IS 'User delivery addresses';
COMMENT ON COLUMN users.addresses.id IS 'Unique address identifier';
COMMENT ON COLUMN users.addresses.user_id IS 'User identifier';
COMMENT ON COLUMN users.addresses.name IS 'Contact name';
COMMENT ON COLUMN users.addresses.phone IS 'Contact phone';
COMMENT ON COLUMN users.addresses.postal_code IS 'Postal code';
COMMENT ON COLUMN users.addresses.prefecture IS 'Prefecture (state/region)';
COMMENT ON COLUMN users.addresses.city IS 'City';
COMMENT ON COLUMN users.addresses.address_line1 IS 'Address line 1';
COMMENT ON COLUMN users.addresses.address_line2 IS 'Address line 2 (optional)';
COMMENT ON COLUMN users.addresses.is_default IS 'Default address flag';
COMMENT ON COLUMN users.addresses.created_at IS 'Creation timestamp';
COMMENT ON COLUMN users.addresses.updated_at IS 'Last update timestamp';
