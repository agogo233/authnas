-- Add GroupNames column to key table for caching user groups
ALTER TABLE key ADD COLUMN group_names TEXT;
