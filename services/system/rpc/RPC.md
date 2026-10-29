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
