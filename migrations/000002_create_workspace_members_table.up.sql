-- Create workspace_members table
CREATE TABLE workspace_members (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id  UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id       UUID NOT NULL,
  email         TEXT NOT NULL,
  role          TEXT NOT NULL CHECK (role IN ('reader', 'editor', 'admin')),
  invited_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  joined_at     TIMESTAMPTZ,
  invited_by    UUID,
  UNIQUE(workspace_id, user_id)
);

CREATE INDEX idx_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_members_user ON workspace_members(user_id);
CREATE INDEX idx_members_email ON workspace_members(email);

-- Enable RLS
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can only see members of their workspace
CREATE POLICY workspace_members_isolation ON workspace_members
  USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
