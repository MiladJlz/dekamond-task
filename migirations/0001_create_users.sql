
\connect dekamond

-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    phone TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index for phone number lookups
CREATE INDEX idx_users_phone ON users(phone);

-- Add index for created_at for sorting
CREATE INDEX idx_users_created_at ON users(created_at DESC);

