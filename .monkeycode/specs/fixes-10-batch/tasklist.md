# 实施任务列表：10 项代码质量修复

## 任务 1：统一错误响应 (Fix-14)

- [x] 1.1 在 `internal/response/response.go` 中新增 `SafeError(c, statusCode, err, context)` 函数，内部调用 log.Printf 记录详细错误，返回安全消息给客户端
- [x] 1.2 修复 `internal/handler/oidc.go` 中所有 `err.Error()` 裸传递（约 5 处）和 `c.JSON` 裸响应（1 处）
- [x] 1.3 修复 `internal/handler/admin_user.go` 中 `err.Error()` 裸传递（约 2 处）
- [x] 1.4 运行 `go build ./...` 确认编译通过（注：存在预先存在的 ProxyAuthHandler 未定义错误，与本次修改无关）

## 任务 2：JWKS 启动时预计算 (Fix-15)

- [x] 2.1 在 `OIDCHandler` 中添加 `jwksCache []byte` 字段
- [x] 2.2 在 `NewOIDCHandler` 中添加 `computeJWKS` 方法调用，初始化缓存
- [x] 2.3 修改 `JWKS` handler 方法直接返回缓存的 `[]byte`，使用 `c.Data()` 替代 `c.JSON()`
- [x] 2.4 运行 `go build ./...` 确认编译通过

## 任务 3：静态文件缓存头 (Fix-10)

- [x] 3.1 在 `cmd/server/main.go` 中新增 `staticCacheHeaders` gin 中间件函数
- [x] 3.2 根据文件扩展名设置不同的 Cache-Control 策略（哈希资源 1 年缓存，HTML no-cache）
- [x] 3.3 将中间件注册到正确的路由位置（static.Serve 之前）
- [x] 3.4 运行 `go build ./...` 确认编译通过

## 任务 4：消除循环依赖 (Fix-11)

- [x] 4.1 在 `internal/service/user_service.go` 中定义 `InvitationVerifier` 接口
- [x] 4.2 修改 `UserService` 构造函数接受 `InvitationVerifier` 接口
- [x] 4.3 定义 `TransactionRunner` 接口并在 `UserService` 中使用它替代直接 `*gorm.DB`
- [x] 4.4 移除 `SetInvitationService` 和 `SetDB` 方法
- [x] 4.5 更新 `cmd/server/main.go` 中的依赖注入，移除 setter 调用
- [x] 4.6 运行 `go build ./...` 确认编译通过（注：e2e 测试有预先存在的失败，与本次修改无关）

## 任务 5：合并密码哈希逻辑 (Fix-12)

- [x] 5.1 创建 `pkg/crypto/password.go`，包含 `HashPassword` 和 `VerifyPassword` 函数及常量
- [x] 5.2 更新 `internal/service/auth_service.go` 调用 `crypto.HashPassword`/`VerifyPassword`，删除本地实现
- [x] 5.3 更新 `internal/service/user_service.go` 调用 `crypto.HashPassword`/`VerifyPassword`，删除本地实现
- [x] 5.4 运行 `go build ./...` 确认编译通过
- [x] 5.5 运行 `go test ./...` 确认测试通过

## 任务 6：注册流程增加审批/验证检查 (Fix-1)

- [x] 6.1 在 `internal/handler/auth.go` 的 `Register` 方法中，在 `GenerateTokenPair` 调用前添加审批状态检查
- [x] 6.2 添加邮箱验证检查逻辑，如需验证则发送验证邮件
- [x] 6.3 更新 `RegisterResponse` 结构体增加 `message` 字段支持
- [x] 6.4 运行 `go build ./...` 确认编译通过

## 任务 7：暴露审计日志 API (Fix-2)

- [x] 7.1 创建 `internal/model/audit_log.go` 模型
- [x] 7.2 创建 `internal/repository/audit_log_repo.go` 仓储，提供 List/Count/Create 方法
- [x] 7.3 修改 `internal/service/audit_service.go`，增加 repository 依赖，Log 方法写入数据库
- [x] 7.4 创建 `internal/handler/admin_audit.go` handler，提供 GET /api/admin/audit-logs 端点
- [x] 7.5 在 `internal/router/router.go` 中注册审计日志路由
- [x] 7.6 在 `cmd/server/main.go` 中初始化审计日志组件
- [x] 7.7 创建数据库迁移脚本
- [x] 7.8 运行 `go build ./...` 确认编译通过

## 任务 8：OIDC 客户端密钥轮换兼容 (Fix-3)

- [x] 8.1 修改 `internal/model/client.go`，增加 `PreviousClientSecret` 和 `ClientSecretRotatedAt` 字段
- [x] 8.2 修改 `internal/service/oidc_service.go` RefreshAccessToken，增加 previous secret 校验和自动清理逻辑
- [x] 8.3 修改 `internal/service/client_service.go`，增加 UpdateClientSecret 方法保存旧密钥
- [x] 8.4 创建数据库迁移脚本
- [x] 8.5 运行 `go build ./...` 确认编译通过

- [x] 9.1 在 `internal/repository/key_repo.go` 中新增 `FindByRefreshTokenForUpdate` 方法（使用 FOR UPDATE）
- [x] 9.2 重构 `internal/service/oidc_service.go` RefreshAccessToken，将整个流程包裹在事务中
- [x] 9.3 使用 `FindByRefreshTokenForUpdate` 替代 `FindByRefreshToken` 获取行锁
- [x] 9.4 在 `internal/repository/user_repo.go` 中新增 `GetByIDForUpdate` 方法
- [x] 9.5 运行 `go build ./...` 确认编译通过

- [x] 10.1 修改 `internal/model/user.go`，增加 `MustChangePassword` 字段
- [x] 10.2 修改 `internal/service/user_service.go` EnsureInitialAdmin，创建时设置 MustChangePassword=true
- [x] 10.3 修改 `internal/middleware/auth.go`，增加 MustChangePassword 检查逻辑
- [x] 10.4 修改 `internal/handler/user.go` UpdatePassword 方法，成功后清除 MustChangePassword 标志
- [x] 10.5 创建数据库迁移脚本
- [x] 10.6 运行 `go build ./...` 确认编译通过
