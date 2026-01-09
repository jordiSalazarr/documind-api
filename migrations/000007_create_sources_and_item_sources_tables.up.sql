-- Create sources table
CREATE TABLE sources (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  source_type   TEXT NOT NULL CHECK (source_type IN ('pdf', 'url', 'manual')),
  title         TEXT NOT NULL,
  url           TEXT,
  file_path     TEXT,
  file_size     BIGINT,
  date          DATE,
  author        TEXT,
  metadata      JSONB DEFAULT '{}',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by    UUID NOT NULL,
  updated_by    UUID NOT NULL
);

CREATE INDEX idx_sources_workspace ON sources(workspace_id);
CREATE INDEX idx_sources_type ON sources(source_type);

-- Enable RLS
ALTER TABLE sources ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY sources_isolation ON sources
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Trigger to update updated_at
CREATE TRIGGER update_sources_updated_at BEFORE UPDATE ON sources
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create item_sources table
CREATE TABLE item_sources (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  item_id       UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  source_id     UUID NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  linked_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  linked_by     UUID NOT NULL,
  deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_item_sources_unique_active ON item_sources(item_id, source_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_item_sources_workspace ON item_sources(workspace_id);
CREATE INDEX idx_item_sources_item ON item_sources(item_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_item_sources_source ON item_sources(source_id) WHERE deleted_at IS NULL;

-- Enable RLS
ALTER TABLE item_sources ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY item_sources_isolation ON item_sources
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
