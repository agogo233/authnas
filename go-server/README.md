# AuthNas Go Server

AuthNas Go 后端服务器，一款开源的 SSO 认证和用户管理服务。

This is the Go backend server for AuthNas, an open-source SSO authentication and user management provider.

---

## 功能特点 | Features

| 中文 | English |
|------|---------|
| OpenID Connect (OIDC) 提供商 | OpenID Connect (OIDC) Provider |
| 代理 ForwardAuth | Proxy ForwardAuth |
| 用户和用户组管理 | User and Groups Management |
| 用户自注册和邀请 | User Self-Registration and Invitations |
| 多因素认证、Passkeys、无密码账户 | Multi-factor Authentication, Passkeys, and Passkey-Only Accounts |
| 安全的密码重置（邮箱验证） | Secure Password Reset with Email Verification |
| 数据加密存储 | Encryption-At-Rest with Postgres or SQLite Database |

---

## 技术栈 | Tech Stack

| 技术 | Technology | 中文说明 |
|------|------------|----------|
| **Language**: Go 1.25+ | Go 1.25+ | 高性能、编译型语言 |
| **Web Framework**: Gin | Gin | 轻量级 Web 框架 |
| **Database**: GORM | GORM | Go 数据库 ORM |
| **Authentication**: WebAuthn, JWT, TOTP | WebAuthn, JWT, TOTP | 认证协议 |

---

## 项目结构 | Project Structure

```
go-server/
├── cmd/server/          # 应用入口 | Main application entry point
├── internal/            # 内部包（不可外部导入）| Internal packages
│   ├── config/          # 配置管理 | Configuration management
│   ├── crypto/          # 密码学工具 | Cryptographic utilities
│   ├── database/        # 数据库连接和迁移 | Database connection and migrations
│   ├── handler/         # HTTP 请求处理器 | HTTP request handlers
│   ├── middleware/       # HTTP 中间件 | HTTP middleware
│   ├── model/           # 数据模型 | Data models
│   ├── oidc/            # OIDC 实现 | OIDC implementation
│   ├── repository/       # 数据访问层 | Data access layer
│   └── service/         # 业务逻辑层 | Business logic
├── pkg/                 # 公共包（可外部导入）| Public packages
│   ├── database/        # 数据库工具 | Database utilities
│   ├── email/           # 邮件发送 | Email sending
│   └── utils/           # 通用工具 | General utilities
├── migrations/          # SQL 迁移文件 | SQL migrations
├── keys/                # JWT 密钥（自动生成）| JWT keys (generated)
└── config/              # 配置文件 | Configuration files
```

---

## 快速开始 | Getting Started

### 前置条件 | Prerequisites

| 条件 | Requirement |
|------|-------------|
| Go >= 1.25 | Go >= 1.25 |
| Node.js >= 18 (仅开发时需要) | Node.js >= 18 (only for development) |
| SQLite 或 PostgreSQL 数据库 | SQLite or PostgreSQL database |

### 一体化架构 | All-in-One Architecture

AuthNas 采用前后端一体化部署架构，Go 服务器同时提供 API 服务和前端页面：

- **API 服务**：`/api/*`, `/oidc/*` 等路由
- **前端页面**：所有其他路由返回 Vue SPA
- **静态文件**：由 Go 程序直接 serve，无需单独运行前端服务器

### Vite 的角色 | Vite's Role

**重要说明**：Vite 只是作为 TypeScript/Vue 的编译工具，将前端代码编译为静态资源文件。

```
web/src/ (Vue/TypeScript 源码)
    ↓ npm run build (Vite 编译)
go-server/static/ (编译后的 HTML/CSS/JS)
    ↓ Go 程序读取
http://localhost:8080/ (统一入口)
```

**测试的真实对象**：
- **真正的测试目标是 Go 服务器** (`localhost:8080`)
- Go server 配置了反向代理，将前端静态资源挂载到根路径
- E2E 测试直接访问 `http://localhost:8080`，由 Go 程序提供所有页面和 API
- 不需要也**不应该**单独运行前端开发服务器 (`localhost:3000`) 进行测试

### 前端开发 vs 测试 | Frontend Development vs Testing

| 场景 | Vite 作用 | 测试入口 |
|------|----------|----------|
| 本地开发 | 热重载开发服务器 | 不需要，Go server 已包含静态资源 |
| 生产构建 | 编译静态资源到 `static/` | 直接测试 Go server |
| E2E 测试 | 只用于构建 `static/` | 测试 Go server (`localhost:8080`) |

### 安装步骤 | Installation

#### 中文

1. **下载 Go 依赖**：
   ```bash
   go mod download
   ```

2. **配置环境变量**：
   ```bash
   cp .example.env config/.env
   # 编辑 config/.env 配置
   ```

3. **生成 JWT 密钥**：
   ```bash
   make keys
   ```

4. **构建前端**（在项目根目录）：
   ```bash
   cd web && npm install && npm run build && cd ..
   ```

5. **运行服务器**：
   ```bash
   go run ./cmd/server
   ```

   服务器启动后访问 `http://localhost:8080` 即可使用。

#### English

1. **Download Go dependencies**:
   ```bash
   go mod download
   ```

2. **Configure environment variables**:
   ```bash
   cp .example.env config/.env
   # Edit config/.env with your settings
   ```

3. **Generate JWT keys**:
   ```bash
   make keys
   ```

4. **Build frontend** (from project root):
   ```bash
   cd web && npm install && npm run build && cd ..
   ```

5. **Run the server**:
   ```bash
   go run ./cmd/server
   ```

   After server starts, visit `http://localhost:8080` to use the application.

---

### 使用 Makefile

| 命令 | Command | 说明 | Description |
|------|---------|------|-------------|
| `make build` | make build | 构建二进制文件 | Build the binary |
| `make run` | make run | 构建并运行 | Build and run |
| `make test` | make test | 运行测试 | Run tests |
| `make clean` | make clean | 清理构建文件 | Clean build files |
| `make deps` | make deps | 下载依赖 | Download dependencies |
| `make keys` | make keys | 生成 JWT 密钥 | Generate JWT keys |

---

## 测试 | Testing

运行所有测试 | Run all tests：

```bash
go test -v ./...
```

---

## Docker 部署 | Docker

使用 Docker 构建并运行 | Build and run with Docker：

```bash
docker build -t authnas .
docker run -p 8080:8080 authnas
```

或使用 Docker Compose（或从项目根目录）| Or use Docker Compose (from project root)：

```bash
docker compose up
```
