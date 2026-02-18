BEGIN;

CREATE TABLE IF NOT EXISTS inbox_event (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    process_response TEXT,
    checkpoint JSONB,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    job_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_inbox_event_status ON inbox_event(status);
CREATE INDEX idx_inbox_event_event_type ON inbox_event(event_type);
CREATE INDEX idx_inbox_event_created_at ON inbox_event(created_at);

CREATE TABLE IF NOT EXISTS outbox_event (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    process_response TEXT,
    checkpoint JSONB,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    job_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_outbox_event_status ON outbox_event(status);
CREATE INDEX idx_outbox_event_event_type ON outbox_event(event_type);
CREATE INDEX idx_outbox_event_created_at ON outbox_event(created_at);

COMMIT;
