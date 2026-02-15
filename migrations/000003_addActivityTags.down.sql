BEGIN;

-- Note: Removed CONCURRENTLY as it cannot run inside a transaction
DROP INDEX IF EXISTS idx_activities_user_date;
DROP INDEX IF EXISTS idx_activities_type;
DROP INDEX IF EXISTS idx_activity_tags_activity;
DROP INDEX IF EXISTS idx_activity_tags_tag;
DROP INDEX IF EXISTS idx_tags_name;

DROP TABLE IF EXISTS activity_tags;

DROP TABLE IF EXISTS tags;

COMMIT;