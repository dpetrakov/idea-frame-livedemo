-- Migration: Add updated_at trigger for users
-- Reuses function update_updated_at_column() defined earlier (see 002_create_initiatives.sql)

BEGIN;

-- Ensure trigger exists to keep users.updated_at in sync on updates
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;