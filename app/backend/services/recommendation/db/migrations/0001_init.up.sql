CREATE TABLE IF NOT EXISTS product_profiles (
    product_id UUID PRIMARY KEY,
    slug VARCHAR(180) NOT NULL,
    category_slug VARCHAR(180) NOT NULL DEFAULT '',
    brand_slug VARCHAR(180) NOT NULL DEFAULT '',
    tags_csv TEXT NOT NULL DEFAULT '',
    price_cents BIGINT NOT NULL DEFAULT 0,
    available BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS recommendation_events (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    account_id UUID NULL,
    type VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS product_profiles_category_idx ON product_profiles (category_slug);
CREATE INDEX IF NOT EXISTS product_profiles_brand_idx ON product_profiles (brand_slug);
CREATE INDEX IF NOT EXISTS product_profiles_available_idx ON product_profiles (available);
CREATE INDEX IF NOT EXISTS recommendation_events_product_created_idx ON recommendation_events (product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS recommendation_events_type_created_idx ON recommendation_events (type, created_at DESC);
