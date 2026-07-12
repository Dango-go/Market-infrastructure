CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY,
    code VARCHAR(120) NOT NULL UNIQUE,
    channel TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'push', 'in_app')),
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL,
    template_id UUID NULL REFERENCES notification_templates(id) ON DELETE SET NULL,
    channel TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'push', 'in_app')),
    status TEXT NOT NULL CHECK (status IN ('draft', 'sent', 'read')),
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    metadata_json TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ NULL,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_account_id_created_at ON notifications (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);

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
