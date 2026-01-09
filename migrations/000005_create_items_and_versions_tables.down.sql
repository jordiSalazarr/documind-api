DROP POLICY IF EXISTS item_versions_isolation ON item_versions;
DROP TABLE IF EXISTS item_versions CASCADE;

DROP TRIGGER IF EXISTS update_items_updated_at ON items;
DROP POLICY IF EXISTS items_isolation ON items;
DROP TABLE IF EXISTS items CASCADE;
