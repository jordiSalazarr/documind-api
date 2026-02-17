-- 000015_create_staleness_tables.up.sql
-- Staleness detection reports and report items

CREATE TABLE IF NOT EXISTS staleness_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    trigger_type VARCHAR(20) NOT NULL CHECK (trigger_type IN ('code_change', 'manual')),
    trigger_ref VARCHAR(255),
    repository_url TEXT,
    diff_summary TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'dismissed', 'resolved')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS staleness_report_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES staleness_reports(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    item_version_id UUID NOT NULL REFERENCES item_versions(id) ON DELETE CASCADE,
    item_title VARCHAR(500) NOT NULL,
    confidence FLOAT NOT NULL DEFAULT 0.0 CHECK (confidence >= 0.0 AND confidence <= 1.0),
    explanation TEXT,
    suggested_update TEXT,
    stale BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for staleness_reports
CREATE INDEX idx_staleness_reports_workspace_id ON staleness_reports(workspace_id);
CREATE INDEX idx_staleness_reports_project_id ON staleness_reports(project_id);
CREATE INDEX idx_staleness_reports_status ON staleness_reports(status);
CREATE INDEX idx_staleness_reports_created_at ON staleness_reports(created_at DESC);

-- Indexes for staleness_report_items
CREATE INDEX idx_staleness_report_items_report_id ON staleness_report_items(report_id);
CREATE INDEX idx_staleness_report_items_item_id ON staleness_report_items(item_id);
CREATE INDEX idx_staleness_report_items_stale ON staleness_report_items(stale);

-- Auto-update updated_at trigger for staleness_reports
CREATE TRIGGER set_staleness_reports_updated_at
    BEFORE UPDATE ON staleness_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE staleness_reports ENABLE ROW LEVEL SECURITY;

CREATE POLICY staleness_reports_workspace_isolation ON staleness_reports
    USING (workspace_id = current_setting('app.workspace_id', true)::uuid);
