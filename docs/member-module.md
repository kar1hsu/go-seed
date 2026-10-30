# C 端会员业务样板

Member 是独立业务领域。后台账号用于管理，会员账号用于 C 端，两者即使用户名或数字 ID 相同也不代表同一个主体。首版采用账号密码，提供可运行的业务闭环及可参考的目录，不依赖短信或邮件服务。

## 已实现的业务

| 入口 | 行为 |
| --- | --- |
| `POST /api/v1/auth/register` | 创建会员，返回 201；账号唯一，注册与审计原子提交 |
| `POST /api/v1/auth/login` | 返回短期 JWT，设置会员 HttpOnly refresh Cookie |
| `POST /api/v1/auth/refresh` | 轮换 refresh token，重用旧 token 撤销整个会话 |
| `POST /api/v1/auth/logout` | 撤销当前会话并清除 Cookie |
| `GET /api/v1/me` | 查询当前会员公开资料 |
| `PATCH /api/v1/me` | 更新当前会员昵称，校验 revision |
| `POST /api/v1/me/password` | 校验原密码、更新密码、撤销全部会话 |
| `GET /admin-api/v1/member/users` | 按账号、昵称、启停状态分页查询 |
| `GET /admin-api/v1/member/users/:id` | 查看会员详情及最近 20 条变更事件 |
| `PUT /admin-api/v1/member/users/:id/status` | 按 revision 启停；停用后旧会话失效，重新启用需重新登录 |

后台菜单为“业务管理 → 会员管理”。普通角色需要显式授予 `member.user.read`；启停还需要 `member.user.update_status`。目录同步不自动授权普通角色。

账号为 3–32 位小写 ASCII，以字母开头，允许数字和 `_.-`，创建后不可修改。昵称为 1–64 个 Unicode 字符，拒绝纯空格和控制字符。密码为 12–72 字节，使用 bcrypt cost 12；不能包含 NUL。JSON body 上限 16 KiB。列表 pageSize 上限 100。

公开资料只有 id、username、nickname、revision、createdAt，不包含密码哈希、token version、启停信息或管理员权限。没有公开会员列表；个人资料接口不接受目标用户 ID。后台不能读取或设置会员密码。

## 身份与会话

- 会员 access JWT：HS256，默认 15 分钟，`kind=member`、issuer `go-seed-member`、audience `go-seed-app-api`，使用独立 `MEMBER_TOKEN_SECRET`。
- 会员 refresh token：随机值，数据库只保存 SHA-256 摘要；会话默认 7 天绝对有效期，刷新不延长整个会话寿命。
- Cookie：`goseed_member_refresh`，Host-only、HttpOnly、SameSite=Strict、Path `/api/v1/auth`；生产启用 `MEMBER_COOKIE_SECURE=true` 并使用 HTTPS。
- 所有写请求要求 `X-GoSeed-CSRF: 1`。浏览器来源必须在 `MEMBER_ALLOWED_ORIGINS`；默认不启用跨域 CORS。C 端页面使用同源代理。CLI 仍需 CSRF 请求头。
- JWT 仅作入口校验。member-rpc 从主库确认账号启用、token version、session 未撤销且未过期；退出、改密和停用后，新请求不能沿用旧 JWT。
- refresh token 只能消费一次；并行刷新可能触发重放保护，客户端必须串行刷新。刷新遇到结果未知的网络错误时重新登录，避免盲目重试旧 token。

会员不使用管理员的 Cookie、密钥、权限或会话表。会员 JWT 无法用于后台；管理员 JWT 无法作为会员访问 `/me`。

## 数据、权限与事务

```text
C 端 HTTP → app-api → member-rpc → account → repository → goseed_member
后台 HTTP → admin-api → member-rpc → system-rpc.CheckAccess
                                 → management → repository → goseed_member
```

MySQL 实例可以共用，逻辑数据库与账号分开。member 数据库包含 users、member_sessions、member_refresh_tokens、audit_events。没有跨库外键或跨域 Join。会员账号的权限是“当前登录主体自己的资料”；首版没有租户、会员角色或面向 C 端的 Casbin 权限模型。

后台用例固定选择权限键，并通过 system 验证原始管理员 JWT、当前会话、账号状态与当前权限。member-rpc 的服务凭证只能调用 system CheckAccess，权限范围固定为 `member.`。授权检查失败时拒绝业务操作，不缓存最终允许结果。授权先于 member 本地事务，已通过授权的在途操作可能完成，不宣称跨数据库原子撤权。

所有会话变更、改密、资料更新与后台启停先锁同一会员行，防止改密/禁用和刷新相互覆盖。个人资料与后台操作带 revision，过期提交返回 409。最近登录时间独立更新，不引发资料编辑冲突。

注册、资料变更、改密和启停的事件与业务写入在本地同一事务提交，审计写入失败会回滚业务。事件记录动作、actor kind/ID、目标会员、时间和 Trace ID，不记录密码、token、昵称正文。事件目前只在会员详情查询，尚未异步汇总到 system 操作日志。

## 本地调用

先执行根 README 的准备流程，启动 `make dev` 或完整 Compose。以下为本地测试账号；命令在同一 shell 执行，jq 用于提取响应。Cookie 文件含登录凭证，使用后删除。

```bash
umask 077
member_url=http://127.0.0.1:8088
curl -sS -H 'Content-Type: application/json' -H 'X-GoSeed-CSRF: 1' \
  -d '{"username":"demo_member","password":"replace-with-long-password","nickname":"示例会员"}' \
  "$member_url/api/v1/auth/register"
member_login=$(curl -sS -c /tmp/goseed-member-cookies -H 'Content-Type: application/json' \
  -H 'X-GoSeed-CSRF: 1' -d '{"username":"demo_member","password":"replace-with-long-password"}' \
  "$member_url/api/v1/auth/login")
member_access=$(printf '%s' "$member_login" | jq -r '.data.accessToken')
curl -sS -H "Authorization: Bearer $member_access" "$member_url/api/v1/me"
curl -sS -b /tmp/goseed-member-cookies -c /tmp/goseed-member-cookies \
  -H 'X-GoSeed-CSRF: 1' -X POST "$member_url/api/v1/auth/refresh"
curl -sS -b /tmp/goseed-member-cookies -H 'X-GoSeed-CSRF: 1' \
  -X POST "$member_url/api/v1/auth/logout"
rm /tmp/goseed-member-cookies
unset member_access member_login
```

个人资料响应示例：

```json
{"code":0,"message":"ok","data":{"id":"1","username":"demo_member","nickname":"示例会员","revision":"2026-10-30T06:00:00.123456Z","createdAt":"2026-10-30T06:00:00Z"},"traceId":"example-trace-id"}
```

修改昵称时将最新 revision 与 nickname 一起发送到 `PATCH /api/v1/me`；收到 409 后重新读取资料再编辑，不自动覆盖。

生成的 [app OpenAPI](openapi/app.openapi.json) 与 [TypeScript 类型](../web/admin/src/api/generated/app-schema.ts) 覆盖全部 C 端契约。[独立 client 示例](../web/admin/src/api/member-client.ts) 不引用管理员 store；显式登录和刷新后由 C 端调用方将 access token 保存在内存，refresh token 由 Cookie 管理。

## 验证与边界

`make test-member` 使用独立临时数据库验证重复注册、双实例刷新重放、并发资料更新、启停/改密撤销、权限失败关闭、审计失败回滚。`make member-smoke` 验证真实 HTTP → RPC 链路、身份隔离、Cookie、DTO、权限拒绝和两个数据库账号的隔离；唯一命名的测试数据结束后清理。

首版未提供短信/邮件、找回密码、第三方登录、头像上传、销户、会员管理端代改密码或 C 端页面。账号和会话的保留/清理策略需按实际项目定义；当前不自动删除过期会话记录。新增模块的复制顺序见[开发教程](new-module-guide.md)。
