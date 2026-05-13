# AuthNas 文档

AuthNas 是一个开源的 SSO（单点登录）认证和用户管理平台，为自托管应用提供安全的身份认证服务。

## 项目概述

**目标用户**: 需要为自托管应用提供统一认证解决方案的开发者和管理员

**核心价值**:
- 提供标准化的 OIDC（OpenID Connect）认证服务
- 支持多种认证方式（密码、Passkeys、TOTP）
- 简洁的管理界面和用户自助服务
- 支持 SQLite 和 PostgreSQL 数据库

## 技术栈

| 类别 | 选择 | 说明 |
|------|------|------|
| 后端语言 | Go 1.25+ | 高性能、并发支持 |
| 后端框架 | Gin | 轻量级 Web 框架 |
| 前端框架 | Vue 3 + TypeScript | 现代化前端框架 |
| UI 组件库 | Naive UI | Vue 3 组件库 |
| 状态管理 | Pinia | Vue 3 状态管理 |
| 数据库 | SQLite / PostgreSQL | 灵活的数据存储 |
| ORM | GORM | Go 数据库 ORM |
| 认证协议 | OIDC | 标准身份认证协议 |

## 核心文档

| 文档 | 描述 |
|------|------|
| [架构设计](./ARCHITECTURE.md) | 系统架构、技术栈、组件结构 |
| [接口文档](./INTERFACES.md) | API 接口、认证方式、事件定义 |
| [开发者指南](./DEVELOPER_GUIDE.md) | 环境搭建、开发规范、常见任务 |

## 快速链接

- [架构](./ARCHITECTURE.md) | [接口](./INTERFACES.md) | [开发者指南](./DEVELOPER_GUIDE.md)

## 项目结构

```
workspace/
├── go-server/                         # Go 后端服务
│   ├── cmd/server/                    # 应用入口
│   │   └── main.go                    # 主函数
│   ├── internal/                     # 内部包（不可外部导入）
│   │   ├── config/                   # 配置管理
│   │   ├── crypto/                   # 密码学工具
│   │   ├── database/                 # 数据库连接和迁移
│   │   ├── handler/                  # HTTP 处理器
│   │   │   ├── admin.go              # 管理后台处理器
│   │   │   ├── auth.go               # 认证处理器
│   │   │   ├── health.go             # 健康检查
│   │   │   ├── oidc.go               # OIDC 协议处理器
│   │   │   ├── passkey.go            # Passkey 处理器
│   │   │   ├── totp.go               # TOTP 处理器
│   │   │   └── user.go               # 用户管理处理器
│   │   ├── middleware/               # HTTP 中间件
│   │   ├── model/                    # 数据模型
│   │   ├── oidc/                     # OIDC 实现
│   │   ├── repository/               # 数据访问层
│   │   └── service/                  # 业务逻辑层
│   │       ├── auth_service.go       # 认证服务
│   │       ├── client_service.go     # 客户端服务
│   │       ├── consent_service.go    # 授权同意服务
│   │       ├── email_service.go      # 邮件服务
│   │       ├── group_service.go      # 用户组服务
│   │       ├── invitation_service.go # 邀请服务
│   │       ├── oidc_service.go       # OIDC 服务
│   │       ├── passkey_service.go    # Passkey 服务
│   │       ├── proxy_auth_service.go # Proxy Auth 服务
│   │       ├── totp_service.go       # TOTP 服务
│   │       └── user_service.go       # 用户服务
│   ├── pkg/                          # 公共包（可外部导入）
│   │   ├── database/                 # 数据库工具
│   │   ├── email/                   # 邮件发送
│   │   └── utils/                   # 通用工具
│   ├── migrations/                   # SQL 迁移文件
│   ├── keys/                         # JWT 密钥（自动生成）
│   └── config/                       # 配置文件
├── web/                              # Vue.js 前端
│   ├── src/
│   │   ├── api/                     # API 调用
│   │   ├── components/              # 组件
│   │   ├── stores/                  # 状态管理
│   │   ├── views/                   # 页面视图
│   │   └── router/                  # 路由配置
│   └── public/                      # 静态资源
```

## 入门指南

### 新加入项目？

按此路径学习：
1. **[架构设计](./ARCHITECTURE.md)** - 了解系统全局
2. **[开发者指南](./DEVELOPER_GUIDE.md)** - 搭建开发环境

### 需要集成？

1. **[接口文档](./INTERFACES.md)** - API 契约和认证方式
2. **[架构设计](./ARCHITECTURE.md)** - 系统边界和数据流

## 常用命令

```bash
# 后端运行
cd go-server && go run ./cmd/server

# 前端运行
cd web && npm run dev

# 运行测试
cd go-server && go test ./...
cd web && npm run test
```
