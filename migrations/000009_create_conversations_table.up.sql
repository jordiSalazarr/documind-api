-- Create conversations table for chat history
CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    service_id      UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    title           TEXT NOT NULL DEFAULT 'New Conversation',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_conversations_workspace ON conversations(workspace_id);
CREATE INDEX idx_conversations_service ON conversations(service_id);
CREATE INDEX idx_conversations_user ON conversations(user_id);
CREATE INDEX idx_conversations_updated ON conversations(updated_at DESC) WHERE deleted_at IS NULL;

-- Enable RLS
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY conversations_isolation ON conversations
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Trigger to update updated_at
CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create messages table
CREATE TABLE messages (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id   UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    workspace_id      UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    role              TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content           TEXT NOT NULL,
    sources           JSONB DEFAULT '[]',
    token_count       INT,
    model             TEXT,
    latency_ms        INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at);
CREATE INDEX idx_messages_workspace ON messages(workspace_id);
CREATE INDEX idx_messages_sources ON messages USING GIN(sources);

-- Enable RLS
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY messages_isolation ON messages
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Comments for documentation
COMMENT ON TABLE conversations IS 'Stores chat conversations for RAG interactions';
COMMENT ON TABLE messages IS 'Stores individual messages within conversations';
COMMENT ON COLUMN messages.role IS 'Message sender: user, assistant (AI), or system';
COMMENT ON COLUMN messages.sources IS 'JSON array of source references used for the response';
COMMENT ON COLUMN messages.token_count IS 'Total tokens used (prompt + completion)';
COMMENT ON COLUMN messages.model IS 'LLM model used for generation (e.g., gpt-4o-mini)';
COMMENT ON COLUMN messages.latency_ms IS 'Response generation time in milliseconds';
