# System RPC

package `goseed.system.v1`，service `System`。契约：[system.proto](system.proto)。

| 方法 | 请求 | 成功响应 | 失败 |
| --- | --- | --- | --- |
| `Ping` | `{}` | `{service, version, time}` | RPC 传输/超时错误 |
| `CheckReady` | `{}` | `{}` | MySQL/Redis 不可用：Unavailable |

`time` 是 UTC RFC3339。上面两个方法只提供无敏感数据探测；认证方法执行逐方法服务身份校验。RPC 只绑定本地或 Compose 内部网络；dev/test 模式开启 reflection，生产配置关闭 reflection 并启用受验证的 TLS 传输。

可选安装 grpcurl 后，从已运行的本地服务查询：

```bash
grpcurl -plaintext -d '{}' 127.0.0.1:19001 goseed.system.v1.System/Ping
grpcurl -plaintext -d '{}' 127.0.0.1:19001 goseed.system.v1.System/CheckReady
```

不安装 grpcurl 也可以使用项目 `bin/probe -rpc 127.0.0.1:19001 -ready`。标准 gRPC Health 只表示进程服务状态，CheckReady 才代表下游就绪。API 将 CheckReady 的失败统一转换成 HTTP 503。

## 管理员认证用例

`AdminLogin`、`AdminRefresh`、`AdminLogout`、`AdminCurrent` 和 `AdminChangePassword` 仅允许已配置的 `admin-api` 服务身份调用。`AdminCurrent`、`AdminChangePassword` 还校验原始管理员 token 和数据库当前会话状态；调用方传入的用户 ID 或角色不能替代校验。

密码和 refresh token 字段只在认证链路内部使用，不记录 RPC 请求正文。原始 refresh token 由 API 转为 HttpOnly Cookie。公共 Ping/CheckReady 不返回认证或业务数据，仍允许探测客户端访问。

## 管理用例

`ListAuditLogs`、`GetAuditLog` 仅允许 admin-api 服务身份调用，management 用例检查当前管理员的 `system.audit.read`。操作日志在业务事务内由 system 构造，不对外提供写日志 RPC。记录范围、字段和查询语义见[操作日志](../../../docs/operation-audit.md)。

`ListAdministrators`、`CreateAdministrator`、`UpdateAdministrator`、`AssignAdministratorRoles`、`ListRoles`、`CreateRole`、`UpdateRole`、`GrantRole`、`DeleteRole` 和 `ListPermissions` 同样只允许 admin-api 服务身份调用，再由 management 用例校验原始用户 token、当前会话、固定的权限键和目标授权范围。无权限为 PermissionDenied，编辑版本过期为 Aborted，已分配成员或最后超级管理员冲突为 FailedPrecondition。完整接口、权限与同步规则见 [角色授权](../../../docs/authorization.md)。

文件元数据与任务方法定义在 `system.proto`：ListFiles/CreateFile/AuthorizeFileUpload/CompleteFileUpload/DownloadFile/DeleteFile、ListTasks/ExportDictionaries/RetryTask。所有方法仅允许 admin-api 服务调用，并在 application 层校验原用户身份；FileTransferResponse 的对象键仅供内部流式适配，禁止下发浏览器。行为与配置见[文件与任务](../../../docs/files-and-tasks.md)。
