CREATE TYPE webhook_delivery_status AS ENUM ('pending', 'succeeded', 'failed', 'exhausted');

CREATE TABLE webhook_deliveries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id       UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type       TEXT NOT NULL,
    payload          JSONB NOT NULL,
    status           webhook_delivery_status NOT NULL DEFAULT 'pending',
    attempt_count    INTEGER NOT NULL DEFAULT 0,
    max_attempts     INTEGER NOT NULL DEFAULT 5,
    last_http_status INTEGER,
    last_error       TEXT,
    next_retry_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX idx_webhook_deliveries_status_retry ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
