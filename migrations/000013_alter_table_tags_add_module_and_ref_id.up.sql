-- +migrate Up
SET search_path TO stackforge_service;

ALTER TABLE tags
ADD COLUMN module VARCHAR(255) NOT NULL,
ADD COLUMN ref_id VARCHAR(255) NOT NULL;    