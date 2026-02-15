BEGIN;

CREATE TABLE
    activity_photos (
        id SERIAL PRIMARY KEY,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        activity_id INT NOT NULL,
        s3_key TEXT NOT NULL,
        thumbnail_key TEXT,
        content_type VARCHAR(100),
        file_size BIGINT NOT NULL,
        uploaded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_activity FOREIGN KEY (activity_id) REFERENCES activities (id) ON DELETE CASCADE,
        CONSTRAINT check_file_size CHECK (
            file_size >= 2
            AND file_size <= 2457600
        )
    );

CREATE INDEX idx_photos_activity_id ON activity_photos (activity_id);

COMMIT;