# Getting Started | 快速开始

## 部署 | Deployment

AuthNas 仅支持 Docker 部署。创建 `compose.yaml` 文件：

### PostgreSQL 部署（推荐）

```yaml
services:
  authnas:
    image: authnas/authnas:latest
    restart: unless-stopped
    volumes:
      - ./authnas/config:/app/config
    environment:
      APP_URL: https://auth.example.com
      STORAGE_KEY: your-32-char-random-key
      DB_PASSWORD: your-db-password
      DB_HOST: authnas-db
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

### SQLite 部署（轻量级）

```yaml
services:
  authnas:
    image: authnas/authnas:latest
    restart: unless-stopped
    volumes:
      - ./authnas/config:/app/config
      - ./authnas/db:/app/db
    environment:
      APP_URL: https://auth.example.com
      STORAGE_KEY: your-32-char-random-key
      DB_ADAPTER: sqlite
```

## 启动 | Startup

```bash
docker compose up -d
```

## 设置管理员账户

首次启动后，在日志中查找管理员密码重置链接：

```bash
docker compose logs authnas
```

访问密码重置链接，设置管理员密码后登录。

> [!IMPORTANT]
> `auth_admins` 组的用户将成为管理员。首次登录后建议创建新用户并将自己加入该组。

## 配置 | Configuration

### 必需环境变量

| 变量 | 说明 |
|------|------|
| APP_URL | AuthNas 访问地址，如 `https://auth.example.com` |
| STORAGE_KEY | 加密密钥，至少 32 字符，随机生成 |
| DB_PASSWORD | 数据库密码（PostgreSQL） |
| DB_HOST | 数据库主机（PostgreSQL），如 `authnas-db` |

### 可选环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| APP_TITLE | AuthNas | 网站标题 |
| APP_COLOR | #906bc7 | 主题色 |
| SIGNUP | false | 是否允许用户自注册 |
| EMAIL_VERIFICATION | false | 是否验证邮箱 |
| MFA_REQUIRED | false | 是否强制 MFA |
| SMTP_HOST | - | SMTP 服务器地址 |

### 自定义 Branding

挂载 `/app/config` 目录后可自定义：

```
/app/config/branding/logo.svg      # Logo
/app/config/branding/favicon.svg   # Favicon
```

## 下一步

- [OIDC 应用配置](OIDC-Setup.md) - 将 AuthNas 集成到支持 OIDC 的应用
- [ProxyAuth 配置](ProxyAuth-and-Trusted-Header-SSO-Setup.md) - 保护不支持 OIDC 的应用
- [用户管理](User-Management.md) - 管理用户和邀请
