-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS vaults (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL,          -- ID dari SSO Authorizer
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    icon            VARCHAR(50),
    is_default      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_vaults_user_id ON vaults(user_id);