-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS credential_custom_fields (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id   UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    field_name      VARCHAR(255) NOT NULL,
    field_value_encrypted TEXT NOT NULL,             -- value (encrypted)
    field_type      VARCHAR(50) DEFAULT 'text',     -- text, password, hidden, totp
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_custom_fields_credential_id ON credential_custom_fields(credential_id);