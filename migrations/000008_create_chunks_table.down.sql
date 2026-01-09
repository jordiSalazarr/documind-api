-- Drop RLS policy
DROP POLICY IF EXISTS chunks_isolation ON chunks;

-- Drop indexes
DROP INDEX IF EXISTS idx_chunks_workspace_published;
DROP INDEX IF EXISTS idx_chunks_fts;
DROP INDEX IF EXISTS idx_chunks_embedding;
DROP INDEX IF EXISTS idx_chunks_metadata;
DROP INDEX IF EXISTS idx_chunks_parent;
DROP INDEX IF EXISTS idx_chunks_workspace;
DROP INDEX IF EXISTS idx_chunks_item_version;

-- Drop chunks table
DROP TABLE IF EXISTS chunks;
