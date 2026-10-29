# System RPC

package `goseed.system.v1`，service `System`。契约：[system.proto](system.proto)。

| 方法 | 请求 | 成功响应 | 失败 |
| --- | --- | --- | --- |
| `Ping` | `{}` | `{service, version, time}` | RPC 传输/超时错误 |
| `CheckReady` | `{}` | `{}` | MySQL/Redis 不可用：Unavailable |

`time` 是 UTC RFC3339。方法只提供无敏感数据探测；第一阶段未接入服务身份校验，因此 RPC 只绑定本地或 Compose 内部网络。dev/test 模式开启 reflection，生产接入身份验证时同步调整配置。

可选安装 grpcurl 后，从已运行的本地服务查询：

```bash
grpcurl -plaintext -d '{}' 127.0.0.1:19001 goseed.system.v1.System/Ping
grpcurl -plaintext -d '{}' 127.0.0.1:19001 goseed.system.v1.System/CheckReady
```

不安装 grpcurl 也可以使用项目 `bin/probe -rpc 127.0.0.1:19001 -ready`。标准 gRPC Health 只表示进程服务状态，CheckReady 才代表下游就绪。API 将 CheckReady 的失败统一转换成 HTTP 503。
