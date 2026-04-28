-- +migrate Down
SET search_path TO stackforge_service;

DROP TRIGGER IF EXISTS update_user_profiles_timestamp ON user_profiles;
DROP FUNCTION IF EXISTS update_timestamp();
DROP TABLE IF EXISTS user_profiles;
