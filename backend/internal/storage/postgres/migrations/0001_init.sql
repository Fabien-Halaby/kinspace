CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 100),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    family_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_family_id ON users(family_id);
