-- Auth service schema: identities, sessions, external-identity links, and the
-- transactional outbox. UUID primary keys, created_at/updated_at, and soft delete on the
-- account aggregate, per platform conventions.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- updated_at trigger function (shared by tables in this schema).
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- accounts: the identity aggregate.
CREATE TABLE accounts (
    id             UUID PRIMARY KEY,
    email          TEXT NOT NULL,
    username       TEXT NOT NULL,
    password_hash  TEXT,
    status         TEXT NOT NULL DEFAULT 'pending_verification',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    CONSTRAINT chk_accounts_status CHECK (
        status IN ('pending_verification', 'active', 'suspended', 'deactivated')
    )
);

-- Email and username are unique among non-deleted accounts (partial unique indexes free
-- a handle for reuse after an account is soft-deleted).
CREATE UNIQUE INDEX uq_accounts_email
    ON accounts (lower(email)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_accounts_username
    ON accounts (lower(username)) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- sessions: refresh-token grants. Only the SHA-256 of the refresh token is stored.
CREATE TABLE sessions (
    id                 UUID PRIMARY KEY,
    account_id         UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    user_agent         TEXT NOT NULL DEFAULT '',
    ip_address         TEXT NOT NULL DEFAULT '',
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    rotated_from       UUID REFERENCES sessions (id) ON DELETE SET NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_sessions_refresh_token_hash ON sessions (refresh_token_hash);
CREATE INDEX idx_sessions_account_active
    ON sessions (account_id, last_used_at DESC) WHERE revoked_at IS NULL;

-- oauth_identities: external provider links.
CREATE TABLE oauth_identities (
    id               UUID PRIMARY KEY,
    account_id       UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    provider         TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    email            TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_oauth_provider CHECK (provider IN ('github', 'google', 'gitlab'))
);

CREATE UNIQUE INDEX uq_oauth_provider_user ON oauth_identities (provider, provider_user_id);
CREATE INDEX idx_oauth_account ON oauth_identities (account_id);

CREATE TRIGGER trg_oauth_updated_at
    BEFORE UPDATE ON oauth_identities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- outbox: transactional outbox drained to Kafka by the relay.
CREATE TABLE outbox (
    id             UUID PRIMARY KEY,
    type           TEXT NOT NULL,
    version        INTEGER NOT NULL DEFAULT 1,
    source         TEXT NOT NULL,
    subject        TEXT NOT NULL,
    correlation_id TEXT NOT NULL DEFAULT '',
    occurred_at    TIMESTAMPTZ NOT NULL,
    data           JSONB NOT NULL,
    published_at   TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unpublished
    ON outbox (occurred_at) WHERE published_at IS NULL;
