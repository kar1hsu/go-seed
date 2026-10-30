# app-api

面向业务客户端的 HTTP 入口。提供基础探测与 C 端会员注册、登录、刷新、退出、个人资料和改密；只通过公开 RPC 客户端访问 system/member，不装配数据库和领域 repository。

从仓库根目录按 [README](../../../README.md) 配置 `.env`、启动依赖并执行 `make dev`。单独运行时先把 `.env` 中的变量注入进程环境，再执行：

```bash
go run ./apps/app/api -f apps/app/api/etc/app-api.yaml
curl http://127.0.0.1:8080/api/v1/system/ping
```

配置位于 [etc/app-api.yaml](etc/app-api.yaml)：`API_HOST`、`APP_API_PORT` 指定本地监听，`SYSTEM_RPC_HOST/PORT` 与 `MEMBER_RPC_HOST/PORT` 指定 RPC 地址。system RPC 超时 2 秒、member RPC 3 秒，HTTP 超时 3 秒，入口统一采用 go-zero 环境变量加载。默认配置为 dev 模式。

- `GET /healthz`：存活，200 与 `{"status":"ok"}`。
- `GET /readyz`：检查 RPC、MySQL、Redis，200 为就绪，503 为不可用；只返回 status，不泄露内部连接信息。
- 业务探测详见 [API.md](API.md) 和 [OpenAPI](../../../docs/openapi/app.openapi.json)。

执行 `go test ./apps/app/api/...` 验证生成 handler → gRPC → HTTP 的成功与失败映射；真实依赖使用 `make smoke`。退出由进程监督器统一处理，下游连接在退出时释放。

会员契约及示例见[会员模块](../../../docs/member-module.md)，[生成类型](../../../web/admin/src/api/generated/app-schema.ts) 和 [client 示例](../../../web/admin/src/api/member-client.ts) 可供 C 端使用。每个写接口要求 CSRF 请求头；会员 Cookie 为 `goseed_member_refresh`。真实链路执行 `make member-smoke`。
