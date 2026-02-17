-- 000014_create_api_keys_table.up.sql
-- API keys for CI/CD integration authentication

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    key_prefix VARCHAR(8) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_api_keys_workspace_id ON api_keys(workspace_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);
CREATE INDEX idx_api_keys_created_by ON api_keys(created_by);

-- Auto-update updated_at trigger
CREATE TRIGGER set_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE api_keys ENABLE ROW LEVEL SECURITY;

CREATE POLICY api_keys_workspace_isolation ON api_keys
    USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
