BEGIN;

DROP INDEX IF EXISTS idx_photos_activity_id;

DROP TABLE IF EXISTS activity_photos;

COMMIT;