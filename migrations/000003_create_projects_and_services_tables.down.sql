DROP TRIGGER IF EXISTS update_services_updated_at ON services;
DROP POLICY IF EXISTS services_isolation ON services;
DROP TABLE IF EXISTS services CASCADE;

DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP POLICY IF EXISTS projects_isolation ON projects;
DROP TABLE IF EXISTS projects CASCADE;
