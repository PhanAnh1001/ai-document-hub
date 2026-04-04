-- Revert: remove updated_at from accounting_entries and its index

DROP TRIGGER IF EXISTS set_updated_at_entries ON accounting_entries;
DROP INDEX IF EXISTS idx_entries_org_status;
ALTER TABLE accounting_entries DROP COLUMN IF EXISTS updated_at;
