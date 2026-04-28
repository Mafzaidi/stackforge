-- +migrate Down
SET search_path TO stackforge_service;

ALTER TABLE tags
DROP COLUMN is_active;