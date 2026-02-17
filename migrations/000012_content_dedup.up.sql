-- Add content hash for deduplication
ALTER TABLE item_versions ADD COLUMN content_hash VARCHAR(64);

-- Index for fast duplicate lookup within a workspace
CREATE INDEX idx_item_versions_content_hash
    ON item_versions(workspace_id, content_hash)
    WHERE content_hash IS NOT NULL;

-- Backfill existing versions (SHA-256 of body_md)
UPDATE item_versions
SET content_hash = encode(sha256(convert_to(body_md, 'UTF8')), 'hex')
WHERE content_hash IS NULL AND body_md IS NOT NULL AND body_md != '';
