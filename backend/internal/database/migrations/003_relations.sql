CREATE TABLE IF NOT EXISTS relations (
    id BIGSERIAL PRIMARY KEY,
    family_id BIGINT NOT NULL REFERENCES families(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    related_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('parent','child','spouse','sibling')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id <> related_user_id),
    UNIQUE (family_id, user_id, related_user_id, type)
);

CREATE INDEX IF NOT EXISTS idx_relations_family_id ON relations(family_id);
