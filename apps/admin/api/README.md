# admin-api

面向管理后台的 HTTP 入口。本阶段提供 `/system/ping`；只通过公开 RPC 客户端访问 system，不装配数据库和领域 repository。

从仓库根目录按 [README](../../../README.md) 配置 `.env`、启动依赖并执行 `make dev`。单独运行时先把 `.env` 中的变量注入进程环境，再执行：

```bash
go run ./apps/admin/api -f apps/admin/api/etc/admin-api.yaml
curl http://127.0.0.1:8081/admin-api/v1/system/ping
```

配置位于 [etc/admin-api.yaml](etc/admin-api.yaml)：`API_HOST`、`ADMIN_API_PORT` 指定本地监听，`SYSTEM_RPC_HOST/PORT` 指定 RPC 地址。RPC 超时 2 秒，HTTP 超时 3 秒，入口统一采用 go-zero 环境变量加载。默认配置为 dev 模式。

- `GET /healthz`：存活，200 与 `{"status":"ok"}`。
- `GET /readyz`：检查 RPC、MySQL、Redis，200 为就绪，503 为不可用；只返回 status，不泄露内部连接信息。
- 业务探测详见 [API.md](API.md) 和 [OpenAPI](../../../docs/openapi/admin.openapi.json)。

执行 `go test ./apps/admin/api/...` 验证生成 handler → gRPC → HTTP 的成功与失败映射；真实依赖使用 `make smoke`。退出由进程监督器统一处理，下游连接在退出时释放。
