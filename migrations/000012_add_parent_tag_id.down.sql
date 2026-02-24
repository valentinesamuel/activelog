BEGIN;

DROP INDEX IF EXISTS idx_tags_parent_tag_id;
ALTER TABLE tags DROP COLUMN IF EXISTS parent_tag_id;

COMMIT;
