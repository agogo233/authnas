# 技术设计：10 项代码质量修复

## 概述

对 AuthNas 项目进行 10 项修复，涵盖错误响应统一、性能优化、架构改进、功能完善和安全加固。

## 修复清单

### 1. 统一错误响应 (Fix-14)

**问题**: 多处 handler 直接传递 `err.Error()` 到响应，可能泄露内部细节。

**方案**:
- 在 `handler/auth.go` 中已有 `safeErrorMessage` 工具函数
- 将其提取到 `response` 包中作为 `SafeError` 函数
- 统一所有 handler 使用 `response.SafeError(c, httpStatus, err, "context")` 模式
- 修复 `oidc.go`、`admin_user.go` 中的裸 `err.Error()` 调用
- 修复 `oidc.go:165` 的裸 JSON 响应

**修改文件**:
- `internal/response/response.go` — 新增 `SafeError` 函数
- `internal/handler/oidc.go` — 替换所有 `err.Error()` 为 `safeErrorMessage`
- `internal/handler/admin_user.go` — 替换所有 `err.Error()` 为 `safeErrorMessage`

### 2. JWKS 启动时预计算 (Fix-15)

**问题**: OIDC JWKS 端点每次请求都重新计算 base64 编码。

**方案**:
- `OIDCHandler` 中添加 `jwksCache []byte` 字段
- `NewOIDCHandler` 中调用 `oidcService.GetPublicKey()` 预计算 JWKS
- 添加 `computeJWKS` 私有方法生成缓存
- `JWKS handler` 直接返回缓存的 `[]byte`
- 保留 public key 获取接口供未来 key 轮换使用

**修改文件**:
- `internal/handler/oidc.go` — 添加缓存字段和预计算逻辑

### 3. 静态文件缓存头 (Fix-10)

**问题**: 静态资源无 Cache-Control 头。

**方案**:
- 在 `main.go` 中添加 `staticCacheHeaders` 中间件
- 对带哈希的静态资源 (.js, .css, .png, .jpg, .svg, .woff2 等) 设置 `public, max-age=31536000, immutable`
- 对 `index.html` 设置 `no-cache`
- 中间件需要在 static.Serve 之前注册

**修改文件**:
- `cmd/server/main.go` — 添加缓存中间件

### 4. 消除循环依赖 (Fix-11)

**问题**: `UserService` 和 `InvitationService` 通过 setter 方法解决循环依赖。

**方案**:
- 定义 `InvitationVerifier` 接口，仅包含 `VerifyAndConsume` 方法
- `UserService` 构造函数接受 `InvitationVerifier` 接口而非具体类型
- `InvitationService` 实现该接口
- 移除 `SetInvitationService` 和 `SetDB` setter 方法
- `UserService` 中需要直接 DB 操作的逻辑，通过定义 `TransactionRunner` 接口解耦

**修改文件**:
- `internal/service/user_service.go` — 使用接口替代具体类型
- `internal/service/invitation_service.go` — 确认实现接口
- `cmd/server/main.go` — 移除 setter 调用

### 5. 合并密码哈希逻辑 (Fix-12)

**问题**: `auth_service.go` 和 `user_service.go` 中各有一套独立的 argon2id 实现。

**方案**:
- 创建 `pkg/crypto/password.go` 包，包含统一的 `HashPassword` 和 `VerifyPassword`
- 删除 `auth_service.go` 中的 `HashPassword`/`VerifyPassword` 方法及其常量
- 删除 `user_service.go` 中的 `hashPassword`/`verifyPassword` 方法及其常量
- 两处改为调用 `crypto.HashPassword` / `crypto.VerifyPassword`
- `AuthService` 不再暴露 `HashPassword`/`VerifyPassword` 方法（如外部有调用则保留委托）

**修改文件**:
- `pkg/crypto/password.go` — 新建，统一实现
- `internal/service/auth_service.go` — 删除重复实现
- `internal/service/user_service.go` — 删除重复实现

### 6. 注册流程增加审批/验证检查 (Fix-1)

**问题**: 邀请注册路径 (`req.InviteID != ""`) 在创建用户后直接颁发 token，未检查 `Approved` 和 `EmailVerified` 状态。

**方案**:
- 在 `auth.go:362` (GenerateTokenPair 调用前) 添加检查:
  - 如果 `isSignupRequiresApproval()` 且 `!user.Approved`，返回 403 + "account pending approval"
  - 如果 `isEmailVerificationRequired()` 且 `!user.EmailVerified`，发送验证邮件并返回 "verification email sent"
- 两种情况都不颁发 token
- 非邀请注册路径已有 `isSignupRequiresApproval()` 检查，保持不变

**修改文件**:
- `internal/handler/auth.go` — Register 方法中添加检查

### 7. 暴露审计日志 API (Fix-2)

**问题**: `AuditService` 仅输出到 stdout，无 API 暴露。

**方案**:
- 新增 `AuditLog` 模型，包含 id, timestamp, event_type, user_id, username, client_id, ip_address, user_agent, success, error_message, metadata (JSON)
- 新增 `AuditLogRepository`，提供 `List` (分页), `Count`, `Create` 方法
- 修改 `AuditService`，增加 repository 依赖，`Log` 方法改为写入数据库 (保留 stdout 作为降级)
- 新增 `AdminAuditHandler`，提供 `GET /api/admin/audit-logs` 端点
- 路由注册到 admin 路由组

**新增文件**:
- `internal/model/audit_log.go` — 审计日志模型
- `internal/repository/audit_log_repo.go` — 审计日志仓储
- `internal/handler/admin_audit.go` — 审计日志 handler

**修改文件**:
- `internal/service/audit_service.go` — 增加 DB 写入能力
- `internal/router/router.go` — 注册审计日志路由
- `cmd/server/main.go` — 初始化审计日志组件
- `internal/database/migrations/` — 新增迁移脚本

### 8. OIDC 客户端密钥轮换兼容 (Fix-3)

**问题**: 客户端密钥轮换后，现有 refresh token 立即失效。

**方案**:
- `Client` 模型增加 `PreviousClientSecret *string` 和 `ClientSecretRotatedAt *time.Time` 字段
- `RefreshAccessToken` 校验时，同时检查 `ClientSecret` 和 `PreviousClientSecret`
- 如果 `ClientSecretRotatedAt` 超过 24 小时，自动清理 `PreviousClientSecret`
- 更新客户端密钥时 (admin 接口)，将旧密钥存入 `PreviousClientSecret` 并记录轮换时间
- 数据库迁移增加这两个字段

**修改文件**:
- `internal/model/client.go` — 增加字段
- `internal/service/oidc_service.go` — RefreshAccessToken 增加 previous secret 校验
- `internal/service/client_service.go` — UpdateSecret 方法保存旧密钥
- `internal/handler/admin_client.go` — 更新密钥逻辑
- `internal/database/migrations/` — 新增迁移

### 9. Token 刷新并发竞态 (Fix-4)

**问题**: `RefreshAccessToken` 中 `FindByRefreshToken` 不加锁，并发刷新可能导致同一 token 被使用两次。

**方案**:
- `KeyRepository` 新增 `FindByRefreshTokenForUpdate` 方法，使用 `SELECT ... FOR UPDATE`
- `OIDCService.RefreshAccessToken` 改用此方法获取行锁
- 将验证和 token 创建逻辑放在同一个事务中
- 使用 `s.db.Transaction` 包裹整个 refresh 流程

**修改文件**:
- `internal/repository/key_repo.go` — 新增 `FindByRefreshTokenForUpdate`
- `internal/service/oidc_service.go` — RefreshAccessToken 重构为事务

### 10. 初始管理员密码强制修改 (Fix-5)

**问题**: 初始管理员密码设置后无首次强制修改机制。

**方案**:
- `User` 模型增加 `MustChangePassword bool` 字段，默认 false
- `EnsureInitialAdmin` 创建管理员时设为 true
- `authMiddleware` 在验证 token 后检查此标志:
  - 如果为 true，检查请求是否为修改密码接口
  - 如果不是，返回 403 + code "MUST_CHANGE_PASSWORD"
- 修改密码成功后清除该标志
- 前端需要处理此错误码并引导用户修改密码

**修改文件**:
- `internal/model/user.go` — 增加 MustChangePassword 字段
- `internal/service/user_service.go` — EnsureInitialAdmin 设置标志
- `internal/middleware/auth.go` — 增加检查逻辑
- `internal/handler/auth.go` — ChangePassword 清除标志
- `internal/database/migrations/` — 新增迁移

## 数据库迁移

需要新增以下迁移:
1. `audit_logs` 表创建
2. `client` 表增加 `previous_client_secret` 和 `client_secret_rotated_at`
3. `user` 表增加 `must_change_password`

## 风险与注意事项

1. 密码哈希统一后需要确保所有调用点正确迁移
2. 审计日志写入可能影响性能，需确保异步或非阻塞
3. 客户端密钥轮换的 24h 宽限期需要在文档中说明
4. MustChangePassword 不应阻断 OIDC 授权流程（第三方客户端回调）
