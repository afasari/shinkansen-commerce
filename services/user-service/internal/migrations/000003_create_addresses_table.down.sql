-- Name: create_addresses_table
-- Description: Drop addresses table and trigger

DROP TRIGGER IF EXISTS trigger_set_default_address ON users.addresses;
DROP FUNCTION IF EXISTS users.set_default_address();
DROP TABLE IF EXISTS users.addresses CASCADE;
