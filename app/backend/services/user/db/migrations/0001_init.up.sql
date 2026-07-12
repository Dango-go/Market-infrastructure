CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE profiles (
    account_id      UUID PRIMARY KEY,
    email           TEXT NOT NULL,
    username        TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    bio             TEXT NOT NULL DEFAULT '',
    phone           TEXT NOT NULL DEFAULT '',
    avatar_url      TEXT NOT NULL DEFAULT '',
    locale          TEXT NOT NULL DEFAULT 'en',
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    marketing_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_profiles_email ON profiles (lower(email)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_profiles_username ON profiles (lower(username)) WHERE deleted_at IS NULL;
CREATE TRIGGER trg_profiles_updated_at BEFORE UPDATE ON profiles FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE addresses (
    id                UUID PRIMARY KEY,
    account_id        UUID NOT NULL,
    label             TEXT NOT NULL DEFAULT '',
    recipient_name    TEXT NOT NULL DEFAULT '',
    line1             TEXT NOT NULL DEFAULT '',
    line2             TEXT NOT NULL DEFAULT '',
    city              TEXT NOT NULL DEFAULT '',
    region            TEXT NOT NULL DEFAULT '',
    postal_code       TEXT NOT NULL DEFAULT '',
    country_code      TEXT NOT NULL DEFAULT '',
    phone             TEXT NOT NULL DEFAULT '',
    is_default_shipping BOOLEAN NOT NULL DEFAULT FALSE,
    is_default_billing BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_addresses_account_active ON addresses (account_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_addresses_defaults_shipping ON addresses (account_id) WHERE deleted_at IS NULL AND is_default_shipping;
CREATE INDEX idx_addresses_defaults_billing ON addresses (account_id) WHERE deleted_at IS NULL AND is_default_billing;
CREATE TRIGGER trg_addresses_updated_at BEFORE UPDATE ON addresses FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE preferences (
    account_id         UUID PRIMARY KEY,
    currency           TEXT NOT NULL DEFAULT 'USD',
    language           TEXT NOT NULL DEFAULT 'en',
    email_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    sms_notifications  BOOLEAN NOT NULL DEFAULT FALSE,
    push_notifications BOOLEAN NOT NULL DEFAULT FALSE,
    marketing_opt_in   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_preferences_updated_at BEFORE UPDATE ON preferences FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE wishlist_items (
    account_id UUID NOT NULL,
    product_id UUID NOT NULL,
    added_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, product_id)
);

CREATE INDEX idx_wishlist_account_added ON wishlist_items (account_id, added_at DESC);

CREATE TABLE processed_events (
    id           UUID PRIMARY KEY,
    topic        TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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

CREATE INDEX idx_outbox_unpublished ON outbox (occurred_at) WHERE published_at IS NULL;
