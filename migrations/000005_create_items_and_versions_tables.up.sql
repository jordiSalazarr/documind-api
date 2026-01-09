-- Create items table
CREATE TABLE items (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  service_id      UUID REFERENCES services(id) ON DELETE SET NULL,
  item_type_id    UUID NOT NULL REFERENCES item_types(id) ON DELETE RESTRICT,
  latest_version  INT NOT NULL DEFAULT 1,
  status          TEXT NOT NULL CHECK (status IN ('draft', 'published', 'deprecated')),
  owner_user_id   UUID NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by      UUID NOT NULL,
  updated_by      UUID NOT NULL,
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_items_workspace ON items(workspace_id);
CREATE INDEX idx_items_project ON items(project_id);
CREATE INDEX idx_items_service ON items(service_id);
CREATE INDEX idx_items_status ON items(status) WHERE deleted_at IS NULL;

-- Enable RLS
ALTER TABLE items ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY items_isolation ON items
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);

-- Trigger to update updated_at
CREATE TRIGGER update_items_updated_at BEFORE UPDATE ON items
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create item_versions table
CREATE TABLE item_versions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id       UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  version       INT NOT NULL,
  title         TEXT NOT NULL,
  summary       TEXT NOT NULL,
  body_md       TEXT NOT NULL,
  custom_fields JSONB DEFAULT '{}',
  tags          TEXT[] DEFAULT '{}',
  status        TEXT NOT NULL CHECK (status IN ('draft', 'published', 'deprecated')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_by    UUID NOT NULL,
  embedding     VECTOR(1536),
  search_vector TSVECTOR GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(summary, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(body_md, '')), 'C')
  ) STORED,
  UNIQUE(item_id, version)
);

CREATE INDEX idx_item_versions_item ON item_versions(item_id);
CREATE INDEX idx_item_versions_workspace ON item_versions(workspace_id);
CREATE INDEX idx_item_versions_status ON item_versions(status);
CREATE INDEX idx_item_versions_fts ON item_versions USING GIN(search_vector);
CREATE INDEX idx_item_versions_embedding ON item_versions USING hnsw (embedding vector_cosine_ops);

-- Enable RLS
ALTER TABLE item_versions ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Workspace isolation
CREATE POLICY item_versions_isolation ON item_versions
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
