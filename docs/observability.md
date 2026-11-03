# API 访问日志与链路排障

两个 API 在 HTTP 路由入口为每个业务请求写一条 JSON 日志，消息为 `http access`。业务代码、RPC 和审计记录使用同一条链路的 ID。网关使用自己的 `requestId`，可通过上游响应的 `traceparent` 找到对应 API 请求。

| API 字段 | 含义 |
| --- | --- |
| service、method、route | 服务、标准方法、已注册路由模板。未匹配记为 unmatched，未知方法记为 OTHER |
| status、duration、bytes | 客户端响应状态、总耗时、实际写出的正文长度。5xx 使用 error 级别 |
| requestBytes | 请求声明的正 Content-Length；未声明时省略 |
| businessCode | 统一错误处理器写入的业务码。框架直接拒绝的请求可能没有 |
| trace | 来自服务端 Traceparent 响应头；404、405 等未进入业务路由的请求可能没有 |
| query | 仅 page、pageSize、enabled、status，值限 1–64 个 ASCII 字母、数字或 `_.:-` |

API 在创建 span 前丢弃外部的 traceparent、tracestate、baggage，从框架记录的请求目标中去掉查询串。业务处理器仍可读取完整查询参数。访问日志不存原始路径、请求或响应正文、头、Cookie、密码、搜索词、客户端 IP 或用户 ID。文件流只统计已写出的字节。探活路径不写 API 访问日志。

网关访问日志记录规范化 URI（不含查询串）、远端地址、HTTP 和上游状态、响应大小、网关总耗时与上游耗时；不记录 Referer。Compose 容器日志按每文件 20 MB、最多 5 个文件轮转。

## 从一次失败请求定位原因

1. 从客户端错误响应复制 32 位 `traceId`，在本地执行以下命令。可先限定时间以减少输出。

   ```bash
   trace_id='替换为响应中的链路ID'
   docker compose --env-file .env -p goseed -f deploy/compose/compose.yaml logs --since=15m gateway admin-api app-api system-rpc member-rpc | rg -F "$trace_id"
   ```

2. 在 `http access` 行确认路由、状态、业务码与耗时，在网关行核对上游状态。RPC 返回内部故障时，`rpc failed` 给出 `rpcMethod` 与 `grpcCode`；预期的鉴权拒绝和取消不产生 RPC 故障日志。RPC 完全不可达时，下游没有接收到请求，应看客户端这一侧的失败记录。
3. 例如修复前搜索“布尔”会收到 503（100503）。对于同类数据库故障，沿同一 trace 查找 RPC 的存储错误行，`mysqlErrno=1270` 和 `sqlState=HY000` 可以定位到字符集比较，而不需要保存搜索词。错误行只写错误类型与 MySQL 编号、SQLState，不写可能含业务数据的驱动错误文本。

管理员写操作的修改前后值在受权限控制的[操作审计](operation-audit.md)中查询。该审计与容器访问日志分别存储。
