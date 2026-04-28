-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS tags (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL,          -- ID dari SSO Authorizer
    name            VARCHAR(100) NOT NULL,
    color           VARCHAR(7),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_tags_user_id ON tags(user_id);