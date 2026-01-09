-- Create item_types table
CREATE TABLE item_types (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  slug          TEXT NOT NULL,
  schema_json   JSONB DEFAULT '{}',
  icon          TEXT,
  color         TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workspace_id, slug)
);

CREATE INDEX idx_item_types_workspace ON item_types(workspace_id);

-- Enable RLS
ALTER TABLE item_types ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY item_types_isolation ON item_types
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Trigger to update updated_at
CREATE TRIGGER update_item_types_updated_at BEFORE UPDATE ON item_types
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create relation_types table
CREATE TABLE relation_types (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  slug            TEXT NOT NULL,
  is_directional  BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(workspace_id, slug)
);

CREATE INDEX idx_relation_types_workspace ON relation_types(workspace_id);

-- Enable RLS
ALTER TABLE relation_types ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY relation_types_isolation ON relation_types
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Trigger to update updated_at
CREATE TRIGGER update_relation_types_updated_at BEFORE UPDATE ON relation_types
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
