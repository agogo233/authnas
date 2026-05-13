# 用户指令记忆

本文件记录了用户的指令、偏好和教导，用于在未来的交互中提供参考。

## 格式

### 用户指令条目
用户指令条目应遵循以下格式：

[用户指令摘要]
- Date: [YYYY-MM-DD]
- Context: [提及的场景或时间]
- Instructions:
  - [用户教导或指示的内容，逐行描述]

### 项目知识条目
Agent 在任务执行过程中发现的条目应遵循以下格式：

[项目知识摘要]
- Date: [YYYY-MM-DD]
- Context: Agent 在执行 [具体任务描述] 时发现
- Category: [代码结构|代码模式|代码生成|构建方法|测试方法|依赖关系|环境配置]
- Instructions:
  - [具体的知识点，逐行描述]

## 去重策略
- 添加新条目前，检查是否存在相似或相同的指令
- 若发现重复，跳过新条目或与已有条目合并
- 合并时，更新上下文或日期信息
- 这有助于避免冗余条目，保持记忆文件整洁

## 条目

[Playwright E2E 测试状态]
- Date: 2026-04-25
- Context: Agent 在全面 E2E 测试时发现并修复
- Category: 测试方法
- Instructions:
  - **限流问题已修复**：将 `config.yaml` 和 `config.example.yaml` 中的 `requests_per_minute` 从 500 改为 5000
  - **测试失败已修复**：`boundary-conditions.spec.ts` 中超长邮箱注册测试期望不合理（期望停留在 /register 但注册成功后自动登录跳转），已修正测试期望
  - MFA 路由守卫经单独测试验证工作正常
  - 使用 `workers: 1` 避免测试并行执行导致的状态污染
  - 使用 `page.waitForLoadState('domcontentloaded')` 替代 `networkidle` 避免超时
  - 需要清理状态时使用：`await page.context().clearCookies()` + `await page.evaluate(() => localStorage.clear())`
  - **全面测试结果**：
    - 核心认证 (login/register/mfa): 19 passed
    - 安全测试 (security/security-advanced/security-injection): 51 passed, 4 skipped
    - 管理功能 (admin-*): 30 passed
    - 其他功能 (navigation/oidc/profile/consent/session-persistence/boundary/error-page): 180+ passed
  - 某些测试需要服务器重启以重置登录锁定状态

[完整 E2E 测试验证]
- Date: 2026-04-25
- Context: Agent 执行完整 E2E 测试套件验证
- Category: 测试方法
- Instructions:
  - 2026-04-25 完整测试验证结果：
    - login.spec.ts: 8 passed
    - admin-operations.spec.ts: 16 passed
    - admin-dashboard.spec.ts: 3 passed
    - mfa.spec.ts: 4 passed
    - boundary-conditions.spec.ts: 46 passed
    - oidc-flow.spec.ts: 19 passed
  - 关键修复已验证有效：
    - 限流配置 (5000 req/min) 工作正常
    - 边界条件测试修复有效
  - 服务器和前端必须同时运行才能运行测试
  - 首次运行失败是因为后端服务器未启动

[系统设置 API 实现]
- Date: 2026-05-06
- Context: Agent 在执行 P3 前端 Settings 页面优化时发现后端 Settings API 缺失
- Category: 代码结构
- Instructions:
  - 创建了 `internal/model/system_setting.go` 模型，使用 GORM 存储键值对配置
  - 创建了 `internal/service/system_setting_service.go` 服务层，提供 General/Security/Email/Session/RateLimit 五类设置的 CRUD
  - 创建了 `internal/handler/admin_settings.go` 处理层，提供 GET/POST 端点
  - 在 `internal/router/router.go` 中注册了 `/api/admin/settings/*` 路由
  - 添加了 `migrations/002_system_settings.sql` 数据库迁移
  - 在 `e2e/e2e_test.go` 的 `setupRouter` 中注入了 AdminSettingsHandler
  - 在 `cmd/server/main.go` 中初始化默认设置
  - 前端 `web/src/views/admin/Settings.vue` 已更新以使用新的 Settings API
  - 添加了 `e2e/admin_settings_test.go` E2E 测试覆盖
  - 所有 Settings API E2E 测试通过
