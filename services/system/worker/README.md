# system-worker

消费 system 的 Asynq 队列，执行字典 CSV 导出、文件删除与过期上传清理，并持续投递数据库 outbox。共享用例位于 `services/system/internal/application/jobs`。

从仓库根目录执行 `make init infra-up migrate bootstrap`，然后 `make dev`；单独运行时加载 `services/system/worker/etc/system-worker.yaml` 所需环境变量，再执行 `bin/system-worker -f services/system/worker/etc/system-worker.yaml`。Compose 自动运行该进程。

依赖 MySQL、队列 Redis 和私有对象存储；按配置显式装配，不连接其他业务域数据库。`GET /healthz` 为进程存活，`GET /readyz` 验证三项依赖；本地默认 `http://127.0.0.1:19201/readyz`。业务操作入口在 admin-api，worker 不提供 HTTP 创建任务接口。

任务类型、payload、重试和幂等规则见[文件与任务](../../../docs/files-and-tasks.md)。运行 `make test-rbac` 验证数据库与真实 Asynq 链路，完整 HTTP 验证使用 `make rbac-smoke`。
