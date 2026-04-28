-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS password_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id   UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    password_encrypted TEXT NOT NULL,                -- password lama (encrypted)
    encryption_iv   VARCHAR(255) NOT NULL,
    changed_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_history_credential_id ON password_history(credential_id);
CREATE INDEX idx_password_history_changed_at ON password_history(changed_at DESC);