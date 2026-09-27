-- +goose Up

CREATE TABLE facebook_connections (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    facebook_user_id TEXT NOT NULL,
    facebook_user_name TEXT NOT NULL DEFAULT '',
    encrypted_access_token BYTEA NOT NULL,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT facebook_connections_user_account_unique UNIQUE (user_id, facebook_user_id)
);

CREATE INDEX facebook_connections_user_id_idx ON facebook_connections (user_id);
CREATE INDEX facebook_connections_active_idx ON facebook_connections (user_id, updated_at)
    WHERE revoked_at IS NULL;

-- +goose Down

DROP TABLE IF EXISTS facebook_connections;
