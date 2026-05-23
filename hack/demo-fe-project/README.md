# demo-fe-project

一个基于 React 19 + Vite + TypeScript 的前端基座项目，内置路由、状态管理、请求封装、UI 组件库、工程化规范，以及 ARMS RUM 监控与 SourceMap 自动解析能力。

## 技术栈

- React 19
- TypeScript
- Vite 8
- React Router 7
- Zustand 5
- Axios
- Ant Design 6 + Less
- ARMS RUM（`@arms/rum-browser`）
- ARMS SourceMap 插件（`@arms/rum-vite-plugin`）
- ESLint + Prettier
- Husky + lint-staged + commitlint

## 快速开始

### 环境要求

- Node.js 18+
- pnpm 10+

### 安装与启动

```bash
pnpm install
pnpm dev
```

默认访问地址：`http://localhost:5173`

## 常用命令

```bash
pnpm dev          # 本地开发
pnpm build        # 生产构建（包含 TS 构建 + Vite 构建）
pnpm preview      # 预览构建产物
pnpm typecheck    # TypeScript 类型检查
pnpm lint         # ESLint 检查
pnpm lint:fix     # ESLint 自动修复
pnpm format       # Prettier 格式化
pnpm format:check # Prettier 检查
```

## 项目结构

```text
src/
  app/            # 应用入口组件
  assets/         # 静态资源
  layouts/        # 页面布局
  monitoring/     # ARMS RUM 初始化与异常模拟
  pages/          # 页面模块
  router/         # 路由配置
  services/       # 请求封装
  store/          # 状态管理
  styles/         # 全局样式与变量
  types/          # 全局类型声明
```

## 运行时环境变量（Vite）

项目默认使用 `.env.development` 与 `.env.production`。

示例：

```env
VITE_API_BASE_URL=/api
VITE_ARMS_ENABLED=true
VITE_ARMS_ENV=prod
VITE_ARMS_ENDPOINT=https://xxx/rum/web/v2?workspace=xxx&service_id=xxx
```

变量说明：

- `VITE_API_BASE_URL`: Axios `baseURL`。
- `VITE_ARMS_ENABLED`: 是否启用 RUM（`true/false`）。
- `VITE_ARMS_ENV`: RUM 环境标识，可选 `prod|gray|pre|daily|local`。
- `VITE_ARMS_ENDPOINT`: ARMS 控制台生成的 Web/H5 接入 endpoint。

## ARMS RUM 接入说明

RUM 在应用启动时初始化（`src/main.tsx` 调用 `initRum()`）。

当前默认采集项：

- 页面性能（perf）
- Web Vitals
- API 请求
- 静态资源
- JS 异常
- Console Error
- 用户行为（action）

项目内置了 SDK 兼容兜底（`default` 导出兼容）和重复初始化保护。

## 异常模拟面板（测试专用）

首页提供了 5 类异常触发按钮：

- JS 同步异常
- Promise 未处理拒绝
- Console 错误
- API 失败
- 资源加载失败

说明：

- 点击“JS 同步异常”会出现 `Uncaught Error`，这是**预期行为**，用于验证 RUM `jsError` 采集链路。

## SourceMap 自动解析（ARMS）

本项目已按 ARMS 官方流程接入 Vite 插件：`@arms/rum-vite-plugin`。

`pnpm build` 时会：

- 给 JS 文件注入 debug UUID。
- 给对应 SourceMap 注入 `debugId`。
- 生成 SourceMap（`build.sourcemap=true`），用于异常堆栈解析。

### 阶段一：手动上传（默认）

默认只做注入，不自动上传。

1. 执行 `pnpm build`
2. 在 ARMS 控制台进入：应用设置 -> 文件管理
3. 上传 `dist/**/*.map`
4. 触发异常后在异常详情页验证堆栈是否自动解析

### 阶段二：自动上传（可选）

当启用以下构建环境变量时，构建后会自动上传 SourceMap 到 RUM OSS：

```env
ARMS_RUM_SOURCEMAP_AUTO_UPLOAD=true
ARMS_RUM_PID=your_rum_pid
ARMS_RUM_REGION=cn-hangzhou
ARMS_RUM_ACCESS_KEY_ID=your_access_key_id
ARMS_RUM_ACCESS_KEY_SECRET=your_access_key_secret
```

自动上传行为：

- `version` 默认取 `package.json` 的 `version`（用于 ARMS 分组管理）。
- 上传成功后会删除构建目录中的 `.map`（`clearSourceMap: true`）。
- 如果 `ARMS_RUM_SOURCEMAP_AUTO_UPLOAD=true` 但缺少必填项（`PID/AK`），会自动降级为手动上传模式并打印告警。

### 安全建议

- `ARMS_RUM_ACCESS_KEY_ID/SECRET` 不要写入仓库。
- 本地使用未追踪的私有环境文件；CI 使用 Secret 注入。
- 建议使用 RAM 子账号并按最小权限配置。

## 路由与页面

- `/` -> 重定向到 `/home`
- `/home` -> 首页（计数器 + RUM 模拟面板）
- `*` -> 404 页面

## 请求层约定

- `src/services/request.ts` 基于 Axios 封装。
- 默认超时 `10s`。
- 响应拦截器会将后端 `message`（若存在）包装为 `Error` 抛出。

## 提交规范

项目启用 Git Hooks：

- `pre-commit`: `lint-staged`
- `commit-msg`: `commitlint`

推荐使用 Conventional Commits：

```text
feat: add user list page
fix: handle request timeout
chore: update deps
```

## 常见问题

### 1. 控制台提示 `init is not a function`

已在 `src/monitoring/rum.ts` 中做 SDK 默认导出兼容处理；若仍出现，优先检查依赖版本与构建缓存。

### 2. `global is not defined`

已在 Vite 配置中通过 `define.global = 'globalThis'` 做浏览器兼容。

### 3. ARMS 看不到 SourceMap 自动解析

请依次检查：

1. 构建产物是否包含 `.map` 且 map 中有 `debugId`。
2. ARMS 文件管理中是否存在对应版本的 SourceMap。
3. 异常详情页该堆栈行是否携带 UUID（不是 `-`）。
4. 自动上传时 `PID/Region/AK` 是否正确，且 ARMS 控制台已开启 OSS 批量上传开关。
