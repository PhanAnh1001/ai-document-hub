-- Revert integration fields from organizations
ALTER TABLE organizations
    DROP COLUMN IF EXISTS ocr_provider,
    DROP COLUMN IF EXISTS misa_api_url,
    DROP COLUMN IF EXISTS misa_api_key;
