BEGIN;

ALTER TABLE tags ADD COLUMN parent_tag_id INT REFERENCES tags(id) ON DELETE SET NULL;

CREATE INDEX idx_tags_parent_tag_id ON tags(parent_tag_id) WHERE parent_tag_id IS NOT NULL;

COMMIT;
