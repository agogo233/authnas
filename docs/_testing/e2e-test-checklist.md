# E2E 测试场景清单

## 重要说明：测试目标

**AuthNas 采用前后端一体化架构**：
- **Go 服务器** (`localhost:8080`) 是测试的真正目标
- Go server 内置了前端静态资源的代理和 serve 功能
- **Vite 只是编译工具**：将 Vue/TypeScript 源码编译为静态资源
- E2E 测试直接访问 `http://localhost:8080`，无需运行前端开发服务器

```
web/src/          (Vue/TypeScript 源码)
       ↓ Vite 编译 (npm run build)
go-server/static/ (静态资源: HTML/CSS/JS)
       ↓ Go Server 读取并服务
测试 http://localhost:8080 (所有 API + 页面)
```

### 前端开发 vs 测试

| 场景 | Vite 作用 | 测试入口 |
|------|----------|----------|
| 本地开发 | 热重载开发服务器 | 不需要，Go server 已包含静态资源 |
| 生产构建 | 编译静态资源到 `static/` | 直接测试 Go server |
| E2E 测试 | 只用于构建 `static/` | 测试 Go server (`localhost:8080`) |

---

## 测试文件列表

| 文件名 | 描述 | 测试数量 |
|--------|------|----------|
| `login.spec.ts` | 登录流程测试 | 8 |
| `register.spec.ts` | 注册流程测试 | 7 |
| `mfa.spec.ts` | MFA 验证测试 | 9 |
| `passkey.spec.ts` | Passkey 测试 | - |
| `passkeys-page.spec.ts` | Passkeys 页面测试 | - |
| `security.spec.ts` | 安全设置测试 | 12 |
| `consent.spec.ts` | 授权同意测试 | 11 |
| `verify-email.spec.ts` | 邮箱验证测试 | - |
| `reset-password.spec.ts` | 密码重置测试 | - |
| `profile.spec.ts` | 用户资料测试 | - |
| `admin-dashboard.spec.ts` | 管理仪表盘测试 | - |
| `admin-pages.spec.ts` | 管理页面测试 | - |
| `admin-users.spec.ts` | 用户管理测试 | 7 |
| `navigation.spec.ts` | 导航测试 | - |
| `auth-redirect.spec.ts` | 认证重定向测试 | 10 |
| `error-page.spec.ts` | 错误页面测试 | - |
| `security-injection.spec.ts` | 安全注入测试 | **35+** |
| `oidc-flow.spec.ts` | OIDC 完整流程测试 | **25+** |
| `boundary-conditions.spec.ts` | 边界条件和错误处理测试 | **40+** |

---

## 核心测试场景分类

### 1. 认证流程 (Authentication) - P0

| 场景 | 登录 | 注册 | MFA | 密码重置 |
|------|------|------|-----|----------|
| 正常流程 | ✅ | ✅ | ✅ | ✅ |
| 空字段验证 | ✅ | ✅ | ✅ | ✅ |
| 错误凭证 | ✅ | - | ✅ | - |
| 弱密码检测 | - | ✅ | - | ✅ |
| 暴力破解防护 | ✅ | - | - | - |
| 会话管理 | ✅ | - | - | - |
| Token 验证 | ✅ | - | ✅ | ✅ |

### 2. 授权与访问控制 (Authorization) - P0

| 场景 | 描述 | 优先级 |
|------|------|--------|
| 未登录访问受保护路由 | 应重定向到登录页 | P0 |
| 管理员路由保护 | 普通用户不能访问 | P0 |
| 权限提升防护 | 不能通过 URL 操纵访问更高权限 | P0 |
| IDOR 防护 | 不能访问其他用户资源 | P0 |
| Consent 页面保护 | 未登录不能访问 | P0 |

### 3. 安全测试 (Security) - P0

#### 3.1 注入攻击防护

| 攻击类型 | 测试场景数 | 覆盖范围 |
|----------|------------|----------|
| XSS | 10+ | 用户名、邮箱、Profile、搜索框 |
| SQL 注入 | 8+ | 登录、注册、搜索、邀请码 |
| 命令注入 | 6+ | 用户名字段 |
| LDAP 注入 | 待补充 | - |

#### 3.2 认证安全

| 测试项 | 描述 |
|--------|------|
| 暴力破解防护 | 连续失败后的锁定/警告 |
| 密码策略 | 最小长度、强度、常见弱密码 |
| 会话安全 | Token 过期、格式验证、并发会话 |
| MFA 安全 | 一次性使用、长度验证、格式验证 |

#### 3.3 数据保护

| 测试项 | 描述 |
|--------|------|
| 敏感信息泄露 | 密码不在 UI/API 响应中 |
| 错误消息 | 不暴露系统内部信息 |
| 安全头部 | X-Frame-Options, CSP, HSTS 等 |
| Token 安全 | 不在 URL 中暴露 |

### 4. OIDC 流程测试 - P1

| 阶段 | 测试项 |
|------|--------|
| Authorization Request | 参数验证、redirect_uri 验证 |
| Consent | 客户端信息、Scopes 显示、授权/拒绝 |
| Token Exchange | Authorization Code 交换、PKCE |
| ID Token | 签名验证、Claims 验证、Nonce |
| UserInfo | Access Token 验证 |
| Refresh Token | Token 刷新、过期处理 |
| Discovery | OpenID Connect Discovery 端点 |
| Security | Code Replay 防护、Scope 限制 |

### 5. 管理后台测试 - P1

| 模块 | CRUD | 搜索/过滤 | 权限控制 |
|------|------|------------|----------|
| 用户管理 | ✅ | ✅ | ✅ |
| 群组管理 | ✅ | - | ✅ |
| 客户端管理 | ✅ | - | ✅ |
| 邀请管理 | ✅ | - | ✅ |
| 系统设置 | ✅ | - | ✅ |

### 6. 边界条件和错误处理 - P2

| 类别 | 测试场景数 |
|------|------------|
| 输入长度边界 | 5+ |
| 特殊字符处理 | 10+ |
| 空值/Null 处理 | 5+ |
| 并发操作 | 4+ |
| 网络错误 | 5+ |
| 状态管理 | 4+ |
| 表单验证边界 | 4+ |
| 导航边界 | 5+ |
| 超时处理 | 2+ |
| 数据一致性 | 2+ |

### 7. 辅助功能 (Accessibility) - P3

| 测试项 | 描述 |
|--------|------|
| 键盘导航 | Tab 键焦点顺序 |
| ARIA 属性 | 表单标签、错误关联 |
| 语义化 HTML | 按钮、表单元素 |
| 屏幕阅读器兼容性 | ARIA live regions |

### 8. 性能测试 - P3

| 指标 | 目标 |
|------|------|
| 登录页面加载 | < 2s |
| 登录响应时间 | < 1s |
| 管理页面加载 | < 3s |
| 大列表分页 | < 3s |

---

## 测试执行矩阵

| 测试类型 | 本地开发 | CI/CD PR | 夜间构建 |
|----------|----------|----------|----------|
| P0 功能测试 | ✅ | ✅ | ✅ |
| P0 安全测试 | ✅ | ✅ | ✅ |
| P1 功能测试 | ✅ | ✅ | ✅ |
| P1 OIDC 测试 | ✅ | ✅ | ✅ |
| P2 边界测试 | ✅ | - | ✅ |
| P3 性能和 A11Y | - | - | ✅ |

---

## 安全测试清单 (详细)

### 注入防护 ✅

- [x] XSS 防护测试 (8+ 场景)
  - [x] 用户名输入框
  - [x] 邮箱输入框
  - [x] Profile 页面
  - [x] MFA 代码输入框
  - [x] 搜索功能

- [x] SQL 注入防护测试 (8+ 场景)
  - [x] 登录用户名
  - [x] 注册邮箱
  - [x] 管理搜索
  - [x] 邀请码

- [x] 命令注入防护测试 (6+ 场景)
  - [x] 用户名输入框

### 认证安全 ✅

- [x] 暴力破解防护
  - [x] 多次失败后显示警告
  - [x] 账户锁定机制
  - [x] 统一错误信息

- [x] 密码策略
  - [x] 最小长度验证
  - [x] 弱密码检测
  - [x] 用户名包含检测
  - [x] 强密码接受

- [x] 会话安全
  - [x] 过期 Token 处理
  - [x] 畸形 Token 处理
  - [x] 登出后状态清理
  - [x] 并发会话管理
  - [x] 撤销所有会话

### 授权安全 ✅

- [x] 特权升级防护
  - [x] 普通用户访问管理员页面
  - [x] 用户 ID 操纵

- [x] IDOR 防护
  - [x] Consent UID 操纵
  - [x] 受保护路由验证

- [x] 重定向安全
  - [x] 外部 URL 重定向
  - [x] javascript: 协议
  - [x] 相对路径重定向

### 数据安全 ✅

- [x] 敏感信息保护
  - [x] 密码不明文显示
  - [x] 密码字段掩码
  - [x] Token 不在 URL 中

- [x] 错误信息泄露
  - [x] 登录失败统一消息
  - [x] 注册失败信息
  - [x] 无堆栈跟踪

- [x] 安全头部
  - [x] X-Frame-Options
  - [x] X-Content-Type-Options
  - [x] Cache-Control

### MFA 安全 ✅

- [x] 代码重用防护
- [x] 长度验证
- [x] 非数字字符拒绝

---

## 运行测试

### 运行所有 E2E 测试
```bash
cd web && npm run test:e2e
```

### 运行特定测试文件
```bash
cd web && npx playwright test e2e/security-injection.spec.ts
```

### 运行带 UI 的测试
```bash
cd web && npm run test:e2e:ui
```

### 运行特定标签的测试
```bash
npx playwright test --grep "XSS"
npx playwright test --grep "SQL Injection"
```

### 在 CI 模式运行
```bash
CI=true npm run test:e2e
```

---

## 测试覆盖率

| 类别 | 覆盖率 |
|------|--------|
| 认证流程 | ~95% |
| 授权流程 | ~90% |
| 安全注入 | ~90% |
| OIDC 流程 | ~85% |
| 管理后台 | ~80% |
| 边界条件 | ~75% |
| 错误处理 | ~80% |

---

## 待补充测试场景

1. **P3 - 性能测试**: 使用 Playwright 性能指标
2. **P3 - 辅助功能**: 更详细的 ARIA 测试
3. **P3 - 视觉回归测试**: 截图对比
4. **P2 - WebSocket 测试**: 实时通知
5. **P1 - 国际化测试**: 多语言支持
6. **P2 - 邮件功能测试**: 邮件模板和发送
