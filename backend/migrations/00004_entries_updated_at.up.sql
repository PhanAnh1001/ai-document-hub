-- Add updated_at to accounting_entries and a composite status index

ALTER TABLE accounting_entries
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Update existing rows so updated_at = created_at
UPDATE accounting_entries SET updated_at = created_at;

-- Trigger to keep updated_at current
CREATE TRIGGER set_updated_at_entries
    BEFORE UPDATE ON accounting_entries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Composite index for the most common query: list by org + status
CREATE INDEX IF NOT EXISTS idx_entries_org_status
    ON accounting_entries(organization_id, status);
