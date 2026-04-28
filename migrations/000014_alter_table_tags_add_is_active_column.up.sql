-- +migrate Up
SET search_path TO stackforge_service;

ALTER TABLE tags
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;