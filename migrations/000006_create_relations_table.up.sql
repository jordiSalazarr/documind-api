-- Create item_relations table
CREATE TABLE item_relations (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id      UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  from_item_id      UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  to_item_id        UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  relation_type_id  UUID NOT NULL REFERENCES relation_types(id) ON DELETE RESTRICT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by        UUID NOT NULL,
  deleted_at        TIMESTAMPTZ,
  CHECK (from_item_id != to_item_id)
);

CREATE UNIQUE INDEX idx_relations_unique_active ON item_relations(from_item_id, to_item_id, relation_type_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_relations_workspace ON item_relations(workspace_id);
CREATE INDEX idx_relations_from ON item_relations(from_item_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_relations_to ON item_relations(to_item_id) WHERE deleted_at IS NULL;

-- Enable RLS
ALTER TABLE item_relations ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY item_relations_isolation ON item_relations
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
