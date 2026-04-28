-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS user_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL UNIQUE,   -- ID dari SSO Authorizer
    encryption_key_hash VARCHAR(512),      -- hash dari encryption key (untuk verifikasi)
    master_password_hash VARCHAR(512),              -- opsional: master password tambahan untuk vault
    password_salt   VARCHAR(255),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

-- Update updated_at automatically
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_profiles_timestamp
BEFORE UPDATE ON user_profiles
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();
