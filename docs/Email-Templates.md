# Email Templates | 邮件模板

## 模板位置

邮件模板位于 `/app/config/email_templates` 目录，每个邮件类型包含：

```
subject.default.ejs   # 邮件主题
html.default.ejs      # HTML 格式
text.default.ejs      # 纯文本格式
```

> [!WARNING]
> 修改模板后必须移除 `.default` 后缀，否则每次启动会被覆盖。

## 可用变量

| 变量 | 说明 |
|------|------|
| `{{.DisplayName}}` | 显示名称 |
| `{{.Username}}` | 用户名 |
| `{{.Email}}` | 邮箱地址 |
| `{{.Link}}` | 邀请/重置链接 |
| `{{.AppName}}` | 应用名称 |

## 邮件类型

- `invitation/` - 邀请邮件
- `password-reset/` - 密码重置邮件
- `email-verification/` - 邮箱验证邮件
