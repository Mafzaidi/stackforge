-- +migrate Down
SET search_path TO stackforge_service;

DROP TABLE IF EXISTS master_data;
