-- +migrate Down
SET search_path TO stackforge_service;

DELETE FROM master_data
WHERE module = 'credential' AND type = 'category';