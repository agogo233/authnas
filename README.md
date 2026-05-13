# AuthNas

[![GitHub Actions](https://img.shields.io/github/actions/workflow/status/authnas/authnas/release.yml)](https://github.com/authnas/authnas/actions)
[![GitHub Release](https://img.shields.io/github/v/release/authnas/authnas?logo=github)](https://github.com/authnas/authnas/releases)
[![License](https://img.shields.io/github/license/authnas/authnas)](LICENSE)

<p align="center">
  <strong>自托管应用的单点登录解决方案</strong>
</p>

<p align="center">
  Single Sign-On for Your Self-Hosted Universe
</p>

<p align="center">
  <a href="https://authnas.app">官网 | Website</a> ·
  <a href="https://github.com/authnas/authnas">源码 | Source</a> ·
  <a href="https://github.com/authnas/authnas/issues">问题 | Issues</a>
</p>

---

## 功能亮点 | Features

| 功能 | Feature |
|------|---------|
| OpenID Connect (OIDC) 提供商 | OpenID Connect (OIDC) Provider |
| Auth ForwardAuth 代理 | Auth ForwardAuth Proxy |
| 用户与用户组管理 | User and Groups Management |
| 用户自注册与邀请码 | Self-Registration & Invitations |
| Passkeys（通行密钥）支持 | Passkeys Support |
| 多因素认证 (MFA) | Multi-factor Authentication |
| 安全的邮件密码重置 | Secure Password Reset via Email |
| 自定义 Logo、标题、主题色 | Custom Logo, Title & Theme Color |
| 数据库透明加密存储 | Encryption-At-Rest (PostgreSQL/SQLite) |

---

## 技术栈 | Tech Stack

| 组件 | 技术 |
|------|------|
| 后端 | Go + Gin |
| 前端 | Vue 3 + TypeScript + Naive UI |
| 数据库 | PostgreSQL / SQLite |

## 重要说明：Vite 只是编译工具

**Vite 的作用是将前端代码编译为静态资源文件，不参与实际服务或测试**。

```
web/src/ (Vue/TypeScript 源码)
    ↓ Vite 编译 (npm run build)
go-server/static/ (编译后的 HTML/CSS/JS)
    ↓ Go 程序读取并服务
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

---

## 快速开始 | Quick Start

### Docker Compose 部署

```yaml
services:
  authnas:
    image: authnas/authnas:latest
    restart: unless-stopped
    volumes:
      - ./authnas/config:/app/config
    environment:
      APP_URL: https://auth.example.com      # 必填
      STORAGE_KEY: your-storage-key          # 必填
      DB_PASSWORD: your-db-password         # 必填
      DB_HOST: authnas-db                    # 必填
    depends_on:
      authnas-db:
        condition: service_healthy

  authnas-db:
    image: postgres:18
    restart: unless-stopped
    environment:
      POSTGRES_PASSWORD: your-db-password
    volumes:
      - db:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]

volumes:
  db:
```

启动后访问 `APP_URL`，在日志中找到初始管理员密码重置链接：

```bash
docker compose logs authnas
```

详细配置请参阅 [官方文档](https://authnas.app)。

---

## 项目结构 | Structure

```
authnas/
├── go-server/          # Go 后端服务 (SSO 核心)
├── web/                # Vue 3 前端应用
├── docs/               # 英文文档
└── .monkeycode/docs/   # 中文开发者文档
```

---

## 支持与贡献 | Support

- 问题反馈：[GitHub Issues](https://github.com/authnas/authnas/issues)
- 讨论与帮助：[GitHub Discussions](https://github.com/orgs/authnas/discussions)
- 贡献指南：[CONTRIBUTING.md](CONTRIBUTING.md)

---

## 免责声明 | Disclaimer

本项目尚未通过安全审计，使用风险自负。

AuthNas has not been audited and uses 3rd party packages for much of its functionality, use at your own risk.
