-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS master_data (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module          VARCHAR(100) NOT NULL,           -- modul: credential_manager, todo, dll.
    type            VARCHAR(100) NOT NULL,           -- jenis: category, dll.
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    icon            VARCHAR(50),                     -- icon identifier (e.g., 'fa-globe')
    color           VARCHAR(7),                      -- hex color (#FF5733)
    sort_order      INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,            -- soft-disable tanpa hapus
    metadata        JSONB,                           -- data tambahan fleksibel
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(module, type, name)                       -- nama unik per modul dan tipe
);

CREATE INDEX idx_master_data_module_type ON master_data(module, type);