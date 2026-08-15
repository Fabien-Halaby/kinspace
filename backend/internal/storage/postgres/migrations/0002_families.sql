CREATE TABLE IF NOT EXISTS families (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 2 AND 120),
    invite_code TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_family_id_fkey;

ALTER TABLE users
    ADD CONSTRAINT users_family_id_fkey
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_families_invite_code ON families(invite_code);
