-- Fix missing 'rejected' status in accounting_entries CHECK constraint
-- Drop and recreate the constraint to add 'rejected'
ALTER TABLE accounting_entries
    DROP CONSTRAINT IF EXISTS accounting_entries_status_check;

ALTER TABLE accounting_entries
    ADD CONSTRAINT accounting_entries_status_check
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'synced'));
