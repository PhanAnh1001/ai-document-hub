-- Add integration fields to organizations
ALTER TABLE organizations
    ADD COLUMN ocr_provider TEXT NOT NULL DEFAULT 'mock' CHECK (ocr_provider IN ('fpt', 'mock')),
    ADD COLUMN misa_api_url TEXT,
    ADD COLUMN misa_api_key TEXT;
