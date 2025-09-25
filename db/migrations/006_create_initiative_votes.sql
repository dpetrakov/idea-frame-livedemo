-- Migration: Create initiative_votes table (up/down voting)
-- Aligns with db/schema.dbml#initiative_votes and PRD V2

BEGIN;

CREATE TABLE IF NOT EXISTS initiative_votes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    initiative_id  UUID NOT NULL REFERENCES initiatives(id) ON DELETE CASCADE ON UPDATE RESTRICT,
    user_id        UUID NOT NULL REFERENCES users(id)        ON DELETE CASCADE ON UPDATE RESTRICT,
    value          SMALLINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure only -1 and 1 are accepted
ALTER TABLE initiative_votes
    ADD CONSTRAINT chk_votes_value CHECK (value IN (-1, 1));

-- One vote per user per initiative
CREATE UNIQUE INDEX IF NOT EXISTS ux_votes_initiative_user ON initiative_votes (initiative_id, user_id);
CREATE INDEX IF NOT EXISTS idx_votes_initiative ON initiative_votes (initiative_id);
CREATE INDEX IF NOT EXISTS idx_votes_user ON initiative_votes (user_id);

-- Reuse timestamp trigger function
CREATE TRIGGER update_initiative_votes_updated_at
    BEFORE UPDATE ON initiative_votes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE initiative_votes IS 'Голоса пользователей за инициативы: -1 (down), 1 (up). Один активный голос на пользователя.';

COMMIT;



