-- Create chunks table for improved RAG retrieval
CREATE TABLE chunks (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_version_id   UUID NOT NULL REFERENCES item_versions(id) ON DELETE CASCADE,
    workspace_id      UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    content           TEXT NOT NULL,
    chunk_index       INT NOT NULL,
    chunk_level       INT NOT NULL DEFAULT 1,
    token_count       INT NOT NULL,
    char_count        INT NOT NULL,
    heading           TEXT,
    parent_chunk_id   UUID REFERENCES chunks(id) ON DELETE SET NULL,
    embedding         VECTOR(1536),
    metadata          JSONB DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,

    CONSTRAINT unique_version_chunk UNIQUE(item_version_id, chunk_index),
    CONSTRAINT valid_chunk_level CHECK (chunk_level IN (1, 2, 3)),
    CONSTRAINT positive_token_count CHECK (token_count > 0),
    CONSTRAINT positive_char_count CHECK (char_count > 0)
);

-- Indexes for performance
CREATE INDEX idx_chunks_item_version ON chunks(item_version_id);
CREATE INDEX idx_chunks_workspace ON chunks(workspace_id);
CREATE INDEX idx_chunks_parent ON chunks(parent_chunk_id) WHERE parent_chunk_id IS NOT NULL;
CREATE INDEX idx_chunks_metadata ON chunks USING GIN(metadata);

-- Vector search index (HNSW for fast approximate nearest neighbor)
CREATE INDEX idx_chunks_embedding ON chunks
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Full-text search on chunk content
CREATE INDEX idx_chunks_fts ON chunks
USING GIN(to_tsvector('english', content));

-- Composite index for workspace + embedding filtered queries
CREATE INDEX idx_chunks_workspace_published ON chunks(workspace_id)
WHERE deleted_at IS NULL;

-- Enable Row Level Security
ALTER TABLE chunks ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY chunks_isolation ON chunks
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Comments for documentation
COMMENT ON TABLE chunks IS 'Stores document chunks for RAG retrieval with embeddings';
COMMENT ON COLUMN chunks.chunk_level IS 'Chunking hierarchy: 1=section-based, 2=paragraph-based, 3=sentence-based';
COMMENT ON COLUMN chunks.metadata IS 'Stores additional chunk info: start_char_offset, end_char_offset, code_block, language, page_numbers';
COMMENT ON COLUMN chunks.embedding IS 'OpenAI text-embedding-3-small (1536 dimensions)';
