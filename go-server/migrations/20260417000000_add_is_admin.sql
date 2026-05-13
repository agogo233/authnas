-- Add is_admin field to user table
-- Version: 20260417000000

ALTER TABLE user ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0;
