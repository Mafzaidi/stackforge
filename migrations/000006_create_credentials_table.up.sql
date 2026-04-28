-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL,          -- ID dari SSO Authorizer
    vault_id        UUID REFERENCES vaults(id) ON DELETE SET NULL,
    category_id     UUID REFERENCES master_data(id) ON DELETE SET NULL,

    -- Informasi akun (terenkripsi)
    title           VARCHAR(255) NOT NULL,          -- Nama/label (e.g., "Gmail Pribadi")
    site_url        TEXT,                           -- URL website
    favicon_url     TEXT,                           -- cached favicon URL

    -- Credential data (ENCRYPTED with user's encryption key)
    username_encrypted   TEXT NOT NULL,             -- username/email (encrypted)
    password_encrypted   TEXT NOT NULL,             -- password (encrypted)

    -- Metadata
    notes_encrypted TEXT,                           -- catatan tambahan (encrypted)
    is_favorite     BOOLEAN DEFAULT FALSE,
    password_strength INTEGER,                      -- skor kekuatan password (0-100)
    last_used_at    TIMESTAMP,
    password_changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at      TIMESTAMP,                      -- tanggal kedaluwarsa password (opsional)
    auto_login      BOOLEAN DEFAULT FALSE,

    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_credentials_user_id ON credentials(user_id);
CREATE INDEX idx_credentials_vault_id ON credentials(vault_id);
CREATE INDEX idx_credentials_category_id ON credentials(category_id);
CREATE INDEX idx_credentials_title ON credentials(title);
CREATE INDEX idx_credentials_is_favorite ON credentials(user_id, is_favorite);
CREATE INDEX idx_credentials_last_used ON credentials(user_id, last_used_at DESC);

CREATE TRIGGER update_credentials_timestamp
BEFORE UPDATE ON credentials
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();
