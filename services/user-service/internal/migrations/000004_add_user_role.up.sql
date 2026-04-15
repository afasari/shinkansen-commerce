ALTER TABLE users.users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'customer';

-- Promote the first user to admin (if any exists)
-- UPDATE users.users SET role = 'admin' WHERE id = (SELECT MIN(id) FROM users.users);
