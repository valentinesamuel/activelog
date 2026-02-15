BEGIN;

DROP INDEX IF EXISTS idx_activities_deleted_at;

ALTER TABLE activities DROP COLUMN deleted_at;
ALTER TABLE activity_photos DROP COLUMN deleted_at;
ALTER TABLE activity_tags DROP COLUMN deleted_at;
ALTER TABLE tags DROP COLUMN deleted_at;
ALTER TABLE users DROP COLUMN deleted_at;

COMMIT;
