DROP TRIGGER IF EXISTS set_updated_at_documents ON documents;
DROP TRIGGER IF EXISTS set_updated_at_organizations ON organizations;
DROP FUNCTION IF EXISTS update_updated_at();
DROP TABLE IF EXISTS accounting_entries;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS organizations;
