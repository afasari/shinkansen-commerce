-- Create addresses table
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
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON users.addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_addresses_postal_code ON users.addresses(postal_code);
CREATE INDEX IF NOT EXISTS idx_addresses_is_default ON users.addresses(user_id, is_default) WHERE is_default = true;

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
CREATE TRIGGER trigger_set_default_address
BEFORE INSERT OR UPDATE ON users.addresses
FOR EACH ROW
WHEN (NEW.is_default = true)
EXECUTE FUNCTION users.set_default_address();
