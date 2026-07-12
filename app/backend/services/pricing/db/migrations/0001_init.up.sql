CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE prices (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL UNIQUE,
    currency TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    compare_at_cents BIGINT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_prices_currency ON prices(currency);
CREATE TRIGGER trg_prices_updated_at BEFORE UPDATE ON prices FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE promotions (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE,
    discount_type TEXT NOT NULL,
    value_cents BIGINT NOT NULL DEFAULT 0,
    percent_off INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_promotions_discount_type CHECK (discount_type IN ('fixed', 'percent'))
);
CREATE INDEX idx_promotions_active ON promotions(active);
CREATE TRIGGER trg_promotions_updated_at BEFORE UPDATE ON promotions FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    source TEXT NOT NULL,
    subject TEXT NOT NULL,
    correlation_id TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL,
    data JSONB NOT NULL,
    published_at TIMESTAMPTZ
);
CREATE INDEX idx_outbox_unpublished ON outbox (occurred_at) WHERE published_at IS NULL;
