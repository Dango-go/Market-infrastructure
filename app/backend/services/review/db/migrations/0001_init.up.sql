CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    account_id UUID NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(255) NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS reviews_product_account_idx ON reviews (product_id, account_id);
CREATE INDEX IF NOT EXISTS reviews_product_created_idx ON reviews (product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS reviews_account_created_idx ON reviews (account_id, created_at DESC);
