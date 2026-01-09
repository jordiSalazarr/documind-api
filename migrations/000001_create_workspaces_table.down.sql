DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
DROP TABLE IF EXISTS workspaces CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column();
