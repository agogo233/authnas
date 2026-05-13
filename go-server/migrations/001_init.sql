-- AuthNas Database Migration
-- Version: 001_init

-- 用户表
CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE,
    username TEXT UNIQUE NOT NULL,
    name TEXT,
    password_hash TEXT,
    email_verified INTEGER NOT NULL DEFAULT 0,
    approved INTEGER NOT NULL DEFAULT 0,
    mfa_required INTEGER NOT NULL DEFAULT 0,
    token_version INTEGER NOT NULL DEFAULT 0,
    expires_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- 分组表
CREATE TABLE IF NOT EXISTS `groups` (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- 用户分组关联表
CREATE TABLE IF NOT EXISTS user_group (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES `groups`(id) ON DELETE CASCADE
);

-- OIDC 客户端表
CREATE TABLE IF NOT EXISTS client (
    id TEXT PRIMARY KEY,
    client_id TEXT UNIQUE NOT NULL,
    client_secret TEXT,
    name TEXT NOT NULL,
    logo_uri TEXT,
    redirect_uris TEXT NOT NULL,
    post_logout_redirect_uris TEXT,
    grant_types TEXT,
    response_types TEXT,
    scopes TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Consent 表
CREATE TABLE IF NOT EXISTS consent (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    client_id TEXT NOT NULL,
    scopes TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES client(id) ON DELETE CASCADE
);

-- Passkey 表
CREATE TABLE IF NOT EXISTS passkey (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT,
    credential_id TEXT UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    attestation_type TEXT,
    transports TEXT,
    last_used_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- Passkey 认证选项临时表（存储 WebAuthn challenge）
CREATE TABLE IF NOT EXISTS passkey_auth_options (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    challenge TEXT NOT NULL,
    options TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);

-- TOTP 表
CREATE TABLE IF NOT EXISTS totp (
    id TEXT PRIMARY KEY,
    user_id TEXT UNIQUE NOT NULL,
    secret TEXT NOT NULL,
    issuer TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- 邀请表
CREATE TABLE IF NOT EXISTS invitation (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    username TEXT,
    code TEXT UNIQUE NOT NULL,
    scopes TEXT,
    group_id TEXT,
    max_uses INTEGER,
    used_count INTEGER NOT NULL DEFAULT 0,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    created_by TEXT,
    FOREIGN KEY (group_id) REFERENCES _group(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES user(id) ON DELETE SET NULL
);

-- OIDC Payload 表（存储 OIDC 会话数据）
CREATE TABLE IF NOT EXISTS oidc_payload (
    id TEXT PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    payload TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL
);

-- 密钥表（存储 Refresh Token 和会话密钥）
CREATE TABLE IF NOT EXISTS key (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    client_id TEXT,
    token_version INTEGER NOT NULL DEFAULT 1,
    refresh_token_hash TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- 邮箱验证表
CREATE TABLE IF NOT EXISTS email_verification (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    email TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- 密码重置表
CREATE TABLE IF NOT EXISTS password_reset (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    code TEXT UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

-- Proxy Auth 表
CREATE TABLE IF NOT EXISTS proxy_auth (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    proxy_url TEXT NOT NULL,
    header_name TEXT NOT NULL,
    scopes TEXT,
    group_id TEXT,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (group_id) REFERENCES `groups`(id) ON DELETE SET NULL
);

-- 邮件日志表
CREATE TABLE IF NOT EXISTS email_log (
    id TEXT PRIMARY KEY,
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    template TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    created_at DATETIME NOT NULL
);

-- ====================
-- Indexes
-- ====================

-- 过期数据清理必需索引
CREATE INDEX IF NOT EXISTS idx_key_expires_at ON key(expires_at);
CREATE INDEX IF NOT EXISTS idx_passkey_auth_options_expires_at ON passkey_auth_options(expires_at);
CREATE INDEX IF NOT EXISTS idx_oidc_payload_expires_at ON oidc_payload(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verification_expires_at ON email_verification(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_expires_at ON password_reset(expires_at);
CREATE INDEX IF NOT EXISTS idx_invitation_expires_at ON invitation(expires_at);

-- 常用查询优化索引
CREATE INDEX IF NOT EXISTS idx_user_email ON user(email);
CREATE INDEX IF NOT EXISTS idx_user_username ON user(username);
CREATE INDEX IF NOT EXISTS idx_user_group_user_id ON user_group(user_id);
CREATE INDEX IF NOT EXISTS idx_user_group_group_id ON user_group(group_id);
CREATE INDEX IF NOT EXISTS idx_consent_user_id ON consent(user_id);
CREATE INDEX IF NOT EXISTS idx_consent_user_client ON consent(user_id, client_id);
CREATE INDEX IF NOT EXISTS idx_passkey_user_id ON passkey(user_id);
CREATE INDEX IF NOT EXISTS idx_invitation_email ON invitation(email);
