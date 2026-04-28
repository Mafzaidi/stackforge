-- +migrate Up
SET search_path TO stackforge_service;

INSERT INTO master_data (id, module, type, name, description, icon, color, sort_order, is_active, metadata, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'credential', 'category', 'Email', 'Email Credential', null, '', 1, true, null, NOW(), NOW()),
  (gen_random_uuid(), 'credential', 'category', 'Social Media', 'Social Media Credential', null, '', 2, true, null, NOW(), NOW()),
  (gen_random_uuid(), 'credential', 'category', 'Key API', 'Key API Credential', null, '', 3, true, null, NOW(), NOW())
ON CONFLICT (module, type, name) DO NOTHING;
