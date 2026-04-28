-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL,          -- ID dari SSO Authorizer
    action          VARCHAR(50) NOT NULL,           -- VIEW, CREATE, UPDATE, DELETE, EXPORT, COPY_PASSWORD
    entity_type     VARCHAR(50),                    -- credential, vault, tag, master_data
    entity_id       UUID,
    ip_address      INET,
    user_agent      TEXT,
    details         JSONB,                          -- detail tambahan dalam format JSON
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);