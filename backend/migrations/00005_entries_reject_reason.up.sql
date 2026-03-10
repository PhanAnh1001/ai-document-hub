-- Add reject_reason column to accounting_entries for storing rejection notes
ALTER TABLE accounting_entries
    ADD COLUMN reject_reason TEXT NULL;
