DROP TRIGGER IF EXISTS update_relation_types_updated_at ON relation_types;
DROP POLICY IF EXISTS relation_types_isolation ON relation_types;
DROP TABLE IF EXISTS relation_types CASCADE;

DROP TRIGGER IF EXISTS update_item_types_updated_at ON item_types;
DROP POLICY IF EXISTS item_types_isolation ON item_types;
DROP TABLE IF EXISTS item_types CASCADE;
