DROP POLICY IF EXISTS item_sources_isolation ON item_sources;
DROP TABLE IF EXISTS item_sources CASCADE;

DROP TRIGGER IF EXISTS update_sources_updated_at ON sources;
DROP POLICY IF EXISTS sources_isolation ON sources;
DROP TABLE IF EXISTS sources CASCADE;
