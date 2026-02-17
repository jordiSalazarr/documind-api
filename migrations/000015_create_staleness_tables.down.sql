-- 000015_create_staleness_tables.down.sql

DROP POLICY IF EXISTS staleness_reports_workspace_isolation ON staleness_reports;
DROP TRIGGER IF EXISTS set_staleness_reports_updated_at ON staleness_reports;
DROP TABLE IF EXISTS staleness_report_items;
DROP TABLE IF EXISTS staleness_reports;
