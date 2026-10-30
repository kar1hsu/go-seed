# system-scheduler

投递 system 的 UTC 小时清理触发消息。只装配队列客户端，不初始化 MySQL、管理员会话或存储。调度消息由 worker 转换为持久化业务任务。

`make dev` 与 Compose 均启动该进程。单独运行时加载 `services/system/scheduler/etc/system-scheduler.yaml` 所需环境变量，再执行 `bin/system-scheduler -f services/system/scheduler/etc/system-scheduler.yaml`。

每十秒尝试当前小时，稳定 TaskID 减少重复入队，数据库唯一调度键最终去重。部署运行一份，更新时先停止旧实例。队列故障后补当前小时，不回放所有停机小时。

`GET /healthz` 为存活，`GET /readyz` 检查队列 Redis；默认 `http://127.0.0.1:19202/readyz`。配置、业务幂等和测试见[文件与任务](../../../docs/files-and-tasks.md)。
