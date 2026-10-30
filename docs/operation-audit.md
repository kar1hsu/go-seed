# 操作日志

操作日志记录已经提交的管理员和角色变更，入口为 `/system/audit-logs`。页面支持操作人/对象关键词、操作类型、操作日期、Trace ID、分页，以及点击操作人继续筛选。详情显示操作时的身份、对象与修改前后值；角色删除后仍保留历史名称与权限。

## 启用与权限

已有环境先执行 `make migrate` 应用 `00003_operation_audit.sql`，再执行 `make sync-resources` 同步 `system.audit.read`，随后更新 API、RPC 和前端。完整容器部署对应 `make compose-migrate`、`make compose-sync-resources`，需使用包含新迁移的镜像。迁移不会补写历史日志。

超级管理员自动拥有新目录权限；普通角色需要显式授予“操作日志 → 查看操作日志”。该权限允许查看整个 system 的操作历史，不受其管理员/角色管理范围限制，适用于负责审计的人员。仅能管理用户或角色不等于可以查看日志。列表和详情都在 RPC 中核验当前会话及权限，撤销后新请求立即拒绝。

## 记录范围

| 操作 | 记录内容 |
| --- | --- |
| 新增、编辑、启停管理员 | 用户名、显示名称、状态 |
| 分配管理员角色 | 角色 ID、标识与当时名称 |
| 新增、编辑、启停、删除角色 | 标识、名称、说明、状态；删除时保留权限 |
| 设置角色权限 | 修改前后的权限键 |
| 修改个人密码 | 操作事实和撤销全部会话，不记录密码与摘要 |
| 领域命令禁用账号 | 来源为 `command`，操作人 `system-cli`，记录账号状态变化 |

API 操作人取自 RPC 验证后的会话，客户端身份头无效。运维命令不伪装成某个管理员，操作人 ID 为 `0`；无请求上下文时 Trace ID 为空。名称是操作时的快照，不随之后改名而变化。

这批记录成功提交的业务修改。校验失败、越权、冲突与事务回滚不会生成“成功”日志；登录/刷新/退出、失败尝试、初始化与权限目录同步不在这批业务审计范围内。访问日志仍由服务日志承担。重复提交且可记录字段未改变时，详情明确显示“本次提交未改变可记录字段”。

## 事务与数据边界

业务写入、审计记录与权限版本更新使用同一个 system 数据库事务。审计插入失败则整次修改回滚并返回 503；不能先返回成功再异步写日志。密码变更失败时也不提交新密码和会话撤销。没有在该流程引入消息队列。

审计使用字段白名单构造快照，不序列化 HTTP/RPC 请求或完整 GORM 实体，不存储密码、密码摘要、Cookie、session ID 或 token。列表只返回摘要，详情独立读取变更字段。数据保存在 `audit_logs`，按时间/ID、操作人、操作类型和 Trace ID 建索引。历史引用不建级联外键，删除角色不会删除日志。

应用只提供追加和只读查询，没有修改、删除、清空或导出日志接口。数据库运维权限仍能修改数据；当前不宣称具备防篡改存证能力。日志保留/归档策略、登录失败安全事件和跨领域汇总留待后续需求明确后接入。

## HTTP 与 RPC

| HTTP | RPC | 权限 |
| --- | --- | --- |
| `GET /admin-api/v1/system/audit-logs` | `ListAuditLogs` | `system.audit.read` |
| `GET /admin-api/v1/system/audit-logs/{id}` | `GetAuditLog` | `system.audit.read` |

列表参数：`page` 默认 1、范围 1–100000；`pageSize` 默认 20、范围 1–100；`keyword` 最多 64 字、按字面包含匹配操作人和对象；`action` 为返回的 `actions` 中的键；`actorId` 为十进制字符串（`0` 为运维命令）；`traceId` 为完整的 32 位小写十六进制标识。

`from`、`to` 为含时区的 RFC3339 时间，分别包含起点、排除终点。前端按浏览器本地日期把结束日转换为次日零点，因此选定日期的整天均包含在内。查询按 `created_at DESC, id DESC` 排序；列表是实时分页，新增记录可能移动已有页的位置。详情 ID 为字符串，不经过 JavaScript 浮点数转换。

响应遵循统一 envelope：`code`、`message`、`data`、`traceId`。列表 `data` 含 `items`、`actions`、`total`、`page`、`pageSize`；详情含 `record` 和 `changes`，每条变更为 `{field,label,before:[],after:[]}`。不存在的记录返回 404，无权限返回 403，无效筛选返回 400，依赖失败返回 503。完整类型见 [OpenAPI](openapi/admin.openapi.json)。

```bash
# TOKEN 为当前管理员 access token，仅用于示意，不写入脚本或日志。
curl -H "Authorization: Bearer $TOKEN" \
  'http://127.0.0.1:8081/admin-api/v1/system/audit-logs?page=1&pageSize=20&action=system.role.grant'
```

## 验证入口

`make test-rbac` 包含独立临时 MySQL 库的日志字段、历史快照、查询边界、权限撤销、写失败回滚和密码脱敏测试（含 race）。`make rbac-smoke` 验证真实 API → RPC 的权限、分页、错误码及 Trace ID 关联。`make test-auth` 回归受日志事务影响的改密和禁用行为。`pnpm --dir web/admin test` 验证慢响应、详情切换/关闭、权限撤销后不恢复旧数据。测试仅清理自身创建的临时库或随机命名记录。
