# OIDC Setup | OIDC 配置

## 创建 OIDC 应用

在 AuthNas 管理面板的 OIDC 页面创建新应用。

### 基础配置

| 字段 | 说明 | 示例 |
|------|------|------|
| Client ID | 应用唯一标识 | `my-app` |
| Client Secret | 密钥，妥善保管 | 随机生成 |
| Auth Method | 认证方式 | `client_secret_post` |
| Redirect URLs | 回调地址 | `https://app.example.com/callback` |

### 可选配置

- **Display Name** - 显示名称
- **Logo URL** - 应用 Logo
- **Groups** - 允许访问的用户组
- **Skip Consent** - 跳过用户授权确认
- **MFA Required** - 是否要求 MFA

### 获取 OIDC 端点信息

在 OIDC 应用页面顶部下拉面板可获取所需的端点信息：

```
OIDC Issuer:       https://auth.example.com/oidc
Auth Endpoint:     https://auth.example.com/oidc/auth
Token Endpoint:    https://auth.example.com/oidc/token
UserInfo Endpoint: https://auth.example.com/oidc/me
JWKs Endpoint:     https://auth.example.com/oidc/jwks
```

> [!NOTE]
> Redirect URLs 支持通配符，但需谨慎使用。

## 应用集成示例

详见 [OIDC 集成指南](OIDC-Guides.md)，包含以下应用配置示例：

Actual Budget, Arcane, AutoCaliWeb, Beszel, ByteStash, Cloudflare ZeroTrust, Dawarich, Dockhand, Grist, Immich, Jellyfin, Jellyseerr, Komodo, Manyfold, Mastodon, Memos, Open WebUI, Pangolin, Paperless-ngx, Portainer, Proxmox PVE, Seafile, Unraid, Vaultwarden, WikiJS 等。
