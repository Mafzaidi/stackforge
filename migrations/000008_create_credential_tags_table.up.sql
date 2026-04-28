-- +migrate Up
SET search_path TO stackforge_service;

CREATE TABLE IF NOT EXISTS credential_tags (
    credential_id   UUID NOT NULL REFERENCES credentials(id) ON DELETE CASCADE,
    tag_id          UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (credential_id, tag_id)
);

CREATE INDEX idx_credential_tags_credential ON credential_tags(credential_id);
CREATE INDEX idx_credential_tags_tag ON credential_tags(tag_id);