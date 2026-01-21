-- Name: create_users_schema
-- Description: Create users schema
-- Schema: public

CREATE SCHEMA IF NOT EXISTS users;

-- Name: create_users_table
-- Description: Create users table
-- Schema: users

CREATE TABLE IF NOT EXISTS users.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(20),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users.users(email);
CREATE INDEX idx_users_active ON users.users(active);

COMMENT ON TABLE users.users IS 'User accounts';
COMMENT ON COLUMN users.users.id IS 'Unique user identifier';
COMMENT ON COLUMN users.users.email IS 'User email (unique)';
COMMENT ON COLUMN users.users.password_hash IS 'BCrypt password hash';
COMMENT ON COLUMN users.users.active IS 'User active status';
COMMENT ON COLUMN users.users.created_at IS 'Creation timestamp';
COMMENT ON COLUMN users.users.updated_at IS 'Last update timestamp';
