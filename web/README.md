# AuthNas Web Frontend

AuthNas 的前端应用，基于 Vue 3 + TypeScript + Vite 构建。

## 重要说明：Vite 只是编译工具

**Vite 的作用是将前端代码编译为静态资源，不参与实际服务或测试**。

```
web/src/ (Vue/TypeScript 源码)
    ↓ npm run build (Vite 编译)
go-server/static/ (编译后的 HTML/CSS/JS)
    ↓ Go 程序读取并服务
http://localhost:8080/ (统一入口)
```

**测试的真实对象**：

- **真正的测试目标是 Go 服务器** (`localhost:8080`)
- Go server 配置了反向代理，将前端静态资源挂载到根路径
- E2E 测试直接访问 `http://localhost:8080`，由 Go 程序提供所有页面和 API
- 不需要也**不应该**单独运行前端开发服务器 (`localhost:3000`) 进行测试

### 开发模式说明

- **Vite 开发服务器** (`npm run dev`) 仅用于前端源码的热重载开发
- **Go 服务器** (`go run ./cmd/server`) 是生产环境的运行方式，同时服务 API 和静态页面
- **E2E 测试** 直接测试 Go server，无需运行 Vite 开发服务器

## 技术栈

- **框架**: Vue 3.5+ (Composition API + `<script setup>`)
- **类型系统**: TypeScript 6.0+
- **构建工具**: Vite 8.0+
- **UI 组件库**: Naive UI 2.44+
- **状态管理**: Pinia 2.3+
- **路由**: Vue Router 4.6+
- **工具库**: VueUse 14+
- **HTTP 客户端**: Axios 1.15+
- **测试**: Vitest + Playwright

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 类型检查
npm run build

# 运行单元测试
npm test

# 运行测试（监听模式）
npm run test:watch

# 运行 E2E 测试
npm run test:e2e

# 代码检查与修复
npm run lint

# 代码格式化
npm run format
```

## 项目结构

```
src/
├── api/            # API 客户端模块
├── assets/         # 静态资源
├── components/     # 公共组件
├── router/         # 路由配置
├── stores/         # Pinia 状态管理
│   └── __tests__/  # Store 单元测试
├── types/          # TypeScript 类型定义
└── views/          # 页面视图
    ├── __tests__/  # 页面单元测试
    ├── admin/      # 管理后台页面
    └── user/       # 用户页面
```

## 代码规范

- 使用 ESLint + TypeScript 进行代码检查
- 使用 Prettier 进行代码格式化
- 提交前会自动运行 lint-staged 进行检查和修复

## 构建

```bash
npm run build
```

构建产物会输出到 `../go-server/static` 目录，与后端服务集成。

## 测试说明

E2E 测试的目标是 **Go 服务器** (`localhost:8080`)，而非 Vite 开发服务器。

### Go Server E2E 测试（主要）

```bash
cd ../go-server && go test -v ./e2e/...
```

这些测试直接测试 Go 服务器，验证完整的 OIDC 流程、API 端点和认证功能。

### 前端 E2E 测试

前端 E2E 测试（Playwright）用于验证 UI 交互，但实际测试入口是 Go server：

```bash
npm run build && npm run test:e2e
```

这会先构建静态资源到 `go-server/static/`，然后启动 Go server 进行测试。

**注意**：不要运行 `npm run dev`（Vite 开发服务器）进行测试，应该测试 Go server。
