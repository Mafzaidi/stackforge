-- +migrate Up
SET search_path TO stackforge_service;

INSERT INTO master_data (id, module, type, name, description, icon, color, sort_order, is_active, metadata, created_at, updated_at)
VALUES
  (gen_random_uuid(), 'navigation', 'sidebar_menu', 'Dashboard', 'Main dashboard', 'layout-dashboard', '', 1, true, '{"url": "/dashboard", "required_permission": null, "parent_id": null}', NOW(), NOW()),
  (gen_random_uuid(), 'navigation', 'sidebar_menu', 'Credentials', 'Credential manager', 'key-round', '', 2, true, '{"url": "/credentials", "required_permission": null, "parent_id": null}', NOW(), NOW()),
  (gen_random_uuid(), 'navigation', 'sidebar_menu', 'Todo List', 'Todo list manager', 'list-todo', '', 3, true, '{"url": "/todolist", "required_permission": "todo.read", "parent_id": null}', NOW(), NOW())
ON CONFLICT (module, type, name) DO NOTHING;
