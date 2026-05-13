# AuthNas 开发者指南

## 项目目的

AuthNas 是一个开源的 SSO 认证和用户管理系统，为自托管应用提供统一的身份认证服务。它通过标准化的 OIDC 协议实现单点登录，支持多种认证方式（密码、Passkeys、TOTP）。

**核心职责**:
- 作为 OIDC 身份提供者（Identity Provider）
- 用户生命周期管理（注册、认证、资料管理）
- 多因素认证支持
- 客户端应用管理

## 环境搭建

### 前置条件

- Go 1.25 或更高版本
- Node.js 18+ 和 npm
- Git

### 安装步骤

1. 克隆仓库

```bash
git clone https://github.com/authnas/authnas.git
cd authnas
```

2. 配置后端

```bash
cd go-server

# 创建配置目录
mkdir -p config

# 创建配置文件
cat > config/config.yaml << EOF
app:
  url: http://localhost:8080
  name: AuthNas
  environment: development

database:
  path: ./data/authnas.db

security:
  storage_key: your-storage-key-here
  password_strength: 3
  password_min_length: 8
  initial_admin_username: admin
  initial_admin_email: admin@example.com
  initial_admin_password: your-admin-password

jwt:
  access_token_expiry: 15m
  refresh_token_expiry: 168h

rate_limit:
  enabled: true
  requests_per_minute: 60
EOF
```

3. 配置前端

```bash
cd ../web

# 安装依赖
npm install
```

### 环境变量

AuthNas 通过配置文件（YAML）进行配置。主要配置项如下：

| 变量 | 必需 | 描述 | 默认值 |
|------|------|------|--------|
| `app.url` | 是 | 应用 URL | `http://localhost:8080` |
| `app.name` | 否 | 应用名称 | `AuthNas` |
| `database.path` | 否 | SQLite 数据库路径 | `./data/authnas.db` |
| `security.storage_key` | 是 | 存储加密密钥 | - |
| `security.initial_admin_password` | 是 | 初始管理员密码 | - |
| `email.enabled` | 否 | 启用邮件功能 | `false` |
| `email.smtp_host` | 否 | SMTP 服务器地址 | - |
| `email.smtp_port` | 否 | SMTP 端口 | `587` |
| `jwt.access_token_expiry` | 否 | Access Token 过期时间 | `15m` |
| `jwt.refresh_token_expiry` | 否 | Refresh Token 过期时间 | `168h` |

### 运行

**开发模式运行后端**：

```bash
cd go-server
go run cmd/server/main.go
```

**开发模式运行前端**：

```bash
cd web
npm run dev
```

前端默认运行在 `http://localhost:3000`，会自动代理 API 请求到后端 `http://localhost:8080`。

### 首次启动

首次启动后，系统会自动：
1. 创建数据库和必要的表
2. 创建初始管理员账户

初始管理员登录信息：
- 用户名：由 `security.initial_admin_username` 配置
- 邮箱：由 `security.initial_admin_email` 配置
- 密码：由 `security.initial_admin_password` 配置

**重要**：首次登录后请立即修改管理员密码或创建新的管理员账户。

## 开发工作流

### 代码质量工具

**Go 后端**：

| 工具 | 用途 |
|------|------|
| `go fmt` | 代码格式化 |
| `go vet` | 代码检查 |
| `go test` | 运行测试 |

```bash
# 格式化代码
cd go-server && go fmt ./...

# 运行测试
cd go-server && go test ./...
```

**Vue 前端**：

| 工具 | 命令 | 用途 |
|------|------|------|
| TypeScript | `npm run typecheck` | 类型检查 |
| ESLint | `npm run lint` | 代码检查 |
| Vite | `npm run build` | 生产构建 |
| Vitest | `npm run test` | 单元测试 |
| Playwright | `npm run test:e2e` | E2E 测试 |

```bash
cd web

# 类型检查
npm run typecheck

# 运行测试
npm run test

# E2E 测试
npm run test:e2e
```

### 分支策略

- `main` - 生产就绪代码
- `feature/*` - 新功能开发
- `fix/*` - Bug 修复

### Pull Request 流程

1. 从 `main` 创建功能分支
2. 编写代码和测试
3. 确保所有测试通过
4. 创建 PR 并填写描述
5. 处理审查反馈
6. Squash 合并到 main

## 常见任务

### 添加新的 API 端点

**需要修改的文件**：

1. `go-server/internal/handler/` - 添加新的处理器
2. `go-server/internal/service/` - 添加业务逻辑
3. `go-server/internal/repository/` - 添加数据访问
4. `go-server/cmd/server/main.go` - 注册路由

**步骤**：

1. 在 `handler/` 目录创建新的处理器文件
2. 在 `service/` 目录创建业务逻辑
3. 在 `repository/` 目录创建数据访问
4. 在 `main.go` 中注册路由

```go
// 示例：在 main.go 中注册新路由
func setupRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        newHandler := handler.NewHandler(service)
        api.GET("/new-endpoint", newHandler.Handle)
    }
}
```

### 添加新的数据库模型

1. 在 `go-server/internal/model/` 创建模型文件
2. 创建 SQL 迁移文件在 `go-server/migrations/`
3. 更新 `go-server/internal/database/sqlite.go` 如需要

```go
// 示例：创建新模型
package model

type NewModel struct {
    ID        string    `gorm:"primaryKey;type:text"`
    Name      string    `gorm:"type:text"`
    CreatedAt time.Time
}

func (NewModel) TableName() string {
    return "new_model"
}
```

### 添加新的前端页面

1. 在 `web/src/views/` 创建 Vue 组件
2. 在 `web/src/router/index.ts` 添加路由
3. 如需要，在 `web/src/api/` 添加 API 调用

```typescript
// 示例：添加新路由
{
  path: '/new-page',
  name: 'NewPage',
  component: () => import('@/views/NewPage.vue'),
  meta: { requiresAuth: true }
}
```

### 配置文件更新

1. 在 `go-server/internal/config/config.go` 添加新配置字段
2. 在 `go-server/config/config.yaml` 添加配置值
3. 在 `.example.env`（如存在）添加示例值

## 编码规范

### Go 编码规范

**文件组织**：
- 每个文件对应一个类型或一组相关功能
- 文件名与主要类型名一致（小写加下划线）

**命名约定**：

| 类型 | 约定 | 示例 |
|------|------|------|
| 包名 | 小写，简短 | `auth`, `user_service` |
| 结构体 | PascalCase | `UserService`, `AuthHandler` |
| 函数/方法 | camelCase | `GetByID`, `CreateUser` |
| 变量 | camelCase 或全大写 | `userID`, `MaxRetries` |
| 常量 | 全大写下划线 | `MaxLoginAttempts` |

**错误处理**：

```go
// 推荐：返回具体错误
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// 避免：忽略错误
result, _ := someFunction()
```

**日志记录**：

```go
// 使用 log.Printf 进行日志记录
log.Printf("[INFO] User %s logged in successfully", userID)
log.Printf("[ERROR] Failed to send email: %v", err)
```

### Vue/TypeScript 编码规范

**文件组织**：
- 组件文件使用 PascalCase：`UserProfile.vue`
- TypeScript 文件使用 camelCase：`authService.ts`
- 类型定义文件使用 kebab-case：`user-types.ts`

**命名约定**：

| 类型 | 约定 | 示例 |
|------|------|------|
| 组件 | PascalCase | `UserProfile.vue` |
| 组合式函数 | camelCase 以 use 开头 | `useAuth.ts` |
| API 模块 | camelCase | `authApi.ts` |
| 状态存储 | camelCase | `authStore.ts` |

**TypeScript 类型**：

```typescript
// 推荐：定义明确的数据类型
interface User {
  id: string
  email: string
  username: string
  emailVerified: boolean
}

// 避免：使用 any
function processData(data: any) { ... }
```

### API 响应格式

**成功响应**：

```json
{
  "success": true,
  "data": { ... }
}
```

**错误响应**：

```json
{
  "success": false,
  "message": "错误描述"
}
```

## 测试

### Go 测试

测试文件命名：`*_test.go`

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/service/...

# 带详细输出
go test -v ./...
```

### Vue 组件测试

测试文件位置：与组件同目录或 `__tests__/` 子目录

```bash
# 运行测试
npm run test

# 监听模式
npm run test:watch

# E2E 测试
npm run test:e2e
```

### 测试编写规范

**Go 测试**：

```go
func TestUserService_GetByID(t *testing.T) {
    // 准备测试数据
    svc := NewUserService()
    
    // 执行测试
    user, err := svc.GetByID("test-id")
    
    // 断言结果
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if user == nil {
        t.Fatal("expected user, got nil")
    }
}
```

**Vue 测试**：

```typescript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import Login from '../Login.vue'

describe('Login', () => {
    it('renders login form', () => {
        const wrapper = mount(Login)
        expect(wrapper.find('form').exists()).toBe(true)
    })
})
```

## 目录结构说明

### 后端目录结构

```
go-server/
├── cmd/server/           # 应用入口
├── internal/             # 内部包
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接和迁移
│   ├── handler/         # HTTP 请求处理器
│   ├── middleware/      # HTTP 中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   └── service/        # 业务逻辑
├── pkg/                 # 公共包
│   ├── email/           # 邮件发送
│   └── utils/           # 工具函数
├── migrations/          # SQL 迁移文件
└── data/                # 数据库文件（开发时）
```

### 前端目录结构

```
web/
├── src/
│   ├── api/            # API 调用模块
│   ├── components/     # 公共组件
│   ├── router/         # 路由配置
│   ├── stores/         # Pinia 状态存储
│   ├── types/          # TypeScript 类型定义
│   ├── views/          # 页面组件
│   │   ├── admin/      # 管理页面
│   │   └── user/      # 用户页面
│   ├── App.vue         # 根组件
│   └── main.ts         # 应用入口
├── public/             # 静态资源
└── package.json
```

## 调试

### 后端调试

使用 Delve 调试器：

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug cmd/server/main.go

# 或附加到运行中的进程
dlv attach <pid>
```

### 前端调试

使用 Vue DevTools 浏览器扩展，或在 IDE 中设置断点调试。

### 日志

后端日志输出到 stdout，生产环境应配置日志收集。

```bash
# 查看实时日志
tail -f /var/log/authnas.log
```
