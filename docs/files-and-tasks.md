# 文件、队列与定时任务

GoSeed 使用私有文件存储和 Asynq 0.26.0 队列。admin-api 流式收发内容；system-rpc 管理元数据、归属、权限和任务。system-worker 消费任务并投递 outbox；system-scheduler 独立投递小时调度，不连接数据库。三个领域进程共享 application / repository，不导入彼此的 logic。

## 文件上传与下载

1. `POST /admin-api/v1/system/files` 提交 `{ "name": "notes.txt", "size": 5, "contentType": "text/plain" }`，返回 201 与 `data.record.id`。
2. `PUT /admin-api/v1/system/files/{id}/content` 以 `application/octet-stream` 上传原始字节。请求必须带 Bearer 和 `X-GoSeed-CSRF: 1`。
3. API 通过 RPC 校验上传人的当前会话、权限、文件归属及有效期，随后写入不可覆盖的存储对象。RPC 从存储重新读取大小、实际类型和 SHA-256，原子提交可用状态与上传审计。
4. `GET /admin-api/v1/system/files/{id}/content` 校验当前登录和归属后返回 attachment；成功响应是二进制，不套 JSON envelope。内容强制 `application/octet-stream`、`nosniff` 与 `private, no-store`。
5. `DELETE /admin-api/v1/system/files/{id}` 返回 202。在同一事务标记不可下载、创建删除任务、写 outbox 与审计。物理对象由 worker 幂等删除。

文件单个上限 **10 MiB**，不允许空文件。支持 PNG、JPEG、PDF、纯文本、CSV、JSON、ZIP；JSON 还验证语法。文件名和审计目标名称均最长 180 字符，不接受路径分隔符或控制字符。声明类型和实际内容不匹配时拒绝完成上传；这不是恶意文件扫描服务。

上传记录有效期一小时。失败后原页面的“重试上传”复用文件 ID；若对象已成功写入但完成回执丢失，会验证并完成已有对象，不替换内容。刷新页面后重新选择文件会创建新记录。超过 24 小时的未完成上传、失败导出的残留文件及临时上传文件由定时清理回收，每次最多处理 200 个元数据对象。

所有文件按管理员本人隔离，超级管理员也不能下载其他管理员的文件。`system.file.read/upload/delete` 是独立权限；上传和删除依赖读取权限。导出文件下载还要求当前 `system.task.export` 权限。对象键只在内部 RPC 中流转，不返回前端；不把存储目录作为静态站点公开。

后台文件列表提供预览抽屉：PNG/JPEG 支持查看、缩放和旋转，PDF 使用固定版本 [PDF.js](https://mozilla.github.io/pdf.js/) 6.4.299 在页面内渲染，支持翻页和缩放，TXT/CSV/JSON 以纯文本展示（最多前 256 KiB，超出时提示下载完整文件）。ZIP 显示下载提示；加密或无法解析的 PDF 提示下载后查看。预览通过同一个带 Bearer 的内容接口读取，不生成公开地址；关闭抽屉、离开页面或权限失效时取消请求、清空内容并释放 Blob URL。文本按文字转义显示，不执行 HTML 或脚本。PDF 仅绘制页面，不激活文档脚本、表单或外部链接；渲染器按需加载，worker、字体和字符映射资源由本站提供，不使用第三方在线预览服务。`pnpm dev/build` 自动从锁定依赖准备被 Git 忽略的 `public/pdfjs/` 资源，关闭预览会销毁文档与 worker。

原始文件名保留在元数据中，用于列表、预览标题和下载文件名；实际存储键按 **`YYYYMMDD/随机文件名`** 生成，例如 `20261030/7c8f91b2d3e44a56b7c8d9e0f1a2b3c4`。日期由记录创建时间转换到统一 `TZ` 配置的时区后生成，默认 `Asia/Shanghai`；导出文件名中的日期和时间也使用该时区。数据库时间和接口时间仍为 UTC；文件名由服务端 `crypto/rand` 生成 16 字节随机数并编码为 32 位十六进制字符串，不带原始名称或扩展名。上传与任务导出都使用该规则，本地磁盘自动创建单层日期目录，S3 使用同样的对象前缀。上传同名文件会产生独立记录和对象，不覆盖已有文件。

新环境和已有环境都须执行迁移 `00008_file_date_paths.sql`，将 `files.object_key` 扩展至 `VARCHAR(64)`。已有的 32 位平铺键继续支持上传重试、读取和删除，时区改变后也按已保存的键读取，无需移动旧文件；不自动回滚缩短字段，以免截断日期路径。

## 存储配置

`STORAGE_DRIVER=local` 使用 `STORAGE_ROOT`，本地默认 `.data/files`；Compose 使用命名卷 `files-data`，挂载到 admin-api、system-rpc 和 system-worker 的同一路径。不同主机上的独立本地目录不能互通；多主机部署使用共享文件系统或 S3。

`STORAGE_DRIVER=s3` 使用 S3 兼容协议，MinIO Go SDK 固定为 7.0.95。配置 `S3_ENDPOINT`（host:port，不带协议）、`S3_BUCKET`、`S3_ACCESS_KEY`、`S3_SECRET_KEY`、`S3_SECURE` 和 `S3_REGION`。预先创建私有 bucket，给进程对象读、写、删除与 bucket 检查权限；凭证仅放部署环境。三个进程必须指向同一个 bucket。

本地写入先落临时文件、同步数据，再原子链接到随机对象键；S3 使用单次 PutObject 和 `If-None-Match: *`，兼容服务必须支持条件写。已有对象返回冲突，不能覆盖。传输采用受限磁盘临时文件，避免把文件正文放进 JSON/protobuf，也不把完整内容作为日志。

## 异步任务

当前真实任务：

| 类型 | 用途 |
| --- | --- |
| `system.dictionary_export.v1` | 导出字典与字典项为 UTF-8 BOM CSV，结果进入自己的文件列表 |
| `system.file_delete.v1` | 可靠删除对象与更新元数据状态 |
| `system.upload_cleanup.v1` | 清理过期上传、失败导出残留及待删除对象 |

任务中心展示本人任务及系统清理任务，包含状态、次数、安全错误信息、时间和 Trace ID。详情为打开时的快照；列表每 5 秒刷新。系统任务仅超级管理员可以人工重试。

`POST /system/tasks/dictionary-export` 接受客户端生成的 UUID v4 `requestKey`。同一管理员和 key 只产生一个任务，返回 202；网络异常可用原 key 重试。导出要求 `system.task.export`、字典读取、文件读取及任务读取权限；创建新导出使用新 key。后台页面会保留失败请求的 key 以防重复创建。

导出执行前及提交结果前重新查询账号状态和权限，不保存长期用户 token。退出会话不取消已授权任务；禁用账号或撤销导出相关权限会阻止结果提交和后续下载。一次最多 1000 个字典、10000 个字典项和 10 MiB；超限失败，不静默截断。CSV 单元格防范公式前缀。崩溃后已有不可变结果被复用，重试不是重新生成当前数据快照。

### 可靠投递与重试

任务、outbox 和发起审计在同一个 MySQL 事务提交。RPC 不需要在请求中访问队列，因此 Redis 短暂故障不会丢失已受理的任务。

worker 每秒扫描到期 outbox。投递成功仅记录时间，业务完成前保留 outbox，每 30 秒检查并重新投递，能够恢复队列丢失的未完成任务。消息仅含任务 ID 和 generation；Trace ID、发起人、结果归属均来自数据库。

数据库行锁领取任务，持有两分钟租约；每次执行上限 50 秒，Asynq timeout 为一分钟。数据库记录最多五次业务尝试，指数退避；权限失效、无效参数和超限直接失败。数据库故障导致无法记录状态时由 Asynq 基础重试兜底，租约到期后 outbox 恢复中断任务。最终提交校验 generation 和 attempt，旧消息和旧 worker 不能覆盖新批次状态。对象不可覆盖、事务状态转换与删除幂等共同保证重复消息无重复业务结果；Asynq TaskID 只是减少重复调度。

`POST /system/tasks/{id}/retry` 需要 `system.task.retry`，仅允许失败任务，返回 202。在事务中重置尝试次数、增加 generation、重建 outbox 并写审计；导出仍需原权限。失败导出文件已被清理时会分配新对象键。

## 定时调度与生命周期

scheduler 每 10 秒尝试投递当前 UTC 小时的清理触发消息；同一小时的 Asynq TaskID 固定、结果保留 48 小时。worker 接收后以“调度类型 + 小时”建立数据库唯一任务和 outbox，再确认消息。重复启动和重复触发不会重复创建业务任务。

部署默认运行一份 scheduler，更新时先停旧实例。调度与业务均有去重保护；短暂队列故障恢复后补当前小时，长时间停机不会回放全部历史小时。任务清理以当前数据为准，无需逐小时补跑。本批提供代码声明的清理计划，未提供后台编辑任意 cron 或执行命令的能力。

`QUEUE_HOST/PORT/PASSWORD/DB/NAME` 配置 Asynq，默认开发 Redis DB 1、队列 `goseed_dev_system`。其他环境使用不同名称与凭证；生产队列使用独立 Redis、持久化和 `noeviction`。`make dev` 启动五个 Go 进程和 Vite；`make compose-up` 启动包含 worker/scheduler 的完整环境。worker 与 scheduler 的健康端口默认 19201 / 19202，Compose 只发布到回环地址。停止时先退出投递循环、关闭接收与在途任务，再关闭连接；超时工作由持久化租约恢复。

新增环境先执行 `make init`，再迁移、同步权限目录。新增权限不会自动授予普通角色。迁移文件为 `00005_files_tasks.sql`，不在服务启动时自动执行。

## 扩展与验证

新增任务类型需定义版本化 payload、领域 application 用例、数据库幂等依据及 outbox，在 worker handler 显式注册。不要把 shell 命令、用户 token 或完整业务文件塞入 payload。

- `make test`：存储原子性/大小边界/S3 条件请求、内容验证、CSV 公式处理及现有回归。
- `make test-storage`：本地存储并发与 S3 条件请求协议测试。配置独立测试服务的 `GOSEED_S3_TEST_ENDPOINT/ACCESS/SECRET` 后额外运行真实 S3 上传/读取/冲突/删除，临时 bucket 以随机名称创建并清理。未配置时明确跳过实服务测试，不操作生产 bucket。
- `make test-rbac`：独立临时 MySQL 库；文件归属、事务审计回滚、重复提交/消费、租约恢复、撤权、人工重试、定时幂等与真实 Asynq 消费。队列使用随机名称，测试结束仅删除该队列。
- `make rbac-smoke`：运行中的上传、私有下载、HTTP 状态、导出、worker 处理、删除及现有管理链路。

官方依据：[Asynq](https://github.com/hibiken/asynq)、[Asynq API](https://pkg.go.dev/github.com/hibiken/asynq@v0.26.0)、[S3 条件写](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html)。

## 故障恢复约定

outbox 在入队前以数据库 CAS 分配递增 delivery 序号和 30 秒投递租约；TaskID 为 `job-<id>-<generation>-delivery-<delivery>`。Redis 结果未知或旧消息已归档时，租约到期使用新 ID；数据库内业务领取、generation 和租约继续保证幂等。投递次数不消耗业务 attempts。

清理某个对象失败时将该文件推迟一小时，继续处理本批其余对象；后续批次按 cleanup_after 过滤，避免失败对象长期占满前 200 条。未来小时的调度消息正常重试，不当作坏消息立即归档。

下载路由单独关闭 go-zero 响应缓冲，按 32 KiB 流式输出；存储上下文和 HTTP 写期限均为 60 秒。上传仍有原超时保护。Nginx 关闭代理响应缓冲。管理 API 在最外层收尾上传请求体，拒绝上传后最多丢弃 10 MiB + 1 字节、最多等待 5 秒，避免未读请求体触发 TCP reset，导致网关把 401/403 替换为 502。

worker 的 Asynq ShutdownTimeout 为 8 秒，go-zero WaitTime 为 20 秒，Compose stop_grace_period 为 30 秒。健康端口监听失败以非零状态退出；服务配置自动重启。worker 不持有 JWT 签名密钥，只按数据库当前身份和权限执行持久化任务。

S3 上传缓存在 `STORAGE_ROOT/.spool`，admin-api 启动及每小时清理自身目录内超过 24 小时的残留；worker 同样清理自身存储目录。进程分布在不同主机时不依赖 worker 访问 API 的临时目录。
