-- Migration: Create email_codes table for e-mail verification codes
-- Aligns with db/schema.dbml#email_codes and PRD V2

BEGIN;

CREATE TABLE IF NOT EXISTS email_codes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL,
    code          VARCHAR(6) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    used_at       TIMESTAMPTZ,
    requested_ip  VARCHAR(64)
);

-- Indexes to support lookups and purging
CREATE INDEX IF NOT EXISTS idx_email_codes_email       ON email_codes (email);
CREATE INDEX IF NOT EXISTS idx_email_codes_email_code  ON email_codes (email, code);
CREATE INDEX IF NOT EXISTS idx_email_codes_expires_at  ON email_codes (expires_at);

COMMENT ON TABLE email_codes IS 'Одноразовые коды подтверждения e‑mail с TTL и audit-полями';

COMMIT;



