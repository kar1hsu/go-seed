# System 领域服务

第一阶段实现独立的 `system-rpc`、GORM MySQL 连接池和 Redis 客户端。`internal/application/health` 持有依赖就绪用例，RPC logic 负责协议适配。未来管理员、角色、权限与任务业务进入本领域；当前没有业务表或数据库迁移。

从仓库根目录按 [README](../../README.md) 启动。单独启动时先注入 `.env` 环境变量：

```bash
go run ./services/system/rpc -f services/system/rpc/etc/system.yaml
# make build-go 后
bin/probe -rpc 127.0.0.1:19001          # 标准 gRPC Health 存活检查
bin/probe -rpc 127.0.0.1:19001 -ready   # 加上 MySQL/Redis 检查
```

配置见 [rpc/etc/system.yaml](rpc/etc/system.yaml)：`SYSTEM_RPC_HOST/PORT`、`MYSQL_*`、`REDIS_HOST/PORT`。本地端口 19001，容器内部 9001。MySQL 连接池默认最大 20、空闲 5、最大生命周期 30 分钟，UTC 时间和 utf8mb4；应用通过 GORM 使用该连接池，schema 只能由后续独立迁移任务修改。

依赖延迟连接允许进程在数据库故障时存活；`CheckReady` 会实际检查依赖并返回 Unavailable。就绪检查上限 1.5 秒，保留调用方更短的 deadline。Redis 采用 go-zero 自带客户端与连接管理，禁止新增全局可变连接。

协议见 [rpc/RPC.md](rpc/RPC.md)。运行 `go test ./services/system/...` 检查依赖失败与取消；`make smoke` 验证真实 MySQL/Redis。队列与定时任务使用 Asynq 的方向不变，worker/scheduler 留到第三阶段实现。
