-- Revert: restore original constraint without 'rejected'
ALTER TABLE accounting_entries
    DROP CONSTRAINT IF EXISTS accounting_entries_status_check;

ALTER TABLE accounting_entries
    ADD CONSTRAINT accounting_entries_status_check
    CHECK (status IN ('draft', 'pending', 'approved', 'synced'));
