# AuthNas E2E 测试场景文档

## 概述

本文档定义 AuthNas SSO 认证系统的完整端到端(E2E)测试场景。测试覆盖认证、用户管理、OIDC协议、管理后台、安全性等所有核心功能模块。

---

## 重要说明：测试目标

**AuthNas 采用前后端一体化架构**：
- **Go 服务器** (`localhost:8080`) 是测试的真正目标
- Go server 内置了前端静态资源的代理和 serve 功能
- **Vite 只是编译工具**：将 Vue/TypeScript 源码编译为静态资源
- E2E 测试直接访问 `http://localhost:8080`，无需运行前端开发服务器

```
web/src/          (Vue/TypeScript 源码)
       ↓ Vite 编译
go-server/static/ (静态资源: HTML/CSS/JS)
       ↓ Go Server 读取并服务
测试 http://localhost:8080 (所有 API + 页面)
```

---

## 目录

1. [认证模块测试场景](#1-认证模块测试场景)
2. [用户管理模块测试场景](#2-用户管理模块测试场景)
3. [OIDC协议测试场景](#3-oidc协议测试场景)
4. [管理后台测试场景](#4-管理后台测试场景)
5. [安全性和授权测试场景](#5-安全性和授权测试场景)
6. [会话管理测试场景](#6-会话管理测试场景)
7. [TOTP和MFA测试场景](#7-totp和mfa测试场景)
8. [Passkey测试场景](#8-passkey测试场景)
9. [邮件验证测试场景](#9-邮件验证测试场景)
10. [邀请注册测试场景](#10-邀请注册测试场景)
11. [密码安全测试场景](#11-密码安全测试场景)
12. [OAuth Client管理测试场景](#12-oauth-client管理测试场景)
13. [Proxy Auth测试场景](#13-proxy-auth测试场景)
14. [边界条件和错误处理测试场景](#14-边界条件和错误处理测试场景)
15. [性能和并发测试场景](#15-性能和并发测试场景)

---

## 1. 认证模块测试场景

### 1.1 用户注册测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| AUTH_REG_001 | 正常用户注册 | 无 | 提交有效的username/email/password | 返回200，access_token和refresh_token | P0 |
| AUTH_REG_002 | 注册缺少用户名 | 无 | 提交email和password，username为空 | 返回400 | P0 |
| AUTH_REG_003 | 注册缺少邮箱 | 无 | 提交username和password，email为空 | 返回400或200(允许) | P1 |
| AUTH_REG_004 | 注册缺少密码 | 无 | 提交username和email，password为空 | 返回400 | P0 |
| AUTH_REG_005 | 注册无效邮箱格式 | 无 | username有效，email为"not-an-email" | 返回400 | P0 |
| AUTH_REG_006 | 注册弱密码 | 无 | password为"123" | 返回400或200(如果密码强度检查关闭) | P1 |
| AUTH_REG_007 | 注册重复用户名 | 无 | 先注册user1，再用相同用户名注册 | 返回400 | P0 |
| AUTH_REG_008 | 注册重复邮箱 | 无 | 先注册user1，再用相同邮箱注册 | 返回400 | P0 |
| AUTH_REG_009 | 注册用户名包含特殊字符 | 无 | username为"user@#$%" | 返回400 | P1 |
| AUTH_REG_010 | 注册用户名包含空格 | 无 | username为"user with spaces" | 返回400 | P1 |
| AUTH_REG_011 | 注册SQL注入攻击 | 无 | username为"user'; DROP TABLE users;--" | 返回400，无数据泄露 | P0 |
| AUTH_REG_012 | 注册XSS攻击 | 无 | username为"<script>alert('xss')</script>" | 返回400 | P0 |
| AUTH_REG_013 | 注册Unicode用户名 | 无 | username为"用户"或其他Unicode | 返回400或200(取决于配置) | P2 |
| AUTH_REG_014 | 注册超长用户名 | 无 | username超过100字符 | 返回400 | P1 |
| AUTH_REG_015 | 注册超长邮箱 | 无 | email超过255字符 | 返回400 | P1 |
| AUTH_REG_016 | 注册超长密码 | 无 | password超过1000字符 | 返回400或200 | P2 |
| AUTH_REG_017 | 注册空body | 无 | 发送空JSON body | 返回400 | P0 |
| AUTH_REG_018 | 注册无效JSON | 无 | 发送"not json"作为body | 返回400 | P0 |
| AUTH_REG_019 | 注册后自动登录验证 | 无 | 注册成功后使用返回的token访问/api/user/me | 返回200，用户信息正确 | P0 |
| AUTH_REG_020 | 注册后token版本初始化 | 无 | 注册后检查token_version为0 | token_version=0 | P1 |

### 1.2 用户登录测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| AUTH_LOGIN_001 | 使用用户名登录成功 | 用户已注册 | input=username, password正确 | 返回200，access_token | P0 |
| AUTH_LOGIN_002 | 使用邮箱登录成功 | 用户已注册 | input=email, password正确 | 返回200，access_token | P0 |
| AUTH_LOGIN_003 | 登录密码错误 | 用户已注册 | password为"wrongpassword" | 返回401 | P0 |
| AUTH_LOGIN_004 | 登录用户不存在 | 无 | input为不存在的用户名 | 返回401 | P0 |
| AUTH_LOGIN_005 | 登录缺少password字段 | 无 | 只提交input | 返回400 | P0 |
| AUTH_LOGIN_006 | 登录缺少input字段 | 无 | 只提交password | 返回400 | P0 |
| AUTH_LOGIN_007 | 登录空body | 无 | 发送空JSON | 返回400 | P0 |
| AUTH_LOGIN_008 | 登录SQL注入攻击 | 用户已注册 | password为"' OR '1'='1" | 返回401 | P0 |
| AUTH_LOGIN_009 | 登录XSS攻击 | 用户已注册 | username包含XSS payload | 返回400或401 | P0 |
| AUTH_LOGIN_010 | 连续登录失败5次 | 用户已注册 | 密码错误连续5次 | 可能触发账户锁定或限流 | P1 |
| AUTH_LOGIN_011 | 登录后访问受保护资源 | 登录成功 | 用返回的token访问/api/user/me | 返回200 | P0 |
| AUTH_LOGIN_012 | 登录返回MFARequired标志 | 用户配置了MFA | 登录检查返回mfa_required字段 | mfa_required为true | P1 |

### 1.3 密码重置测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| AUTH_PWD_RESET_001 | 忘记密码-用户存在 | 用户已注册 | 提交已注册邮箱 | 返回200 | P0 |
| AUTH_PWD_RESET_002 | 忘记密码-用户不存在 | 无 | 提交不存在邮箱 | 返回200(防用户枚举) | P0 |
| AUTH_PWD_RESET_003 | 重置密码-无效code | 无 | 提交无效的reset code和新密码 | 返回400 | P0 |
| AUTH_PWD_RESET_004 | 重置密码-empty code | 无 | 提交空code和新密码 | 返回400 | P0 |
| AUTH_PWD_RESET_005 | 重置密码-empty新密码 | 无 | 提交有效code，new_password为空 | 返回400 | P0 |
| AUTH_PWD_RESET_006 | 重置密码-新密码弱 | 无 | 提交有效code，新密码为"123" | 返回400 | P1 |
| AUTH_PWD_RESET_007 | 重置密码-成功 | 已有重置链接 | 使用有效code重置密码成功 | 密码被更新，可登录 | P0 |
| AUTH_PWD_RESET_008 | 重置密码-使用后失效 | 密码已重置 | 再次使用同一code重置 | 返回400，code已失效 | P1 |

### 1.4 密码修改测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| AUTH_PWD_UPDATE_001 | 修改密码-正确旧密码 | 用户已登录 | old_password正确，new_password有效 | 返回200 | P0 |
| AUTH_PWD_UPDATE_002 | 修改密码-错误旧密码 | 用户已登录 | old_password错误 | 返回400 | P0 |
| AUTH_PWD_UPDATE_003 | 修改密码-未认证 | 无token | 访问修改密码接口 | 返回401 | P0 |
| AUTH_PWD_UPDATE_004 | 修改密码-新密码与旧密码相同 | 用户已登录 | old和new相同 | 返回200或400 | P2 |
| AUTH_PWD_UPDATE_005 | 修改密码-新密码弱 | 用户已登录 | old正确，new为"123" | 返回400 | P1 |
| AUTH_PWD_UPDATE_006 | 修改密码-缺少old_password | 用户已登录 | 只提交new_password | 返回400 | P0 |
| AUTH_PWD_UPDATE_007 | 修改密码-缺少new_password | 用户已登录 | 只提交old_password | 返回400 | P0 |
| AUTH_PWD_UPDATE_008 | 修改密码后旧token失效 | 密码已修改 | 用修改前的token访问 | 返回401 | P1 |
| AUTH_PWD_UPDATE_009 | 修改密码后新token可用 | 密码已修改 | 用新密码登录获取新token | 新token可用 | P0 |

---

## 2. 用户管理模块测试场景

### 2.1 用户资料测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| USER_PROFILE_001 | 获取本人资料 | 已登录 | GET /api/user/me | 返回200，包含id/username/email/name | P0 |
| USER_PROFILE_002 | 获取资料-未认证 | 无token | GET /api/user/me | 返回401 | P0 |
| USER_PROFILE_003 | 更新本人名字 | 已登录 | PUT /api/user/me，name="新名字" | 返回200，名字已更新 | P0 |
| USER_PROFILE_004 | 更新本人邮箱 | 已登录 | PUT /api/user/me，email="new@email.com" | 返回200 | P0 |
| USER_PROFILE_005 | 更新本人资料-空名字 | 已登录 | PUT /api/user/me，name="" | 返回200或400 | P2 |
| USER_PROFILE_006 | 更新本人资料-无效邮箱 | 已登录 | PUT /api/user/me，email="invalid" | 返回400 | P1 |
| USER_PROFILE_007 | 更新本人资料-未认证 | 无token | PUT /api/user/me | 返回401 | P0 |
| USER_PROFILE_008 | 获取他人资料 | 已登录 | GET /api/admin/users/:id | 返回200或403 | P2 |

### 2.2 用户会话测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| USER_SESSION_001 | 获取会话列表 | 已登录 | GET /api/user/me/sessions | 返回200 | P1 |
| USER_SESSION_002 | 撤销所有会话 | 已登录 | DELETE /api/user/me/sessions | 返回200 | P0 |
| USER_SESSION_003 | 撤销会话后token失效 | 会话已撤销 | 用撤销前的token访问 | 返回401 | P0 |
| USER_SESSION_004 | 撤销指定会话 | 已登录，有多会话 | DELETE /api/user/me/sessions/:id | 返回200 | P1 |
| USER_SESSION_005 | 撤销会话-未认证 | 无token | DELETE /api/user/me/sessions | 返回401 | P0 |
| USER_SESSION_006 | 撤销不存在会话 | 已登录 | DELETE /api/user/me/sessions/nonexistent | 返回404 | P1 |

---

## 3. OIDC协议测试场景

### 3.1 OIDC发现端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_DISC_001 | 获取openid-configuration | 无 | GET /.well-known/openid-configuration | 返回200，包含所有必需字段 | P0 |
| OIDC_DISC_002 | configuration包含issuer | 无 | 验证issuer字段存在且正确 | issuer非空 | P0 |
| OIDC_DISC_003 | configuration包含授权端点 | 无 | 验证authorization_endpoint存在 | authorization_endpoint非空 | P0 |
| OIDC_DISC_004 | configuration包含token端点 | 无 | 验证token_endpoint存在 | token_endpoint非空 | P0 |
| OIDC_DISC_005 | configuration包含userinfo端点 | 无 | 验证userinfo_endpoint存在 | userinfo_endpoint非空 | P0 |
| OIDC_DISC_006 | configuration包含jwks_uri | 无 | 验证jwks_uri存在 | jwks_uri非空 | P0 |
| OIDC_DISC_007 | configuration支持code响应类型 | 无 | 验证response_types_supported包含"code" | 包含"code" | P0 |
| OIDC_DISC_008 | configuration支持openid作用域 | 无 | 验证scopes_supported包含"openid" | 包含"openid" | P0 |
| OIDC_DISC_009 | configuration支持RS256算法 | 无 | 验证id_token_signing_alg_values_supported包含"RS256" | 包含"RS256" | P0 |

### 3.2 JWKS端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_JWKS_001 | 获取JWKS | 无 | GET /oidc/jwks | 返回200 | P0 |
| OIDC_JWKS_002 | JWKS包含RSA密钥 | 无 | 验证keys数组包含kty="RSA" | 至少一个RSA密钥 | P0 |
| OIDC_JWKS_003 | JWKS密钥用途为签名 | 无 | 验证use="sig" | 所有密钥用于签名 | P0 |
| OIDC_JWKS_004 | JWKS使用RS256算法 | 无 | 验证alg="RS256" | 算法正确 | P0 |
| OIDC_JWKS_005 | JWKS包含 modulus和exponent | 无 | 验证n和e字段存在 | n和e非空 | P0 |
| OIDC_JWKS_006 | JWKS包含kid | 无 | 验证kid字段存在 | kid非空 | P1 |
| OIDC_JWKS_007 | JWKS缓存和轮换 | 无 | 多次获取，对比kid | kid可能变化(密钥轮换) | P2 |

### 3.3 OIDC授权端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_AUTH_001 | 授权请求-有效参数 | Client已创建 | GET /oidc/auth?client_id=xxx&redirect_uri=xxx&response_type=code&scope=openid | 302重定向到redirect_uri带uid | P0 |
| OIDC_AUTH_002 | 授权请求-缺少client_id | 无 | 授权请求无client_id | 400 | P0 |
| OIDC_AUTH_003 | 授权请求-缺少redirect_uri | 无 | 授权请求无redirect_uri | 400 | P0 |
| OIDC_AUTH_004 | 授权请求-无效client_id | 无 | client_id为"invalid" | 400 | P0 |
| OIDC_AUTH_005 | 授权请求-无效redirect_uri | Client已创建 | redirect_uri与注册不匹配 | 400 | P0 |
| OIDC_AUTH_006 | 授权请求-无效response_type | Client已创建 | response_type为"invalid" | 400 | P0 |
| OIDC_AUTH_007 | 授权请求-支持PKCE | Client已创建 | 请求包含code_challenge和code_challenge_method | 接受PKCE参数 | P1 |
| OIDC_AUTH_008 | 授权请求-缺少scope | Client已创建 | scope参数为空 | 使用默认scope | P2 |
| OIDC_AUTH_009 | 授权请求-state参数 | Client已创建 | 包含state="xyz" | state在回调中返回 | P1 |
| OIDC_AUTH_010 | 授权请求-nonce参数 | Client已创建 | scope包含openid时包含nonce | nonce在id_token中 | P1 |

### 3.4 OIDC Token端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_TOKEN_001 | Token请求-authorization_code | 已授权 | POST /oidc/token grant_type=authorization_code | 返回access_token/refresh_token/id_token | P0 |
| OIDC_TOKEN_002 | Token请求-refresh_token | 已获取refresh_token | grant_type=refresh_token | 返回新access_token | P0 |
| OIDC_TOKEN_003 | Token请求-缺少grant_type | 无 | POST无grant_type | 400 | P0 |
| OIDC_TOKEN_004 | Token请求-invalid_code | 无 | 使用无效authorization_code | 400 | P0 |
| OIDC_TOKEN_005 | Token请求-expired_code | code已过期 | 使用过期code | 400 | P1 |
| OIDC_TOKEN_006 | Token请求-refresh_token无效 | 无 | 使用无效refresh_token | 400 | P0 |
| OIDC_TOKEN_007 | Token请求-GET方法 | 无 | GET /oidc/token | 返回错误提示用POST | P1 |
| OIDC_TOKEN_008 | Token响应-id_token格式 | 授权码模式 | 验证id_token为有效JWT | JWT格式正确 | P0 |
| OIDC_TOKEN_009 | Token响应-access_token格式 | 授权码模式 | 验证access_token存在 | access_token非空 | P0 |
| OIDC_TOKEN_010 | Token响应-refresh_token存在 | 授权码模式 | 验证refresh_token存在 | refresh_token非空 | P0 |
| OIDC_TOKEN_011 | Token响应-token_type | 授权码模式 | 验证token_type="Bearer" | token_type正确 | P0 |
| OIDC_TOKEN_012 | Token响应-expiry | 授权码模式 | 验证expires_in存在 | expires_in>0 | P0 |

### 3.5 OIDC UserInfo端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_USERINFO_001 | UserInfo请求-有效token | 有access_token | GET /oidc/userinfo with Bearer token | 返回200，用户信息 | P0 |
| OIDC_USERINFO_002 | UserInfo请求-无token | 无 | GET /oidc/userinfo无Authorization头 | 401 | P0 |
| OIDC_USERINFO_003 | UserInfo请求-invalid token | 无 | 使用无效token | 401 | P0 |
| OIDC_USERINFO_004 | UserInfo请求-malformed token | 无 | Authorization格式错误 | 401 | P0 |
| OIDC_USERINFO_005 | UserInfo返回标准声明 | 有效token | 验证sub/email/name等 | 包含标准OIDC声明 | P0 |
| OIDC_USERINFO_006 | UserInfo请求-POST方法 | 有效token | POST /oidc/userinfo | 可能支持或不支持 | P2 |

### 3.6 OIDC Token撤销测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_REVOKE_001 | 撤销token-有效token | 有access_token | POST /oidc/token/revocation | 返回200 | P0 |
| OIDC_REVOKE_002 | 撤销token-无token | 无 | POST无token参数 | 400 | P0 |
| OIDC_REVOKE_003 | 撤销token-无效token | 无 | 撤销不存在的token | 200(防用户枚举) | P0 |
| OIDC_REVOKE_004 | 撤销后token失效 | token已撤销 | 用撤销的token访问userinfo | 401 | P1 |

### 3.7 OIDC交互端点测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| OIDC_INTERACT_001 | 获取交互信息-有效uid | 有活跃交互 | GET /oidc/interaction/:uid | 返回200，交互详情 | P1 |
| OIDC_INTERACT_002 | 获取交互信息-无效uid | 无 | GET /oidc/interaction/invalid | 404 | P0 |
| OIDC_INTERACT_003 | 确认交互-有效uid | 有交互 | POST /oidc/interaction/:uid/confirm | 200，授权完成 | P1 |
| OIDC_INTERACT_004 | 确认交互-无效uid | 无 | POST /oidc/interaction/invalid/confirm | 404 | P0 |
| OIDC_INTERACT_005 | 取消交互-有效uid | 有交互 | DELETE /oidc/interaction/:uid/cancel | 200，交互取消 | P1 |
| OIDC_INTERACT_006 | 取消交互-无效uid | 无 | DELETE /oidc/interaction/invalid/cancel | 404 | P0 |

---

## 4. 管理后台测试场景

### 4.1 管理员用户管理测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ADMIN_USER_001 | 管理员列出所有用户 | 管理员已登录 | GET /api/admin/users | 返回200，用户列表 | P0 |
| ADMIN_USER_002 | 管理员获取用户数量 | 管理员已登录 | GET /api/admin/users/count | 返回200，count>0 | P1 |
| ADMIN_USER_003 | 管理员创建用户 | 管理员已登录 | POST /api/admin/users | 返回200，user_id | P0 |
| ADMIN_USER_004 | 管理员获取指定用户 | 有用户存在 | GET /api/admin/users/:id | 返回200 | P0 |
| ADMIN_USER_005 | 管理员更新用户 | 有用户存在 | PUT /api/admin/users/:id | 返回200 | P0 |
| ADMIN_USER_006 | 管理员删除用户 | 有用户存在 | DELETE /api/admin/users/:id | 返回200 | P0 |
| ADMIN_USER_007 | 管理员不能删除自己 | 管理员已登录 | DELETE /api/admin/users/self | 400 | P0 |
| ADMIN_USER_008 | 管理员重置用户密码 | 有用户存在 | POST /api/admin/users/:id/reset-password | 返回200 | P0 |
| ADMIN_USER_009 | 管理员审批用户 | 有待审批用户 | POST /api/admin/users/:id/approve | 返回200 | P0 |
| ADMIN_USER_010 | 管理员获取不存在的用户 | 无 | GET /api/admin/users/nonexistent | 404 | P0 |
| ADMIN_USER_011 | 管理员更新不存在的用户 | 无 | PUT /api/admin/users/nonexistent | 404 | P0 |
| ADMIN_USER_012 | 管理员删除不存在的用户 | 无 | DELETE /api/admin/users/nonexistent | 404 | P0 |
| ADMIN_USER_013 | 管理员创建用户-缺少字段 | 管理员已登录 | POST缺少username | 400 | P0 |
| ADMIN_USER_014 | 管理员创建重复用户名 | 有用户存在 | POST username已存在 | 400 | P0 |
| ADMIN_USER_015 | 管理员创建用户-弱密码 | 管理员已登录 | password为"123" | 400或200(配置相关) | P1 |
| ADMIN_USER_016 | 管理员更新用户邮箱 | 有用户存在 | PUT email="new@email.com" | 返回200 | P0 |
| ADMIN_USER_017 | 管理员更新用户审批状态 | 有用户存在 | PUT approved=true | 返回200 | P0 |
| ADMIN_USER_018 | 普通用户不能访问管理员接口 | 普通用户已登录 | GET /api/admin/users | 403 | P0 |
| ADMIN_USER_019 | 未认证用户不能访问管理员接口 | 无token | GET /api/admin/users | 401 | P0 |

### 4.2 管理员用户组管理测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ADMIN_GROUP_001 | 管理员创建用户组 | 管理员已登录 | POST /api/admin/groups | 返回200，group_id | P0 |
| ADMIN_GROUP_002 | 管理员列出用户组 | 有用户组 | GET /api/admin/groups | 返回200，列表 | P0 |
| ADMIN_GROUP_003 | 管理员获取指定用户组 | 有用户组 | GET /api/admin/groups/:id | 返回200 | P0 |
| ADMIN_GROUP_004 | 管理员更新用户组 | 有用户组 | PUT /api/admin/groups/:id | 返回200 | P0 |
| ADMIN_GROUP_005 | 管理员删除用户组 | 有用户组 | DELETE /api/admin/groups/:id | 返回200 | P0 |
| ADMIN_GROUP_006 | 管理员创建重复组名 | 有组存在 | POST 同名 | 400或200 | P2 |
| ADMIN_GROUP_007 | 管理员获取不存在的用户组 | 无 | GET /api/admin/groups/nonexistent | 404 | P0 |
| ADMIN_GROUP_008 | 管理员更新不存在的用户组 | 无 | PUT /api/admin/groups/nonexistent | 404 | P0 |
| ADMIN_GROUP_009 | 管理员删除不存在的用户组 | 无 | DELETE /api/admin/groups/nonexistent | 404 | P0 |
| ADMIN_GROUP_010 | 管理员创建组-无描述 | 管理员已登录 | POST只有name | 返回200 | P0 |
| ADMIN_GROUP_011 | 管理员创建组-空描述 | 管理员已登录 | POST description="" | 返回200 | P0 |

### 4.3 管理员OAuth Client管理测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ADMIN_CLIENT_001 | 管理员创建Client | 管理员已登录 | POST /api/admin/clients | 返回200，client_id | P0 |
| ADMIN_CLIENT_002 | 管理员列出Client | 有Client | GET /api/admin/clients | 返回200，列表 | P0 |
| ADMIN_CLIENT_003 | 管理员获取指定Client | 有Client | GET /api/admin/clients/:id | 返回200 | P0 |
| ADMIN_CLIENT_004 | 管理员更新Client | 有Client | PUT /api/admin/clients/:id | 返回200 | P0 |
| ADMIN_CLIENT_005 | 管理员删除Client | 有Client | DELETE /api/admin/clients/:id | 返回200 | P0 |
| ADMIN_CLIENT_006 | 管理员创建Client-设置redirect_uri | 管理员已登录 | POST with redirect_uris | 返回200 | P0 |
| ADMIN_CLIENT_007 | 管理员更新Client-redirect_uri | 有Client | PUT redirect_uris | 返回200 | P0 |
| ADMIN_CLIENT_008 | 管理员更新Client-scopes | 有Client | PUT scopes="openid profile" | 返回200 | P0 |
| ADMIN_CLIENT_009 | 管理员创建重复client_id | 有Client | POST 相同client_id | 400 | P0 |
| ADMIN_CLIENT_010 | 管理员获取不存在的Client | 无 | GET /api/admin/clients/nonexistent | 404 | P0 |
| ADMIN_CLIENT_011 | 管理员更新不存在的Client | 无 | PUT /api/admin/clients/nonexistent | 404 | P0 |
| ADMIN_CLIENT_012 | 管理员删除不存在的Client | 无 | DELETE /api/admin/clients/nonexistent | 404 | P0 |

### 4.4 管理员邀请链接管理测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ADMIN_INVITE_001 | 管理员创建邀请 | 管理员已登录 | POST /api/admin/invitations | 返回200，code | P0 |
| ADMIN_INVITE_002 | 管理员列出邀请 | 有邀请 | GET /api/admin/invitations | 返回200，列表 | P0 |
| ADMIN_INVITE_003 | 管理员获取指定邀请 | 有邀请 | GET /api/admin/invitations/:id | 返回200 | P0 |
| ADMIN_INVITE_004 | 管理员删除邀请 | 有邀请 | DELETE /api/admin/invitations/:id | 返回200 | P0 |
| ADMIN_INVITE_005 | 管理员创建邀请-含用户名 | 管理员已登录 | POST email和username | 返回200 | P0 |
| ADMIN_INVITE_006 | 管理员创建邀请-仅邮箱 | 管理员已登录 | POST只有email | 返回200 | P0 |
| ADMIN_INVITE_007 | 管理员获取不存在的邀请 | 无 | GET /api/admin/invitations/nonexistent | 404 | P0 |
| ADMIN_INVITE_008 | 管理员删除不存在的邀请 | 无 | DELETE /api/admin/invitations/nonexistent | 404 | P0 |

### 4.5 管理员Proxy Auth配置测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ADMIN_PROXY_001 | 管理员创建ProxyAuth | 管理员已登录 | POST /api/admin/proxyauth | 返回200 | P0 |
| ADMIN_PROXY_002 | 管理员列出ProxyAuth | 有配置 | GET /api/admin/proxyauth | 返回200，列表 | P0 |
| ADMIN_PROXY_003 | 管理员获取指定ProxyAuth | 有配置 | GET /api/admin/proxyauth/:id | 返回200 | P0 |
| ADMIN_PROXY_004 | 管理员更新ProxyAuth | 有配置 | PUT /api/admin/proxyauth/:id | 返回200 | P0 |
| ADMIN_PROXY_005 | 管理员删除ProxyAuth | 有配置 | DELETE /api/admin/proxyauth/:id | 返回200 | P0 |
| ADMIN_PROXY_006 | 管理员创建ProxyAuth-启用状态 | 管理员已登录 | POST enabled=true | 返回200 | P0 |
| ADMIN_PROXY_007 | 管理员更新ProxyAuth-禁用 | 有配置 | PUT enabled=false | 返回200 | P0 |
| ADMIN_PROXY_008 | 管理员获取不存在的ProxyAuth | 无 | GET /api/admin/proxyauth/nonexistent | 404 | P0 |

---

## 5. 安全性和授权测试场景

### 5.1 认证中间件测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SEC_AUTH_001 | 访问受保护资源-无token | 无token | GET /api/user/me | 401 | P0 |
| SEC_AUTH_002 | 访问受保护资源-invalid token | 无 | Bearer "invalid" | 401 | P0 |
| SEC_AUTH_003 | 访问受保护资源-malformed token | 无 | Bearer "not.a.valid.jwt" | 401 | P0 |
| SEC_AUTH_004 | 访问受保护资源-expired token | 过期token | 使用过期JWT | 401 | P0 |
| SEC_AUTH_005 | 访问受保护资源-正确token | 有效token | 使用有效Bearer token | 200 | P0 |
| SEC_AUTH_006 | 访问受保护资源-Bearer大小写 | 有效token | "bearer xxx" (小写) | 200或401取决于实现 | P2 |
| SEC_AUTH_007 | 访问受保护资源-空Bearer | 有效token | "Bearer " (只有Bearer空格) | 401 | P1 |
| SEC_AUTH_008 | 访问受保护资源-token前缀错误 | 有效token | "Basic xxx" | 401 | P1 |

### 5.2 管理员授权测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SEC_ADMIN_001 | 普通用户访问admin端点 | 普通用户登录 | GET /api/admin/users | 403 Forbidden | P0 |
| SEC_ADMIN_002 | 普通用户创建admin用户 | 普通用户登录 | POST /api/admin/users | 403 | P0 |
| SEC_ADMIN_003 | 普通用户删除任意用户 | 普通用户登录 | DELETE /api/admin/users/xxx | 403 | P0 |
| SEC_ADMIN_004 | 普通用户访问admin组 | 普通用户登录 | GET /api/admin/groups | 403 | P0 |
| SEC_ADMIN_005 | 普通用户访问admin客户端 | 普通用户登录 | GET /api/admin/clients | 403 | P0 |
| SEC_ADMIN_006 | 普通用户访问admin邀请 | 普通用户登录 | GET /api/admin/invitations | 403 | P0 |
| SEC_ADMIN_007 | 普通用户访问admin ProxyAuth | 普通用户登录 | GET /api/admin/proxyauth | 403 | P0 |
| SEC_ADMIN_008 | 未认证用户访问admin | 无token | GET /api/admin/users | 401 | P0 |
| SEC_ADMIN_009 | 管理员访问admin正常 | 管理员登录 | GET /api/admin/users | 200 | P0 |
| SEC_ADMIN_010 | 修改自己为管理员 | 普通用户登录 | 尝试将自己设为admin | 403或400 | P1 |
| SEC_ADMIN_011 | 尝试越权访问其他用户数据 | 普通用户登录 | PUT /api/admin/users/other_id | 403 | P0 |

### 5.3 JWT安全测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SEC_JWT_001 | JWT签名验证-有效签名 | 有效JWT | 使用正确签名的token | 200 | P0 |
| SEC_JWT_002 | JWT签名验证-无效签名 | 篡改token | 修改signature | 401 | P0 |
| SEC_JWT_003 | JWT过期时间验证 | 过期JWT | 使用exp过期的token | 401 | P0 |
| SEC_JWT_004 | JWT生效时间验证 | nbf未来 | 使用not_before未来的token | 401或200取决于容差 | P1 |
| SEC_JWT_005 | JWT issuer验证 | 篡改iss | 修改issuer claim | 401 | P0 |
| SEC_JWT_006 | JWT audience验证 | 篡改aud | 修改audience claim | 401 | P0 |
| SEC_JWT_007 | 密码修改后旧token失效 | token已颁发 | 修改密码后用旧token | 401 | P0 |
| SEC_JWT_008 | 会话撤销后token失效 | token已颁发 | 撤销所有会话后用token | 401 | P0 |
| SEC_JWT_009 | JWT包含必要声明 | 有效JWT | 解析token验证标准声明 | 包含sub/exp/iat等 | P0 |

### 5.4 输入验证和安全测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SEC_INPUT_001 | SQL注入-登录 | 无 | username输入SQL注入 | 401，无SQL错误信息泄露 | P0 |
| SEC_INPUT_002 | SQL注入-注册 | 无 | username参数包含SQL | 400，无SQL错误 | P0 |
| SEC_INPUT_003 | SQL注入-邮箱 | 无 | email参数包含SQL | 400 | P0 |
| SEC_INPUT_004 | XSS防护-用户名 | 无 | username包含<script> | 400或特殊字符转义 | P0 |
| SEC_INPUT_005 | XSS防护-邮箱 | 无 | email包含XSS payload | 400 | P0 |
| SEC_INPUT_006 | 路径遍历-用户ID | 已登录 | user_id包含"../" | 404或400 | P1 |
| SEC_INPUT_007 | 参数污染-Content-Type | 无 | 同时发送application/json和form | 正确处理 | P2 |
| SEC_INPUT_008 | JSON注入 | 无 | JSON body包含额外字段 | 忽略额外字段或400 | P2 |
| SEC_INPUT_009 | HTTP头注入 | 无 | 邮箱包含\r\n | 400或转义处理 | P1 |
| SEC_INPUT_010 | 批量赋值防护 | 已登录 | 尝试提交不属于自己的字段 | 忽略或400 | P1 |

### 5.5 CORS和Headers测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SEC_CORS_001 | CORS预检请求-OPTIONS | 无 | OPTIONS请求带Origin | 包含CORS headers | P0 |
| SEC_CORS_002 | CORS-有效Origin | 无 | 请求带允许的Origin | Access-Control-Allow-Origin正确 | P0 |
| SEC_CORS_003 | CORS-无效Origin | 无 | 请求带未允许的Origin | 无CORS头或拒绝 | P0 |
| SEC_CORS_004 | CORS-Allow-Methods | 预检 | OPTIONS检查Allowed-Methods | 包含POST/GET/PUT/DELETE | P1 |
| SEC_CORS_005 | CORS-Allow-Headers | 预检 | OPTIONS检查Allowed-Headers | 包含Content-Type等 | P1 |
| SEC_CORS_006 | CORS-Allow-Credentials | 带cookie | 请求带credentials | Access-Control-Allow-Credentials正确 | P1 |
| SEC_HEAD_001 | 安全头-X-Content-Type-Options | 无 | GET /api/health | X-Content-Type-Options: nosniff | P1 |
| SEC_HEAD_002 | 安全头-X-Frame-Options | 无 | GET /oidc/auth | X-Frame-Options设置 | P1 |
| SEC_HEAD_003 | 安全头-Strict-Transport-Security | HTTPS | 检查HSTS头 | 存在或无(仅HTTP环境) | P2 |
| SEC_HEAD_004 | Content-Type正确 | 无 | GET /api/health | application/json | P0 |
| SEC_HEAD_005 | Cache-Control敏感端点 | 已登录 | GET /api/user/me | no-store或no-cache | P1 |

---

## 6. 会话管理测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| SESSION_001 | 登录创建新会话 | 已注册用户 | 登录后检查会话创建 | 新会话记录创建 | P1 |
| SESSION_002 | 同一用户多设备登录 | 已注册用户 | 不同设备登录 | 多个有效会话 | P0 |
| SESSION_003 | 查看所有会话列表 | 多会话 | GET /api/user/me/sessions | 返回所有会话 | P0 |
| SESSION_004 | 撤销指定会话 | 多会话 | DELETE /api/user/me/sessions/:id | 指定会话被撤销 | P0 |
| SESSION_005 | 撤销所有会话 | 多会话 | DELETE /api/user/me/sessions | 所有会话被撤销 | P0 |
| SESSION_006 | 撤销自己不影响其他会话 | 多会话 | 撤销一个会话后其他会话仍有效 | 其他会话正常 | P0 |
| SESSION_007 | 撤销会话后token失效 | 会话已撤销 | 用被撤销session的token | 401 | P0 |
| SESSION_008 | 修改密码撤销所有会话 | 多会话 | 修改密码 | 所有会话被撤销 | P0 |
| SESSION_009 | 会话超时 | token过期 | 使用过期session token | 401 | P0 |
| SESSION_010 | 会话持久化 | 服务器重启 | 重启后验证session仍有效 | session有效(取决于存储) | P2 |

---

## 7. TOTP和MFA测试场景

### 7.1 TOTP注册测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| MFA_TOTP_REG_001 | 注册TOTP | 已登录 | POST /api/totp/registration | 200，返回secret和QRCode | P0 |
| MFA_TOTP_REG_002 | 注册TOTP-未认证 | 无token | POST /api/totp/registration | 401 | P0 |
| MFA_TOTP_REG_003 | 注册TOTP-secret格式 | 已登录 | 检查返回的secret | Base32编码格式 | P0 |
| MFA_TOTP_REG_004 | 注册TOTP-qrcode_uri格式 | 已登录 | 检查QRCode URI | otpauth://totp/格式正确 | P0 |
| MFA_TOTP_REG_005 | 重复注册TOTP | 已有TOTP | POST /api/totp/registration | 200，生成新secret | P1 |
| MFA_TOTP_REG_006 | 注册TOTP后验证 | TOTP已注册 | 使用正确code验证 | 验证成功 | P0 |
| MFA_TOTP_REG_007 | 注册TOTP后无效code | TOTP已注册 | 使用错误code验证 | 验证失败 | P0 |

### 7.2 TOTP验证测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| MFA_TOTP_VERIFY_001 | 验证有效TOTP | TOTP已注册 | POST /api/totp/verify token=有效code | 200 | P0 |
| MFA_TOTP_VERIFY_002 | 验证无效TOTP | TOTP已注册 | token="000000" | 401或400 | P0 |
| MFA_TOTP_VERIFY_003 | 验证TOTP-未认证 | 无token | POST /api/totp/verify | 401 | P0 |
| MFA_TOTP_VERIFY_004 | 验证TOTP-empty token | 已登录 | token="" | 400 | P0 |
| MFA_TOTP_VERIFY_005 | 验证TOTP-过期code | TOTP已注册 | 使用30秒前的code | 验证失败 | P0 |
| MFA_TOTP_VERIFY_006 | 验证TOTP-未来window | TOTP已注册 | 使用1个period后的code | 取决于服务器配置 | P1 |
| MFA_TOTP_VERIFY_007 | 验证TOTP-time drift | TOTP已注册 | 手动计算正确code验证 | 验证成功 | P0 |

### 7.3 TOTP删除测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| MFA_TOTP_DEL_001 | 删除TOTP-有效验证 | TOTP已注册 | DELETE /api/totp + 有效token | 200，TOTP删除 | P0 |
| MFA_TOTP_DEL_002 | 删除TOTP-invalid token | TOTP已注册 | DELETE /api/totp + 错误code | 401/400 | P0 |
| MFA_TOTP_DEL_003 | 删除TOTP-未认证 | 无token | DELETE /api/totp | 401 | P0 |
| MFA_TOTP_DEL_004 | 删除TOTP-无TOTP注册 | 用户无TOTP | DELETE /api/totp | 404或400 | P0 |
| MFA_TOTP_DEL_005 | 删除TOTP后登录行为 | TOTP已删除 | 登录检查是否还要求MFA | 不再要求MFA | P0 |

### 7.4 MFA强制测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| MFA_FORCE_001 | 启用MFARequired配置 | 系统配置 | 用户登录检查mfa_required | 根据配置返回 | P1 |
| MFA_FORCE_002 | MFA强制用户注册TOTP | MFARequired=true | 用户登录 | 要求注册TOTP | P1 |
| MFA_FORCE_003 | 无TOTP用户受限操作 | 用户无TOTP | 访问受保护操作 | 可能受限或提示 | P1 |

---

## 8. Passkey测试场景

### 8.1 Passkey认证测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PASSKEY_AUTH_001 | Passkey认证开始-用户存在 | 用户已注册 | POST /api/auth/passkey/start username=xxx | 200，返回challenge | P0 |
| PASSKEY_AUTH_002 | Passkey认证开始-用户不存在 | 无 | username不存在 | 400 | P0 |
| PASSKEY_AUTH_003 | Passkey认证开始-缺少username | 无 | 无username参数 | 400 | P0 |
| PASSKEY_AUTH_004 | Passkey认证结束-有效响应 | 已开始认证 | POST /api/auth/passkey/end | 返回token | P0 |
| PASSKEY_AUTH_005 | Passkey认证结束-无效响应 | 已开始认证 | credential响应无效 | 400 | P0 |
| PASSKEY_AUTH_006 | Passkey认证结束-未开始 | 无challenge | 直接调用end | 400 | P0 |

### 8.2 Passkey注册测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PASSKEY_REG_001 | Passkey注册开始 | 已登录 | POST /api/passkey/registration/start | 200，返回options | P0 |
| PASSKEY_REG_002 | Passkey注册开始-未认证 | 无token | POST /api/passkey/registration/start | 401 | P0 |
| PASSKEY_REG_003 | Passkey注册完成 | 已开始注册 | POST /api/passkey/registration/end | 200 | P0 |
| PASSKEY_REG_004 | Passkey注册完成-无效 | 无效credential | POST无效options | 400 | P0 |
| PASSKEY_REG_005 | Passkey注册完成-未开始 | 无 | 直接调用end | 400 | P0 |

### 8.3 Passkey管理测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PASSKEY_MGMT_001 | 列出用户Passkeys | 已注册Passkey | GET /api/passkey | 返回passkey列表 | P0 |
| PASSKEY_MGMT_002 | 列出Passkeys-新用户无 | 新用户无Passkey | GET /api/passkey | 返回空列表 | P0 |
| PASSKEY_MGMT_003 | 列出Passkeys-未认证 | 无token | GET /api/passkey | 401 | P0 |
| PASSKEY_MGMT_004 | 删除指定Passkey | 有Passkey | DELETE /api/passkey/:id | 200 | P0 |
| PASSKEY_MGMT_005 | 删除Passkey-不存在 | 无Passkey | DELETE /api/passkey/nonexistent | 404 | P0 |
| PASSKEY_MGMT_006 | 删除他人Passkey | 其他用户有Passkey | 尝试删除他人Passkey | 403 | P0 |
| PASSKEY_MGMT_007 | 删除最后Passkey | 只有一个Passkey | 删除后检查行为 | 可能受限或警告 | P1 |

---

## 9. 邮件验证测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| EMAIL_VER_001 | 发送验证邮件-用户存在 | 用户已注册 | POST /api/auth/send_verify_email | 200 | P0 |
| EMAIL_VER_002 | 发送验证邮件-用户不存在 | 无 | POST不存在的邮箱 | 404(防枚举) | P0 |
| EMAIL_VER_003 | 验证邮箱-有效code | 有待验证 | POST /api/auth/verify_email | 200，邮箱验证通过 | P0 |
| EMAIL_VER_004 | 验证邮箱-无效code | 有待验证 | code无效 | 400 | P0 |
| EMAIL_VER_005 | 验证邮箱-empty user_id | 无 | user_id为空 | 400 | P0 |
| EMAIL_VER_006 | 验证邮箱-empty challenge | 无 | challenge为空 | 400 | P0 |
| EMAIL_VER_007 | 验证邮箱-已验证用户 | 邮箱已验证 | 再次验证 | 可能200或400 | P2 |
| EMAIL_VER_008 | 验证邮箱-过期code | code过期 | 使用过期验证code | 400 | P1 |
| EMAIL_VER_009 | 验证后用户email_verified=true | 验证成功 | 查询用户状态 | email_verified=true | P0 |
| EMAIL_VER_010 | 未验证邮箱用户受限操作 | 邮箱未验证 | 尝试需要验证的操作 | 根据配置受限 | P1 |

---

## 10. 邀请注册测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| INVITE_001 | 获取邀请信息-有效code | 有邀请 | GET /api/auth/invitation/:id/:challenge | 200，返回邀请详情 | P0 |
| INVITE_002 | 获取邀请信息-无效code | 无 | code无效 | 404 | P0 |
| INVITE_003 | 获取邀请信息-expired | 邀请已过期 | 过期邀请code | 404 | P1 |
| INVITE_004 | 使用邀请注册-新用户 | 有有效邀请 | 注册并提供invitation_id | 注册成功，关联邀请 | P0 |
| INVITE_005 | 使用邀请注册-邀请已用 | 邀请已使用 | 再次使用同一邀请 | 400或错误 | P0 |
| INVITE_006 | 使用邀请注册-用户已存在 | 用户已存在 | 用邀请注册同名用户 | 400 | P0 |
| INVITE_007 | 管理员创建邀请 | 管理员登录 | POST /api/admin/invitations | 返回invitation_id和code | P0 |
| INVITE_008 | 管理员删除邀请 | 有邀请 | DELETE /api/admin/invitations/:id | 200 | P0 |
| INVITE_009 | 管理员列出邀请 | 有邀请 | GET /api/admin/invitations | 返回列表 | P0 |
| INVITE_010 | 邀请注册后自动审批 | 邀请注册 | 注册后检查approved状态 | approved=true(取决于配置) | P2 |

---

## 11. 密码安全测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PWD_SEC_001 | 密码强度-极弱(123) | 系统启用强度检查 | 注册password="123" | 400或被拒绝 | P0 |
| PWD_SEC_002 | 密码强度-弱(password) | 系统启用强度检查 | password="password" | 400或被拒绝 | P0 |
| PWD_SEC_003 | 密码强度-中等 | 系统启用强度检查 | password="中等强度123!@" | 200或400 | P1 |
| PWD_SEC_004 | 密码强度-强 | 系统启用强度检查 | password="Str0ng!@#$%^&*()" | 200 | P1 |
| PWD_SEC_005 | 密码包含用户名 | 系统启用强度检查 | password包含username | 被拒绝或警告 | P0 |
| PWD_SEC_006 | 密码哈希验证 | 注册用户 | 检查数据库中密码哈希不是明文 | 密码哈希安全存储 | P0 |
| PWD_SEC_007 | 密码哈希算法 | 检查数据库 | 验证使用bcrypt/argon2等 | 使用强哈希算法 | P0 |
| PWD_SEC_008 | 密码不能相同 | 修改密码 | new_password=old_password | 400或警告 | P1 |
| PWD_SEC_009 | 密码修改历史 | 系统配置 | 尝试使用最近使用过的密码 | 被拒绝(如果配置) | P2 |

---

## 12. OAuth Client管理测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| CLIENT_001 | 创建Client-最小参数 | 管理员登录 | POST client_id和name | 200 | P0 |
| CLIENT_002 | 创建Client-完整参数 | 管理员登录 | POST包含redirect_uris/scopes | 200 | P0 |
| CLIENT_003 | 创建Client-无client_id | 管理员登录 | POST无client_id | 400 | P0 |
| CLIENT_004 | 创建Client-无name | 管理员登录 | POST无name | 400 | P0 |
| CLIENT_005 | 创建Client-重复client_id | 已存在 | POST相同client_id | 400 | P0 |
| CLIENT_006 | 更新Client-redirect_uri | 有Client | PUT添加新redirect_uri | 200 | P0 |
| CLIENT_007 | 更新Client-无效redirect_uri | 有Client | PUT无效uri格式 | 400 | P1 |
| CLIENT_008 | OIDC授权-未注册redirect_uri | Client有已注册uri | 使用未注册的uri | 400 | P0 |
| CLIENT_009 | OIDC授权-空redirect_uri | Client有uri | redirect_uri为空 | 400 | P0 |
| CLIENT_010 | Client删除-有活跃授权 | 有用户授权 | DELETE该Client | 可能失败或成功(级联) | P2 |
| CLIENT_011 | 列出Client-分页 | 多个Client | GET /api/admin/clients | 返回分页结果 | P1 |
| CLIENT_012 | Client认证-methods | 无 | 检查client认证方式 | 支持client_secret_basic等 | P2 |

---

## 13. Proxy Auth测试场景

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PROXY_001 | 创建ProxyAuth配置 | 管理员登录 | POST包含proxy_url/header_name | 200 | P0 |
| PROXY_002 | ProxyAuth-启用禁用 | 有配置 | PUT enabled切换状态 | 200 | P0 |
| PROXY_003 | ProxyAuth-转发头部 | 用户登录 | 访问proxy配置的端点 | 头部正确转发 | P1 |
| PROXY_004 | ProxyAuth-无效URL | 管理员登录 | POST无效proxy_url | 400 | P1 |
| PROXY_005 | ProxyAuth-缺少header_name | 管理员登录 | POST无header_name | 400 | P0 |
| PROXY_006 | ProxyAuth-列出发送的头部 | 有配置 | 验证转发哪些header | 包含配置的头部 | P1 |
| PROXY_007 | ProxyAuth-用户信息映射 | 有配置 | 配置头部映射用户字段 | 头部包含正确用户信息 | P1 |
| PROXY_008 | ProxyAuth-未认证请求 | proxy端点 | 不带token访问 | 401或转发错误 | P0 |
| PROXY_009 | ProxyAuth-无效配置ID | 无 | PUT /api/admin/proxyauth/invalid | 404 | P0 |
| PROXY_010 | ProxyAuth-认证后转发 | 有效token | 带token访问 | 用户信息正确转发 | P0 |

---

## 14. 边界条件和错误处理测试场景

### 14.1 HTTP方法测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ERR_HTTP_001 | GET端点-POST方法 | 无 | POST /api/health | 405 Method Not Allowed | P0 |
| ERR_HTTP_002 | POST端点-GET方法 | 无 | GET /api/auth/register | 405 | P0 |
| ERR_HTTP_003 | PUT端点-GET方法 | 无 | GET /api/user/me/password | 405 | P0 |
| ERR_HTTP_004 | DELETE端点-GET方法 | 无 | GET /api/user/me/sessions | 405 | P0 |
| ERR_HTTP_005 | 不支持的HTTP方法 | 无 | PATCH /api/user/me | 405 | P1 |

### 14.2 路径测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ERR_PATH_001 | 不存在的端点 | 无 | GET /api/nonexistent | 404 | P0 |
| ERR_PATH_002 | OIDC不存在的端点 | 无 | GET /oidc/nonexistent | 404 | P0 |
| ERR_PATH_003 | 管理员不存在的端点 | 管理员登录 | GET /api/admin/nonexistent | 404 | P0 |
| ERR_PATH_004 | URL编码的特殊字符 | 无 | GET /api/user/me%2F.. | 404或400 | P1 |
| ERR_PATH_005 | 超长路径 | 无 | GET /api/ + 1000字符 | 404或400 | P1 |

### 14.3 请求格式测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ERR_REQ_001 | 无效JSON | 无 | Body="not json" | 400 | P0 |
| ERR_REQ_002 | 空JSON body | 无 | Body="{}" | 200或400(取决于端点) | P0 |
| ERR_REQ_003 | 缺失必需字段 | 无 | Body只有部分字段 | 400 | P0 |
| ERR_REQ_004 | 字段类型错误 | 无 | Body字段类型不对(string给number) | 400或类型转换 | P0 |
| ERR_REQ_005 | 超大请求体 | 无 | Body超过10MB | 413 Payload Too Large | P1 |
| ERR_REQ_006 | 重复字段名 | 无 | JSON有重复key | 取决于解析器 | P2 |
| ERR_REQ_007 | 非常规编码 | 无 | Content-Type不是UTF-8 | 正确处理 | P2 |

### 14.4 服务器错误处理

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| ERR_SRV_001 | 数据库连接失败 | DB不可用 | 发起请求 | 500，带错误信息 | P0 |
| ERR_SRV_002 | 数据库超时 | DB慢响应 | 发起耗时查询 | 504或200 | P1 |
| ERR_SRV_003 | 内存不足 | 大量数据 | 发起消耗大请求 | 500或503 | P2 |
| ERR_SRV_004 | 文件描述符耗尽 | 大量连接 | 发起大量请求 | 503 Service Unavailable | P2 |
| ERR_SRV_005 | 错误信息不泄露 | 服务器错误 | 触发500错误 | 日志有详情，响应无敏感信息 | P0 |

---

## 15. 性能和并发测试场景

### 15.1 并发认证测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PERF_CONC_001 | 并发登录同一用户 | 已注册用户 | 10个并发登录请求 | 全部成功或正确处理 | P1 |
| PERF_CONC_002 | 并发注册不同用户 | 无 | 10个并发注册不同用户 | 全部成功 | P0 |
| PERF_CONC_003 | 并发注册相同用户名 | 无 | 10个并发注册同一用户名 | 只有1个成功 | P0 |
| PERF_CONC_004 | 并发密码修改 | 已登录用户 | 10个并发修改密码 | 只有1个成功 | P1 |
| PERF_CONC_005 | 并发获取token | 同一用户多会话 | 10个并发refresh token | 全部成功 | P1 |

### 15.2 速率限制测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PERF_RATE_001 | 登录限流-短时间多次 | 无 | 10秒内20次登录失败 | 开始限流 | P0 |
| PERF_RATE_002 | 注册限流 | 无 | 短时间内大量注册 | 开始限流 | P0 |
| PERF_RATE_003 | 限流后正常请求 | 被限流 | 等待限流恢复后 | 正常处理 | P1 |
| PERF_RATE_004 | 限流响应格式 | 被限流 | 检查429响应 | Retry-After header | P1 |
| PERF_RATE_005 | API全局限流 | 无 | 短时间内大量不同API | 触发全局限流 | P1 |

### 15.3 性能基准测试

| 测试ID | 测试场景 | 前置条件 | 测试步骤 | 预期结果 | 优先级 |
|--------|----------|----------|----------|----------|--------|
| PERF_BENCH_001 | 登录响应时间 | 用户已注册 | 测量登录API | <500ms | P1 |
| PERF_BENCH_002 | 受保护资源响应时间 | 已登录 | 测量/api/user/me | <100ms | P1 |
| PERF_BENCH_003 | OIDC发现端点响应时间 | 无 | 测量openid-configuration | <100ms | P1 |
| PERF_BENCH_004 | JWKS端点响应时间 | 无 | 测量/oidc/jwks | <100ms | P1 |
| PERF_BENCH_005 | 并发用户支持 | 无 | 50并发用户操作 | 成功率>95% | P2 |
| PERF_BENCH_006 | 大用户列表查询 | 100+用户 | GET /api/admin/users | <2s | P1 |

---

## 附录

### A. 测试数据模板

```yaml
# 用户数据模板
test_users:
  regular:
    username: "testuser"
    email: "test@example.com"
    password: "TestPass123!"
  admin:
    username: "adminuser"
    email: "admin@example.com"
    password: "AdminPass123!"
    is_admin: true
  mfa:
    username: "mfauser"
    email: "mfa@example.com"
    password: "MFAPass123!"
    totp_enabled: true

# Client数据模板  
test_clients:
  basic:
    client_id: "test-client"
    name: "Test Client"
    redirect_uris: "https://example.com/callback"
    scopes: "openid profile email"
```

### B. 测试优先级定义

| 优先级 | 定义 | 测试数量占比 |
|--------|------|-------------|
| P0 | 核心功能，必须通过 | 约30% |
| P1 | 重要功能，应该通过 | 约40% |
| P2 | 增强功能，期望通过 | 约30% |

### C. 测试状态定义

| 状态 | 定义 |
|------|------|
| Ready | 测试场景已定义，待实现 |
| Implemented | 测试已实现 |
| Passing | 测试通过 |
| Failing | 测试失败，需修复 |
| Blocked | 测试被阻塞，等待其他依赖 |

### D. 安全测试清单

- [ ] SQL注入防护
- [ ] XSS防护
- [ ] CSRF防护
- [ ] 会话固定
- [ ] 敏感数据暴露
- [ ] 认证绕过
- [ ] 授权绕过
- [ ] 密码强度
- [ ] Token安全
- [ ] CORS配置
- [ ] 安全头
- [ ] 错误信息泄露
- [ ] 暴力破解防护
- [ ] 账户锁定
- [ ] 敏感操作日志

---

## 变更历史

| 版本 | 日期 | 变更描述 |
|------|------|----------|
| 1.0.0 | 2026-04-19 | 初始版本，创建完整E2E测试场景文档 |
