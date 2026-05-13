# AuthNas 接口文档

## 概述

本文档详细描述 AuthNas 系统的所有 HTTP API 接口，包括认证接口、用户管理接口、管理后台接口和 OIDC 协议接口。所有接口均返回 JSON 格式数据。

## 基础信息

- **基础路径**: `/api`
- **OIDC 基础路径**: `/oidc`
- **默认端口**: `8080`（后端）
- **内容类型**: `application/json`

## 认证接口

### POST /api/auth/login

用户登录接口。

**请求体**:
```json
{
  "input": "用户名或邮箱",
  "password": "密码",
  "remember": false
}
```

**响应（成功）**:
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应（MFA 需要）**:
```json
{
  "success": false,
  "mfa_required": true,
  "message": "MFA verification required"
}
```

**响应（失败）**:
```json
{
  "success": false,
  "message": "invalid credentials"
}
```

### POST /api/auth/register

用户注册接口。

**请求体**:
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password123",
  "name": "显示名称",
  "invite_id": "邀请ID（可选）",
  "challenge": "邀请挑战（可选）"
}
```

**响应**:
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### POST /api/auth/passkey/start

开始 Passkey 认证。

**请求体**:
```json
{
  "username": "用户名（可选）"
}
```

**响应**:
```json
{
  "success": true,
  "challenge": "随机挑战字符串",
  "options": "WebAuthn 选项 JSON 字符串"
}
```

### POST /api/auth/passkey/end

完成 Passkey 认证。

**请求体**:
```json
{
  "credential_id": "凭证ID",
  "challenge": "挑战字符串",
  "response": "WebAuthn 响应 JSON"
}
```

**响应**:
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### POST /api/auth/totp

验证 TOTP（多因素认证）码。

**请求头**: `Authorization: Bearer <access_token>`

**请求体**:
```json
{
  "challenge": "挑战字符串",
  "token": "TOTP 验证码"
}
```

**响应**:
```json
{
  "success": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### POST /api/auth/verify_email

验证用户邮箱。

**请求体**:
```json
{
  "user_id": "用户ID",
  "challenge": "邮箱验证挑战"
}
```

### POST /api/auth/send_verify_email

发送邮箱验证链接。

**请求体**:
```json
{
  "email": "user@example.com"
}
```

### GET /api/auth/invitation/:id/:challenge

获取邀请信息。

**响应**:
```json
{
  "success": true,
  "email": "invited@example.com",
  "username": "预设用户名"
}
```

### POST /api/auth/forgot_password

请求密码重置。

**请求体**:
```json
{
  "email": "user@example.com"
}
```

**响应**:
```json
{
  "success": true,
  "message": "If an account with that email exists, a password reset link has been sent"
}
```

### POST /api/auth/reset_password

重置密码。

**请求体**:
```json
{
  "code": "重置代码",
  "new_password": "新密码"
}
```

## 用户接口

### GET /api/user/profile

获取当前用户资料。

**请求头**: `Authorization: Bearer <access_token>`

**响应**:
```json
{
  "id": "用户ID",
  "email": "user@example.com",
  "username": "username",
  "name": "显示名称",
  "email_verified": true,
  "approved": true,
  "is_admin": false,
  "mfa_required": false,
  "created_at": "2024-01-01T00:00:00Z",
  "groups": [
    {
      "id": "组ID",
      "name": "组名"
    }
  ]
}
```

### PUT /api/user/profile

更新用户资料。

**请求头**: `Authorization: Bearer <access_token>`

**请求体**:
```json
{
  "name": "新显示名称"
}
```

### PUT /api/user/password

修改密码。

**请求头**: `Authorization: Bearer <access_token>`

**请求体**:
```json
{
  "current_password": "当前密码",
  "new_password": "新密码"
}
```

## 安全设置接口

### GET /api/user/security/passkey

获取用户的 Passkeys 列表。

**请求头**: `Authorization: Bearer <access_token>`

**响应**:
```json
{
  "passkeys": [
    {
      "id": "passkey_id",
      "name": "我的笔记本",
      "created_at": "2024-01-01T00:00:00Z",
      "last_used_at": "2024-06-01T00:00:00Z"
    }
  ]
}
```

### POST /api/user/security/passkey/register/start

开始注册 Passkey。

**响应**:
```json
{
  "challenge": "随机挑战",
  "options": "WebAuthn 选项"
}
```

### POST /api/user/security/passkey/register/end

完成 Passkey 注册。

**请求体**:
```json
{
  "name": "Passkey 名称",
  "response": "WebAuthn 响应"
}
```

### DELETE /api/user/security/passkey/:id

删除 Passkey。

### GET /api/user/security/totp

获取 TOTP 状态。

**响应**:
```json
{
  "enabled": true,
  "otpauth_url": "otpauth://totp/..."
}
```

### POST /api/user/security/totp/enable

启用 TOTP。

**请求体**:
```json
{
  "token": "TOTP 验证码"
}
```

### DELETE /api/user/security/totp

禁用 TOTP。

## 管理接口

### GET /api/admin/users

获取用户列表。

**请求头**: `Authorization: Bearer <access_token>`

**查询参数**:
- `page`: 页码（默认 1）
- `page_size`: 每页数量（默认 20）
- `search`: 搜索关键词

**响应**:
```json
{
  "users": [...],
  "total": 100,
  "page": 1,
  "page_size": 20
}
```

### POST /api/admin/users

创建用户。

**请求体**:
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password",
  "name": "显示名称",
  "is_admin": false,
  "approved": true
}
```

### GET /api/admin/users/:id

获取指定用户详情。

### PUT /api/admin/users/:id

更新用户。

### DELETE /api/admin/users/:id

删除用户。

### POST /api/admin/users/:id/approve

批准用户。

### POST /api/admin/users/:id/send_verify_email

发送验证邮件。

## 客户端管理接口

### GET /api/admin/clients

获取 OIDC 客户端列表。

**响应**:
```json
{
  "clients": [
    {
      "id": "客户端ID",
      "client_id": "oauth2_client_id",
      "name": "应用名称",
      "redirect_uris": "https://app.example.com/callback",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### POST /api/admin/clients

创建 OIDC 客户端。

**请求体**:
```json
{
  "name": "应用名称",
  "redirect_uris": "https://app.example.com/callback",
  "post_logout_redirect_uris": "https://app.example.com",
  "scopes": "openid profile email"
}
```

### GET /api/admin/clients/:id

获取客户端详情。

### PUT /api/admin/clients/:id

更新客户端。

### DELETE /api/admin/clients/:id

删除客户端。

## 用户组管理接口

### GET /api/admin/groups

获取用户组列表。

### POST /api/admin/groups

创建用户组。

**请求体**:
```json
{
  "name": "组名",
  "description": "组描述"
}
```

### PUT /api/admin/groups/:id

更新用户组。

### DELETE /api/admin/groups/:id

删除用户组。

### POST /api/admin/groups/:id/users

添加用户到组。

### DELETE /api/admin/groups/:id/users/:userId

从组中移除用户。

## 邀请管理接口

### GET /api/admin/invitations

获取邀请列表。

### POST /api/admin/invitations

创建邀请。

**请求体**:
```json
{
  "email": "invited@example.com",
  "username": "预设用户名",
  "send_email": true,
  "expires_in": 72
}
```

### DELETE /api/admin/invitations/:id

删除邀请。

## OIDC 接口

### GET /.well-known/openid-configuration

OIDC 发现端点。

**响应**:
```json
{
  "issuer": "http://localhost:8080",
  "authorization_endpoint": "http://localhost:8080/oidc/auth",
  "token_endpoint": "http://localhost:8080/oidc/token",
  "userinfo_endpoint": "http://localhost:8080/oidc/userinfo",
  "jwks_uri": "http://localhost:8080/oidc/jwks",
  "response_types_supported": ["code"],
  "subject_types_supported": ["public"],
  "id_token_signing_alg_values_supported": ["RS256"]
}
```

### GET /oidc/jwks

获取 JSON Web Key Set。

**响应**:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "authnas-key-1",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

### GET /oidc/auth

OIDC 授权端点。

**查询参数**:
- `client_id`: 客户端 ID
- `redirect_uri`: 回调地址
- `response_type`: 响应类型（固定为 code）
- `scope`: 作用域（默认 openid）
- `state`: 状态参数
- `nonce`: 随机数
- `code_challenge`: PKCE 挑战
- `code_challenge_method`: PKCE 方法

### POST /oidc/token

Token 端点。

**请求体**:
- `grant_type`: authorization_code 或 refresh_token
- `code`: 授权码（authorization_code 时）
- `redirect_uri`: 回调地址（authorization_code 时）
- `refresh_token`: 刷新令牌（refresh_token 时）

**响应**:
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
  "id_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

### GET /oidc/userinfo

获取用户信息。

**请求头**: `Authorization: Bearer <access_token>`

**响应**:
```json
{
  "sub": "用户ID",
  "email": "user@example.com",
  "email_verified": true,
  "name": "显示名称",
  "preferred_username": "username"
}
```

### POST /oidc/token/revocation

撤销 Token。

**请求体**:
```
token=<access_token 或 refresh_token>
```

## 公共接口

### GET /api/health

健康检查接口。

**响应**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### GET /api/config

获取公开配置。

**响应**:
```json
{
  "app_name": "AuthNas",
  "oidc_issuer": "http://localhost:8080",
  "available_signup": true,
  "password_reset_enabled": true,
  "email_verification_required": false
}
```

## 错误响应格式

所有接口的错误响应格式如下：

```json
{
  "success": false,
  "message": "错误描述信息"
}
```

或

```json
{
  "error": "error_code",
  "error_description": "详细错误描述"
}
```

## HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或认证失败 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 429 | 请求过于频繁（限流） |
| 500 | 服务器内部错误 |
