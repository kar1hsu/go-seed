# GoSeed 管理工作台

Vue 3 + TypeScript + Vite + Pinia + Vue Router + Element Plus。布局参考 [Vben Admin](https://github.com/vbenjs/vue-vben-admin) 浅色主题：白色可折叠侧栏、面包屑、可关闭页签、浅灰内容区和蓝色主色。保留项目自己的 feature 结构，未引入 Vben 框架。提供管理员登录、刷新恢复、退出、修改密码和真实 API/RPC 状态。角色与菜单权限功能尚未接入。

| 路由 | 页面与行为 |
| --- | --- |
| `/login` | 管理员登录，安全地恢复站内跳转目标 |
| `/account/security` | 修改密码，成功后重新登录 |
| `/` | 工作台：链路状态、响应耗时、RPC 版本、最近五次检查和快捷入口 |
| `/monitor` | 服务状态：手动探测、Trace ID/结果筛选、分页、清空记录、详情抽屉 |
| `/api-docs` | 接口文档：Ping 契约说明、同源 curl 示例、实际请求和响应 |
| `/about` | 项目信息：项目版本、依赖版本、仓库与当前实现范围 |

检查记录只存于 Pinia 内存，保留最近 20 条真实请求，刷新页面后清空。页签跟踪本次打开的页面；工作台固定，关闭当前页签返回此前页签。窄屏使用抽屉式导航。

推荐从仓库根目录执行 `make dev`，会统一加载 `.env` 中端口并启动所有进程。只运行前端时：

```bash
pnpm --dir web/admin install --frozen-lockfile
pnpm --dir web/admin dev
pnpm --dir web/admin typecheck
pnpm --dir web/admin format:check
pnpm --dir web/admin build
```

上述独立启动默认代理 `127.0.0.1:8081`，可通过进程环境的 `ADMIN_API_PORT` 和 `WEB_PORT` 调整。浏览器只请求同源 `/admin-api`，开发由 Vite 代理，Compose 由 Nginx 代理。

- `src/app` / `layouts` / `router`：入口、布局与路由。
- `src/features/auth`：管理员会话、登录页、改密页、请求认证中间件。
- `src/features/dashboard`：探测接口适配、Pinia 状态、页面。
- `src/features/api-docs` / `features/about`：接口调试与项目信息页面。
- `src/lib/http.ts`：openapi-fetch 基础客户端、5 秒超时和 APIError。
- `src/api/generated/schema.ts`：由 `.api` → goctl Swagger → OpenAPI 3 → openapi-typescript 派生，禁止手改。

执行根目录 `make gen-web` 更新派生文档与类型。Element Plus 按组件引入。请求失败清除旧结果，展示错误并允许手动重试，不自动重试写操作；access token 仅保存在 Pinia 内存；refresh token 使用 HttpOnly Cookie。刷新页面恢复会话；同页请求合并刷新，支持 Web Locks 的浏览器还会跨标签页串行轮换。退出通过 BroadcastChannel 通知其他标签页。写操作不会在 401 后自动重放。页面显示的“连接正常”只代表 API/RPC 连通，依赖就绪请访问 API 的 `/readyz`。

pnpm lockfile 固定依赖，workspace 仅允许 esbuild 构建脚本。Prettier 跳过生成目录，避免与生成器格式互相覆盖。
