# Troubleshooting | 常见问题

## 管理员重置链接失效

链接有效期为一天。解决方法：

1. 使用 CLI 生成新链接：
   ```bash
   docker compose run authnas generate password-reset auth_admin
   ```
2. 或删除数据库重新初始化

## Could Not Create Session

检查：
- `APP_URL` 环境变量是否正确设置
- `SESSION_DOMAIN` 是否为有效域名（不能是顶级域名如 `.com`）

## Invalid Client

确保 OIDC 应用的 Client ID 与 AuthNas 中配置完全一致。

## Invalid Redirect Uri

检查回调地址是否与应用文档中一致，可参考 [OIDC 集成指南](OIDC-Guides.md)。

## 页面无法找到（OIDC 认证时）

检查 AuthNas OIDC 端点 URL 配置是否正确，在管理面板 OIDC 应用页面顶部可获取正确地址。

## 登录后未重定向（ProxyAuth）

检查反向代理是否正确设置了 `X-Forwarded-*` 头，详见 [ProxyAuth 配置](ProxyAuth-and-Trusted-Header-SSO-Setup.md)。

## 日志问题

**IP 地址错误或缺失**
检查反向代理的 trusted IP 配置。
