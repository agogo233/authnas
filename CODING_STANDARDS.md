# AuthNas SSO 代码规范文档

本文档定义了 AuthNas SSO 认证系统的代码编写规范，适用于 Go 后端和 Vue 3 前端项目。

---

## 一、项目架构概览

```
AuthNas SSO
├── go-server/          # Go 后端服务
│   ├── cmd/            # 程序入口
│   ├── internal/       # 内部包
│   │   ├── config/     # 配置管理
│   │   ├── handler/    # HTTP 处理层
│   │   ├── middleware/ # 中间件
│   │   ├── model/      # 数据模型
│   │   ├── repository/# 数据访问层
│   │   ├── response/   # 统一响应格式
│   │   └── service/    # 业务逻辑层
│   ├── migrations/     # 数据库迁移
│   ├── pkg/            # 公共包
│   ├── keys/           # 密钥文件
│   └── e2e/            # 端到端测试
└── web/                # Vue 3 前端
    ├── src/
    │   ├── api/        # API 调用模块
    │   ├── assets/     # 静态资源
    │   ├── components/ # Vue 组件
    │   ├── router/     # 路由配置
    │   ├── stores/     # Pinia 状态管理
    │   ├── views/      # 页面视图
    │   └── ...
    ├── e2e/            # Playwright E2E 测试
    └── ...
```

---

## 二、Go 后端规范

### 2.1 项目结构

| 目录/包 | 用途 | 引入规范 |
|---------|------|----------|
| `cmd/` | 程序入口点 | 每个应用一个 `main.go` |
| `internal/` | 私有应用程序代码 | 不能被外部导入 |
| `internal/config/` | 配置加载和验证 | 配置项必须有文档注释 |
| `internal/handler/` | HTTP 处理函数 | 按功能模块拆分文件 |
| `internal/middleware/` | Gin 中间件 | 单一职责，如 `auth.go`、`ratelimit.go` |
| `internal/model/` | 数据模型 | 与数据库表一一对应 |
| `internal/repository/` | 数据访问层 | 接口驱动，依赖 service 层 |
| `internal/service/` | 业务逻辑层 | 复杂业务逻辑存放处 |
| `internal/response/` | 统一响应格式 | 所有 API 响应必须使用此包 |
| `migrations/` | 数据库迁移文件 | 使用 GORM AutoMigrate |
| `e2e/` | 端到端测试 | 测试真实 HTTP 交互 |

### 2.2 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 纯小写单词，简短 | `handler`, `service`, `repository` |
| 文件名 | 小写字母，单词用下划线分隔 | `auth_handler.go`, `user_service.go` |
| 结构体名 | PascalCase | `UserService`, `LoginRequest` |
| 接口名 | PascalCase，以 `er` 结尾 | `UserRepository`, `AuthHandler` |
| 函数名 | PascalCase，动词或动词短语 | `GetUser`, `CreateUser`, `VerifyPassword` |
| 变量名 | 驼峰命名 | `userID`, `accessToken` |
| 常量名 | 全大写，下划线分隔 | `MaxLoginAttempts`, `TokenExpiry` |
| 错误变量 | 以 `Err` 开头 | `ErrUserNotFound`, `ErrInvalidCredentials` |

### 2.3 代码组织

#### Handler 文件拆分原则

按功能模块拆分 `handler` 包，避免 `handler.go` 单文件过大：

```
internal/handler/
├── admin.go           # 管理后台主路由注册（精简版）
├── admin_user.go      # 用户管理
├── admin_group.go     # 群组管理
├── admin_client.go    # OIDC 客户端管理
├── admin_invitation.go # 邀请管理
├── admin_proxy_auth.go # ProxyAuth 配置
├── oidc.go           # OIDC 协议端点
├── auth.go           # 认证相关（登录/注册/密码）
├── user.go           # 用户资料和会话
├── response.go      # 统一响应辅助函数
└── handler_test.go  # 测试文件
```

#### 代码示例：Handler 结构体

```go
type AuthHandler struct {
    authService       *service.AuthService
    userService       *service.UserService
    totpService       *service.TOTPService
    passkeyService    *service.PasskeyService
    invitationService *service.InvitationService
    emailService      *service.EmailService
    cfg               *config.Config
}

func NewAuthHandler(
    cfg *config.Config,
    authService *service.AuthService,
    userService *service.UserService,
    totpService *service.TOTPService,
    passkeyService *service.PasskeyService,
    invitationService *service.InvitationService,
    emailService *service.EmailService,
) *AuthHandler {
    return &AuthHandler{
        authService:       authService,
        userService:       userService,
        totpService:       totpService,
        passkeyService:    passkeyService,
        invitationService: invitationService,
        emailService:      emailService,
        cfg:               cfg,
    }
}
```

### 2.4 错误处理

#### 必须使用 `response` 包

所有 HTTP 错误响应必须使用 `internal/response` 包：

```go
import "github.com/authnas/authnas/go-server/internal/response"

// 正确用法
response.BadRequest(c, "invalid email format")
response.Unauthorized(c, "invalid credentials")
response.Forbidden(c, "access denied")
response.NotFound(c, "user not found")
response.InternalServerError(c, "failed to process request")
response.TooManyRequests(c, "rate limit exceeded")
response.ServiceUnavailable(c, "service temporarily unavailable")
response.MFARequired(c, "MFA verification required")

// 成功响应
response.Success(c, data)
response.SuccessWithMessage(c, "operation completed")
```

#### 内部错误日志

使用 `safeErrorMessage` 模式避免泄露内部错误：

```go
func safeErrorMessage(err error, context string) string {
    log.Printf("[ERROR] %s: %v", context, err)
    return "an internal error occurred"
}
```

### 2.5 请求验证

使用 Gin's binding tag 进行请求体验证：

```go
type LoginRequest struct {
    Input    string `json:"input" binding:"required"`
    Password string `json:"password" binding:"required"`
    Remember bool   `json:"remember"`
}

type RegisterRequest struct {
    Email    string `json:"email" binding:"max=250"`
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
    Name     string `json:"name"`
    InviteID string `json:"inviteId"`
    Challenge string `json:"challenge"`
}
```

### 2.6 路由注册

使用 Gin RouterGroup 组织路由：

```go
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
    auth := r.Group("/auth")
    {
        auth.POST("/login", h.Login)
        auth.POST("/register", h.Register)
        auth.POST("/passkey/start", h.PasskeyStart)
        auth.POST("/passkey/end", h.PasskeyEnd)
        auth.POST("/totp", h.TOTPVerify)
        // ...
    }
}
```

### 2.7 配置管理

使用 Viper 管理配置，配置结构体必须有文档注释：

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Security SecurityConfig
    Email    EmailConfig
    // ...
}

type SecurityConfig struct {
    JWTsecret          string
    TokenExpiry        time.Duration
    RefreshTokenExpiry time.Duration
    MFARequired        bool
    PasswordStrength   int
    PasswordMinLength  int
    EmailVerification  bool
    SignupRequiresApproval bool
}
```

### 2.8 测试规范

E2E 测试文件命名：`*_e2e_test.go`

```go
func TestE2E_Auth_Login(t *testing.T) {
    // 测试逻辑
}

func TestE2E_Auth_Register(t *testing.T) {
    // 测试逻辑
}
```

---

## 三、Vue 3 前端规范

### 3.1 项目结构

```
web/src/
├── api/                 # API 调用模块
│   ├── index.ts        # 统一导出
│   ├── auth.ts         # 认证相关 API
│   └── admin/          # 管理后台 API
│       ├── index.ts
│       ├── users.ts
│       ├── groups.ts
│       ├── clients.ts
│       ├── invitations.ts
│       └── proxyauth.ts
├── assets/             # 静态资源
├── components/          # 公共组件
│   ├── common/         # 通用组件
│   └── layout/         # 布局组件
├── router/             # 路由配置
├── stores/             # Pinia 状态管理
│   ├── auth.ts
│   └── ...
├── types/              # TypeScript 类型定义
├── utils/              # 工具函数
├── views/              # 页面视图
│   ├── login/
│   ├── admin/
│   └── user/
├── App.vue
└── main.ts
```

### 3.2 API 模块化

按功能模块拆分 API 调用，禁止在组件中直接使用 Axios：

```
api/
├── index.ts           # 统一导出所有 API
├── auth.ts           # 认证相关（登录/注册/MFA）
├── user.ts           # 用户资料相关
└── admin/            # 管理后台 API
    ├── index.ts      # 统一导出 admin 模块
    ├── users.ts      # 用户管理
    ├── groups.ts     # 群组管理
    ├── clients.ts    # 客户端管理
    ├── invitations.ts # 邀请管理
    └── proxyauth.ts  # ProxyAuth 配置
```

### 3.3 API 调用示例

```typescript
// api/admin/users.ts
import axios from '@/utils/request'

export interface User {
  id: string
  username: string
  email?: string
  name?: string
  isAdmin: boolean
  approved: boolean
}

export const adminUserAPI = {
  list() {
    return axios.get<User[]>('/api/admin/users')
  },

  get(id: string) {
    return axios.get<User>(`/api/admin/users/${id}`)
  },

  create(data: { username: string; email?: string; password: string }) {
    return axios.post<User>('/api/admin/users', data)
  },

  update(id: string, data: Partial<User>) {
    return axios.put<User>(`/api/admin/users/${id}`, data)
  },

  delete(id: string) {
    return axios.delete(`/api/admin/users/${id}`)
  },

  resetPassword(id: string) {
    return axios.post(`/api/admin/users/${id}/reset-password`)
  },
}
```

```typescript
// api/admin/index.ts
export { adminUserAPI } from './users'
export { adminGroupAPI } from './groups'
export { adminClientAPI } from './clients'
export { adminInvitationAPI } from './invitations'
export { adminProxyAuthAPI } from './proxyauth'
```

```typescript
// api/index.ts
export * from './auth'
export * from './user'
export * from './admin'
```

### 3.4 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 文件名 | 小写字母，单词用连字符分隔 | `user-profile.vue`, `api-request.ts` |
| 组件名 | PascalCase | `UserProfile.vue`, `LoginForm.vue` |
| 变量名 | 驼峰命名 | `userId`, `accessToken` |
| 常量名 | 全大写，下划线分隔 | `MaxLoginAttempts`, `API_BASE_URL` |
| 接口名 | PascalCase，可加 `I` 前缀 | `User`, `IUserResponse` |
| 函数名 | 驼峰命名，动词开头 | `fetchUser`, `updateUser` |
| Store 名 | 驼峰命名 | `useAuthStore`, `useUserStore` |
| 类型名 | PascalCase | `LoginRequest`, `UserResponse` |

### 3.5 TypeScript 类型定义

类型定义文件放在 `types/` 目录或内联在 API 文件中：

```typescript
// types/user.ts
export interface User {
  id: string
  username: string
  email?: string
  name?: string
  emailVerified: boolean
  approved: boolean
  isAdmin: boolean
  mfaRequired: boolean
  createdAt: string
  updatedAt?: string
  expiresAt?: string
}

export interface LoginRequest {
  input: string
  password: string
  remember?: boolean
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
  user: User
}
```

### 3.6 Pinia Store 规范

```typescript
// stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, LoginRequest } from '@/api'
import { authAPI } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)

  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.isAdmin ?? false)

  async function login(credentials: LoginRequest) {
    const response = await authAPI.login(credentials)
    accessToken.value = response.data.accessToken
    refreshToken.value = response.data.refreshToken
    user.value = response.data.user
    // 持久化到 localStorage
  }

  async function logout() {
    // 调用 logout API
    // 清除状态
  }

  return {
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    isAdmin,
    login,
    logout,
  }
})
```

### 3.7 组件规范

```vue
<script setup lang="ts">
import { ref } from 'vue'
import type { User } from '@/types'

interface Props {
  user: User
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update', user: User): void
  (e: 'delete', userId: string): void
}>()

const isEditing = ref(false)

function handleSubmit() {
  // 处理提交
}
</script>

<template>
  <div class="user-card">
    <!-- 组件内容 -->
  </div>
</template>

<style scoped>
.user-card {
  /* 样式 */
}
</style>
```

---

## 四、API 响应格式

### 4.1 成功响应

```json
{
  "success": true,
  "data": { ... }
}
```

或带消息：

```json
{
  "success": true,
  "message": "operation completed",
  "data": { ... }
}
```

### 4.2 错误响应

```json
{
  "success": false,
  "error": "error message"
}
```

### 4.3 分页响应

```json
{
  "success": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 100
  }
}
```

---

## 五、安全规范

### 5.1 认证与授权

- 所有敏感 API 必须要求认证
- 使用 JWT Bearer Token 认证
- Token 必须在 Authorization header 中传递
- 管理员操作必须验证 `isAdmin` 权限

### 5.2 输入验证

**后端**：
- 所有用户输入必须验证
- 使用 Gin's binding tag
- 额外安全验证（XSS、SQL 注入）
- 密码强度检查

**前端**：
- 表单验证
- 类型检查
- 长度限制

### 5.3 敏感数据处理

- 密码永远不返回给前端
- Token 存储在内存或 httpOnly cookie
- 敏感操作需要二次确认

### 5.4 CORS 配置

```go
config := cors.DefaultConfig()
config.AllowOrigins = []string{"http://localhost:3000"}
config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
config.AllowHeaders = []string{"Authorization", "Content-Type"}
```

---

## 六、Git 提交规范

### 6.1 Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 6.2 Type 类型

| Type | 描述 |
|------|------|
| feat | 新功能 |
| fix | Bug 修复 |
| docs | 文档更新 |
| style | 代码格式（不影响功能） |
| refactor | 重构 |
| test | 测试相关 |
| chore | 构建/工具链 |

### 6.3 示例

```
feat(auth): add MFA TOTP verification

Add TOTP-based multi-factor authentication support including:
- TOTP registration with QR code generation
- TOTP verification during login
- TOTP management in user settings

Closes #123
```

---

## 七、测试规范

### 7.1 后端 E2E 测试

```bash
# 运行所有 E2E 测试
cd go-server/e2e && go test -v ./...

# 运行特定模块
go test -v -run "TestE2E_Auth"

# 运行带覆盖率
go test -v -coverprofile=coverage.out ./...
```

### 7.2 前端测试

```bash
# 运行单元测试
cd web && npm run test

# 运行 E2E 测试
npm run test:e2e

# 代码检查
npm run lint
```

### 7.3 E2E 测试场景

参考 `go-server/e2e/TEST_SCENARIOS_COMPLETE.md` 了解完整的测试场景清单。

---

## 八、文档更新记录

| 版本 | 日期 | 描述 |
|------|------|------|
| 1.0.0 | 2026-04-23 | 初始版本，包含 Go 后端和 Vue 3 前端规范 |
