BEGIN;

CREATE TABLE IF NOT EXISTS daily_stats (
    id                      SERIAL PRIMARY KEY,
    user_id                 INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date                    DATE NOT NULL,
    total_activities        INTEGER NOT NULL DEFAULT 0,
    total_distance_km       NUMERIC(10, 2) NOT NULL DEFAULT 0,
    total_duration_minutes  INTEGER NOT NULL DEFAULT 0,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_stats_user_date ON daily_stats (user_id, date);

COMMIT;
