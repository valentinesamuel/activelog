BEGIN;

ALTER TABLE activities ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE activity_photos ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE activity_tags ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE tags ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Index for faster queries
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;

COMMIT;
