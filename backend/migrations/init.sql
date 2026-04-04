-- Initial schema for AI Document Hub
-- Run this against a fresh PostgreSQL database

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS vector;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id          VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email       VARCHAR(255) UNIQUE NOT NULL,
    hashed_password VARCHAR(255) NOT NULL,
    full_name   VARCHAR(255),
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Documents table
CREATE TABLE IF NOT EXISTS documents (
    id                  VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id             VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename            VARCHAR(500) NOT NULL,
    original_filename   VARCHAR(500) NOT NULL,
    file_path           VARCHAR(1000),
    file_size           FLOAT,
    mime_type           VARCHAR(100),
    doc_type            VARCHAR(50) DEFAULT 'other',
    status              VARCHAR(50) DEFAULT 'uploaded',
    ocr_text            TEXT,
    ocr_confidence      FLOAT,
    extracted_data      JSONB,
    error_message       TEXT,
    created_at          TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_doc_type ON documents(doc_type);

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Document chunks (for RAG)
CREATE TABLE IF NOT EXISTS document_chunks (
    id              VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    document_id     VARCHAR(36) NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index     FLOAT NOT NULL,
    chunk_text      TEXT NOT NULL,
    embedding_json  JSONB,  -- fallback when pgvector not available
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chunks_document_id ON document_chunks(document_id);

-- Chat history
CREATE TABLE IF NOT EXISTS chat_history (
    id              VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id         VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question        TEXT NOT NULL,
    answer          TEXT,
    context_chunks  JSONB,
    created_at      TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_history_user_id ON chat_history(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_history_created_at ON chat_history(created_at DESC);

-- Insert default system user (used for unauthenticated MVP)
INSERT INTO users (id, email, hashed_password, full_name)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system@ai-document-hub.local',
    '$2b$12$placeholder_not_for_login',
    'System User'
) ON CONFLICT (id) DO NOTHING;
