# 管理员认证

本文说明管理员登录与会话管理，使用 GORM、MySQL 和 go-zero Redis。管理员列表、角色授权和菜单按钮见 [角色授权](authorization.md)；使用独立身份与密钥的业务会员见 [会员模块](member-module.md)。

## 初始化

```bash
make init
make infra-up
make migrate
make bootstrap
make dev
```

`make init` 保留已有 `.env` 值，为空缺的管理员 Ed25519 密钥对、独立 `ADMIN_REFRESH_KEY`、`ADMIN_RPC_SECRET` 和 `ADMIN_BOOTSTRAP_PASSWORD` 生成随机值；已有密钥会校验配对关系，不自动轮换；文件权限为 0600。初始账号为 `admin`，密码在本地 `.env` 中查看。bootstrap 只创建缺失账号，不覆盖密码、显示名或启用状态；首次初始化还会授予超级管理员角色，之后不恢复被移除的角色关联。

只使用容器时：

```bash
make init
make compose-up
make compose-migrate
make compose-bootstrap
```

迁移与初始化由独立命令执行，不在 API/RPC 启动过程中自动修改表。健康探测检查进程及 MySQL/Redis 连接；登录前必须完成认证表迁移。

自定义初始账号或禁用账号：

```bash
python3 scripts/system.py -username operator -name 运维管理员 bootstrap
python3 scripts/system.py -username operator disable
# 容器方式在参数最前添加 --compose
```

这些是拥有数据库配置的本地运维命令。禁用操作增加 token version 并撤销全部会话；重新启用账号也不会恢复旧会话。管理员管理接口同时校验角色权限；最后一个启用的超级管理员不能被禁用。

## 数据与会话

- `administrators` 保存规范化小写账号、bcrypt 密码摘要、启用状态和 token version。用户名为 3–32 位小写字母、数字、点、下划线或连字符，以字母开头；初始密码和新密码为 12–72 字节。
- `admin_sessions` 保存管理员会话，默认绝对有效期 7 天；access token 默认有效期 15 分钟。
- `admin_refresh_tokens` 仅保存 refresh token 的 SHA-256 摘要、所属会话、使用时间，以及重试标识和后继令牌的摘要。原始值不入库。
- 刷新在数据库事务中串行轮换。刷新可携带 `X-GoSeed-Refresh-Key`（32 字节随机数的无 padding base64url，43 字符）：相同标识在 60 秒内重试，且后继 token 尚未消费时，返回同一后继 refresh token。其他已消费 token 重用仍提交会话撤销并返回 401。伪造的随机值不会撤销其他有效会话。
- 退出撤销当前会话；改密和禁用撤销全部会话。新授权检查读取主库，检查账号、版本、session 到期及撤销状态，不缓存允许结果。
- 两个 RPC 实例共享数据库即可保持上述语义，不依赖进程内锁或 Redis 通知。已经通过校验的在途请求不追溯取消。

goose **v3.26.0** 使用 Provider API 和嵌入的 SQL 文件，MySQL 命名锁用于串行迁移。DDL 不支持事务回滚，因此 SQL 使用可重复执行的建表语句；Down 会删除认证数据，仅供明确的数据重置流程使用，当前命令不暴露自动回滚。

## HTTP 与 RPC

API 契约为 [auth.api](../apps/admin/api/desc/auth/auth.api)，RPC 契约为 [system.proto](../services/system/rpc/system.proto)。生成的 [OpenAPI](openapi/admin.openapi.json) 包含 Bearer、Cookie 与 CSRF 头说明。

HTTP 成功响应保持 `{code,message,data,traceId}`。登录与刷新返回 accessToken、expiresAt 和 admin，refresh token 仅通过 `Set-Cookie` 发给浏览器。

Cookie 为 `goseed_refresh`，HttpOnly、SameSite=Strict、host-only、Path=`/admin-api/v1/auth`。所有认证 POST 要求 `X-GoSeed-CSRF: 1`；有 Origin 时须匹配 `ADMIN_ALLOWED_ORIGINS`，跨站请求被拒绝。GET `/me` 和 POST `/password` 使用 Bearer token，不信任客户端的用户 ID 或角色头。

API 和 RPC 都验证管理员 token 的签名、EdDSA（Ed25519）算法、issuer、audience、有效期和主体类型。RPC 对认证方法额外校验 `admin-api` 服务凭证和方法白名单；未知方法默认拒绝。公开 Ping/CheckReady 只提供非敏感健康信息。

登录使用 Redis 分层限流：每个可信客户端 IP 每分钟 60 次、每个 IP + 账号每分钟 10 次。改密按已认证账号每分钟 5 次；认证服务最多并发执行 4 个密码操作，满载直接返回 429。Redis 不可用时失败关闭。默认只信任 HTTP 连接的源地址；仅当直连代理命中 `TRUSTED_PROXY_CIDRS` 时从右向左解析 X-Forwarded-For，不能用客户端自填的 IP 或账号锁死另一来源的配额。Compose 只信任独立 ingress 网络上的 gateway 地址；直接访问 API 不信任转发头。

`GOSEED_MODE=pro` 时要求 Secure Cookie、HTTPS Origin 和经过 CA/主机名验证的 RPC TLS；参考[生产配置](production.md)。密钥由部署系统注入，不复用开发 `.env`。

API 关闭包含正文的默认请求日志，RPC 关闭认证方法的正文记录，API 的 RPC 客户端关闭可能输出完整请求和响应的 Duration 日志。监控指标和 Trace 仍可用。

## 浏览器行为

刷新重试标识在 sessionStorage 中暂存，网络失败保留、成功或退出清除；它不是访问凭证。access token 只保存在 Pinia 内存，刷新页面通过 HttpOnly Cookie 恢复会话。单页并发请求合并刷新；支持 Web Locks 的浏览器还会跨标签页串行 Cookie 轮换。不支持 Web Locks 的浏览器并发刷新可能触发重用保护并要求重新登录。

退出和改密通过 BroadcastChannel 通知同源标签页。路由切换会检查当前会话；短期 token 的剩余寿命按响应 Date 和单调时钟计算；401 时尝试刷新，读取请求最多重试一次，刷新确认失效才跳转登录并保留原页面。网络错误、503 不清除会话。写请求不会自动重放。密码更新成功后清除本机 Cookie，并要求所有设备重新登录。

## 验证

```bash
make check test build
make test-auth   # MySQL 上的认证用例，包含 race 与两个服务实例并发刷新
make auth-smoke  # 运行中的 admin-api HTTP 登录/刷新/退出/改密链路
```

测试创建随机账号并在结束后删除自身记录，不修改初始管理员。`AUTH_SMOKE_URL` 可以指向同源 gateway，例如 `http://127.0.0.1:8088`。迁移重复执行、bootstrap 不覆盖、旧刷新令牌重用、会话撤销和非管理员 token 拒绝均有检查。

登录频率限制覆盖客户端 IP 和 IP + 账号维度。过期会话及刷新历史目前保留以支持重用检测，尚未实现其定时清理；当前 Asynq 清理任务处理文件、临时上传目录及按配置过期的审计记录。

### 本地验收记录（2026-10-29）

| 检查 | 实测结果 |
| --- | --- |
| `make check test build` | 格式、文档链接、边界、vet、生成一致性、Go 测试、前端类型及构建通过 |
| `go test -race ./pkg/... ./apps/... ./services/...` | 通过；macOS 链接器有 LC_DYSYMTAB 警告，测试退出码为 0 |
| `make test-auth` | MySQL 集成与 race 通过，覆盖两个独立服务实例、并发刷新、重用撤销、改密、禁用和过期 |
| `make smoke` | 两个真实 API → RPC → MySQL/Redis 链路通过 |
| `make compose-migrate compose-bootstrap` | 容器内迁移可重复执行，已有管理员保持不变 |
| `AUTH_SMOKE_URL=http://127.0.0.1:8088 make auth-smoke` | HTTP 登录、Cookie、当前用户、刷新、退出、改密、来源和主体类型校验通过 |
| 浏览器 | 错误密码提示、登录、页面刷新恢复、账户安全页面、双标签页同时刷新及退出同步通过 |

HTTP 集成测试使用临时账号验证改密，浏览器检查未修改初始管理员密码。日志抽查未发现认证密钥、初始密码及测试密码样本。

## JWT 签发与验签密钥

JWT 库保持 `github.com/golang-jwt/jwt/v4 v4.5.2`。`system-rpc` 使用 `ADMIN_TOKEN_PRIVATE_KEY` 签发，`admin-api` 仅使用 `ADMIN_TOKEN_PUBLIC_KEY` 验签；两者分别装配 `identity.Signer` 与 `identity.Verifier`，后者没有签发或刷新凭证派生方法。算法固定为 EdDSA，不接受 HS256、none 或客户端选择的其他算法；issuer、audience、kind、有效期、会话与版本校验保持原有规则。

`ADMIN_REFRESH_KEY` 独立用于刷新响应丢失后的 HMAC 恢复，只交给签发领域，不发给 API 或 worker。JWT 密钥轮换不影响该派生结果。会员使用 `MEMBER_` 对应的独立密钥；格式、进程分配及更新方式见[生产配置](production.md#身份密钥)。

切换后旧 HS256 access token 会被拒绝。有效的 opaque refresh Cookie 仍可正常换取 EdDSA access token，无需清空账号或会话表；切换前已经发生响应丢失且仍处于旧 HMAC 恢复窗口的请求，可能需要重新登录。两端 API/RPC 应协调更新，避免新旧算法实例混跑。
