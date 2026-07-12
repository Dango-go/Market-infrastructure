CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    account_id UUID NULL,
    session_id VARCHAR(120) NOT NULL,
    product_id UUID NULL,
    event_type VARCHAR(40) NOT NULL,
    path VARCHAR(500) NOT NULL DEFAULT '',
    referrer VARCHAR(500) NOT NULL DEFAULT '',
    query VARCHAR(255) NOT NULL DEFAULT '',
    user_agent VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS analytics_events_created_idx ON analytics_events (created_at DESC);
CREATE INDEX IF NOT EXISTS analytics_events_session_idx ON analytics_events (session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS analytics_events_product_idx ON analytics_events (product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS analytics_events_type_idx ON analytics_events (event_type, created_at DESC);
