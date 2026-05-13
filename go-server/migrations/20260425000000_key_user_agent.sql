-- Add user_agent field to key table
-- Version: 20260425000000

ALTER TABLE key ADD COLUMN user_agent TEXT;
