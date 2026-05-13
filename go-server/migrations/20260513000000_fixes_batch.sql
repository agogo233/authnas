-- AuthNas Database Migration
-- Version: 20260513000000_fixes_batch
-- Adds: audit_log table, user.must_change_password, client.previous_client_secret

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_log (
    id TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    event_type TEXT NOT NULL,
    user_id TEXT,
    username TEXT,
    client_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    success INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata TEXT
);

CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);

-- 用户表增加 must_change_password 字段
ALTER TABLE user ADD COLUMN must_change_password INTEGER NOT NULL DEFAULT 0;

-- 客户端表增加 previous_client_secret 和 client_secret_rotated_at 字段
ALTER TABLE client ADD COLUMN previous_client_secret TEXT;
ALTER TABLE client ADD COLUMN client_secret_rotated_at DATETIME;
