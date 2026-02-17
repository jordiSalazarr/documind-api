-- Remove foreign key from workspace_members
ALTER TABLE workspace_members DROP CONSTRAINT IF EXISTS fk_workspace_members_user;

-- Drop users table
DROP TABLE IF EXISTS users;
