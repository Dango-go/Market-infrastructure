CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    account_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'preparing', 'in_transit', 'delivered', 'cancelled', 'returned')),
    carrier VARCHAR(120) NOT NULL,
    service_level VARCHAR(120) NOT NULL,
    tracking_number VARCHAR(120) NOT NULL DEFAULT '',
    destination_address TEXT NOT NULL,
    eta TIMESTAMPTZ NULL,
    shipped_at TIMESTAMPTZ NULL,
    delivered_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_shipments_account_id_created_at ON shipments (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments (order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_status ON shipments (status);

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
