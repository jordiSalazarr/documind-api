-- 000014_create_api_keys_table.down.sql

DROP POLICY IF EXISTS api_keys_workspace_isolation ON api_keys;
DROP TRIGGER IF EXISTS set_api_keys_updated_at ON api_keys;
DROP TABLE IF EXISTS api_keys;
