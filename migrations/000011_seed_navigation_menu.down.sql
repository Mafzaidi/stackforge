-- +migrate Down
SET search_path TO stackforge_service;

DELETE FROM master_data
WHERE module = 'navigation' AND type = 'sidebar_menu';
