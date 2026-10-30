# Member RPC

公开契约为 [member.proto](member.proto)。使用 `make gen-rpc DOMAIN=member` 生成 pb、client、server 和初始骨架；业务逻辑保留在领域 application，调用方只导入 pb/client。

| 方法 | 调用服务 | 用户身份 / 权限 |
| --- | --- | --- |
| Register / Login | app-api | 无会员会话；账号校验与 Redis 限流 |
| Refresh / Logout | app-api | 验证 refresh token 摘要与会话 |
| Current / UpdateProfile / ChangePassword | app-api | 验证会员 JWT 与数据库会话 |
| ListUsers / GetUser | admin-api | system CheckAccess：`member.user.read` |
| SetUserStatus | admin-api | system CheckAccess：`member.user.update_status` |
| CheckReady | 健康探针 | 检查 MySQL、Redis 和 system-rpc |

服务凭证通过 `pkg/rpcx` 发送，用户 JWT 通过经过入口校验的上下文传播。member-rpc 在权限检查中自己选择权限键，不使用 HTTP 请求传来的权限或用户 ID。system 仅允许 `member-rpc` 查询 `member.` 权限，不允许查询 system 权限或调用管理员管理方法。

HTTP 请求示例在[会员模块说明](../../../docs/member-module.md)。对应的 RPC 请求为 `RegisterRequest {username, password, nickname}`，响应 `Profile {id, username, nickname, revision, created_at}`。`SetUserStatusRequest {id, enabled, revision}` 返回 Empty；actor ID 由 system 验证后返回，不由调用方提供。

错误沿用统一 gRPC status：参数错误 InvalidArgument；会话无效 Unauthenticated；服务或资源无权限 PermissionDenied；不存在 NotFound；用户名重复 AlreadyExists；revision 不匹配 Aborted；限流 ResourceExhausted；存储或授权服务故障 Unavailable。HTTP 映射见[开发约定](../../../docs/development-guide.md)。

`Register` 不自动登录，网络中断后重复注册可能得到用户名冲突；使用原凭证登录可确认是否已创建。更新、启停以 revision 防止旧请求覆盖新状态。数据库事务不跨 RPC：先完成管理员授权，再开始 member 本地事务；授权后已进入执行的请求可能在撤权之后完成。

配置与启动见[领域 README](../README.md)。真实服务由 `make member-smoke` 验证，内存与数据库测试由 `make test`、`make test-member` 验证。RPC 不对公网暴露；生产服务链路应配置 TLS 和每服务凭证。
