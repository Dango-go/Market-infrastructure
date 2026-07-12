CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    cart_id UUID NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'processing', 'shipped', 'completed', 'cancelled')),
    currency VARCHAR(8) NOT NULL,
    subtotal_cents BIGINT NOT NULL DEFAULT 0,
    shipping_cents BIGINT NOT NULL DEFAULT 0,
    total_cents BIGINT NOT NULL DEFAULT 0,
    delivery_method VARCHAR(80) NOT NULL,
    delivery_address TEXT NOT NULL,
    customer_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_account_id_created_at ON orders (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);

CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    version INTEGER NOT NULL,
    source TEXT NOT NULL,
    subject TEXT NOT NULL,
    correlation_id TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    data JSONB NOT NULL,
    published_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished ON outbox (published_at, occurred_at) WHERE published_at IS NULL;
