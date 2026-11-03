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

API 在创建 span 前丢弃外部的 traceparent、tracestate、baggage，从框架记录的请求目标中去掉查询串。业务处理器仍可读取完整查询参数。两个 API 在 go-zero 创建 HTTP span 时隔离 User-Agent 和 X-Forwarded-For，再将原始头交给业务中间件，可信代理校验与限流仍使用原来的地址链。Alloy 在采样前也删除客户端地址、网络对端地址和 User-Agent 属性，覆盖滚动升级期间的旧生产者。已写入 Tempo 的旧数据按原有保留期过期。访问日志不存原始路径、请求或响应正文、头、Cookie、密码、搜索词、客户端 IP 或用户 ID。文件流只统计已写出的字节。探活路径不写 API 访问日志。

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

## 本地集中查询

```bash
make init                         # 生成或补齐 .env 中的 Grafana 登录凭据
make observability-up
make observability-smoke
# 额外验证中断和恢复。执行期间 system-rpc 和 Alloy 会分别短暂停止。
python3 scripts/observability-smoke.py --fault
```

默认 Compose 环境不启动观测组件。启用 profile 时，两个 API 和两个 RPC 通过内部网络向 Alloy 的 OTLP gRPC 端口上报。服务端完整采样，Alloy 先过滤两个领域的 CheckReady 和 gRPC 标准健康检查 `grpc.health.v1.Health/Check`，再保留含 HTTP 5xx 或 gRPC 服务故障（Unknown、DeadlineExceeded、Internal、Unavailable、DataLoss）的链路。耗时至少 1 秒的链路也会保留。快速的鉴权失败、权限拒绝和限流只按 `TRACE_BASELINE_PERCENT` 采样，默认 100%，可设为 0–100。若同一链路还包含真正的服务故障，仍可查到整条链路。worker、scheduler 只收集容器日志。首版链路覆盖 HTTP 和 RPC，数据库与异步任务未创建单独的 span。

| 本机入口 | 默认地址 |
| --- | --- |
| Grafana Explore | http://127.0.0.1:13000/explore |
| Loki HTTP / Tempo HTTP | http://127.0.0.1:13100 / http://127.0.0.1:13200 |
| 宿主机 Go 进程使用的 OTLP gRPC | 127.0.0.1:14317 |

打开 Grafana Explore 后，用本地 `.env` 中的 `GRAFANA_ADMIN_USER` 和 `GRAFANA_ADMIN_PASSWORD` 登录。默认账号为 `admin`，`make init` 为 Grafana 单独生成随机密码，重复执行会保留已有值。

Grafana 只在首次创建数据卷里的管理员时读取这两个值。已有数据卷沿用原账号。在旧匿名环境升级或需要将原管理员的密码更新为 `.env` 中的值时，先执行上面的 `make init` 和 `make observability-up`，再运行以下命令。默认管理员 ID 为 1；若改过用户名，请将 `GRAFANA_ADMIN_USER` 设为已有用户名。

```bash
docker compose --env-file .env -p goseed -f deploy/compose/compose.yaml exec -T grafana \
  sh -ec 'printf "%s\n" "$GF_SECURITY_ADMIN_PASSWORD" | grafana cli admin reset-admin-password --password-from-stdin'
```

日志采集限定 Compose 项目 `goseed`，包括网关、应用与 worker。查询标签为 `compose_project`、`compose_service`。trace ID 保留在日志行里，高基数 ID 不作为索引标签。Loki 和 Tempo 在本地命名卷保留数据 168 小时。停止时先停止 Alloy 和 Grafana。Tempo 收到退出信号后预留 35 秒让查询队列完成清理，再将内存中的链路写盘。Compose 为整个退出流程预留 90 秒。

1. 在 Grafana Explore 选择 Loki，用下列 LogQL 查响应中的 trace ID。展开 API 的日志行，在关联字段 Trace 旁点 **View Trace**。
2. Tempo 中展开 API 及其 RPC span，比较各段时间与状态。在 span 详情点击 **Related logs**，跳回同一时间窗口和 trace ID 的 Loki 记录。尚在批处理或等待采样决定的链路需要稍后查询。

```logql
{compose_project="goseed"} |= "替换为32位traceId"
{compose_service="admin-api"} | json | content="http access" | status >= 500
```

停止观测组件、恢复默认应用配置：

```bash
make observability-down
make compose-up
make smoke
```

`observability-down` 保留数据卷，只停止四个观测容器。随后 `make compose-up` 在默认的空 `COMPOSE_OTEL_ENDPOINT` 下重建应用。`make compose-down` 则移除所有 profile 的容器和项目网络，保留命名卷。运行中的观测组件使用 `unless-stopped`，与应用一同随 Docker 恢复；手动停止的组件保持停止。

## 宿主机开发

`OTEL_ENDPOINT` 配置宿主机 Go 进程，优先级为 shell、项目 .env、.env.example；显式传空值可关闭本次上报。`COMPOSE_OTEL_ENDPOINT` 配置容器中的 Go 进程。之前在 .env 用 OTEL_ENDPOINT 配置外部容器采集器的项目，请将该容器地址移到 COMPOSE_OTEL_ENDPOINT。`make observability-up` 会为本次启动指定容器网络地址 `alloy:4317`。

宿主机跑 API / RPC 时，先释放完整 Compose 栈占用的应用端口。以下命令保留数据，只在 Docker 中启动基础设施和观测组件：

```bash
make compose-down
make infra-up
docker compose --env-file .env -p goseed -f deploy/compose/compose.yaml --profile observability up -d alloy grafana
OTEL_ENDPOINT=127.0.0.1:14317 make dev
```

`make test-observability` 在独立的临时 Compose 项目中以 0% 基础采样验证拒绝、故障、慢链路和隐私字段，同时检查全栈关闭后网络被移除、命名卷仍在。`observability-smoke` 的正常请求关联检查使用默认的 100% 基础采样。

## 生产配置

- 本地 Grafana 只绑定回环地址，查询需要登录。生产环境按角色分配查询权限并启用 TLS。Loki 和 Tempo 的本机直连端口独立于 Grafana 登录，对外部署时限制为内部网络访问。
- Alloy 挂载 Docker socket 的权限接近宿主机 root，只读挂载仍可调用 Docker API。生产使用限制 API 的 socket proxy，或读取容器日志文件。
- go-zero v1.10.2 的 `otlpgrpc` 使用 `WithInsecure()`。当前 gRPC 上报只走同机受控网络；跨主机上报改用 `otlphttp`、启用 `OtlpHttpSecure` 并验证证书。
- 生产按流量和容量调低 Alloy 的普通链路采样率，保留服务故障和慢调用策略。保持应用完整上报才能在链路结束后做决定。调整采样等待时间和内存容量时，应覆盖实际最长的请求与导出延迟。
