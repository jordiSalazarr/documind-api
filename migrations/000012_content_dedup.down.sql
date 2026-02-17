DROP INDEX IF EXISTS idx_item_versions_content_hash;
ALTER TABLE item_versions DROP COLUMN IF EXISTS content_hash;
