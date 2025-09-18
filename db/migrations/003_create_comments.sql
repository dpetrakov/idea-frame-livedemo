-- Migration: Create comments table (chat-style comments)
-- Aligns with db/schema.dbml#comments

BEGIN;

-- Create table
CREATE TABLE IF NOT EXISTS comments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initiative_id  UUID NOT NULL REFERENCES initiatives(id) ON DELETE CASCADE ON UPDATE RESTRICT,
    author_id      UUID NOT NULL REFERENCES users(id)        ON DELETE RESTRICT ON UPDATE RESTRICT,
    text           VARCHAR(1000) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Constraints
ALTER TABLE comments
    ADD CONSTRAINT chk_comments_text_length
    CHECK (char_length(text) >= 1 AND char_length(text) <= 1000);

-- Indexes to support lookups and ordering
CREATE INDEX IF NOT EXISTS idx_comments_initiative_created ON comments (initiative_id, created_at);
CREATE INDEX IF NOT EXISTS idx_comments_author            ON comments (author_id);

COMMIT;