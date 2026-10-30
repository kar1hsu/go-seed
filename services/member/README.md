# Member 会员领域

这是 GoSeed 的业务样板，提供 C 端账号密码注册、登录、个人资料与会话管理，以及后台会员查询和启停。管理员与会员使用独立的账号、数据库、签名密钥、issuer、audience、身份类型和 Cookie。

## 结构与运行

```text
internal/application/account/     注册、登录、会话、改密、个人资料
internal/application/management/  后台会员列表、详情与启停
internal/repository/              GORM 查询、行锁、事务和审计
internal/model/                   领域私有数据实体
rpc/member.proto                 公开 RPC 契约
rpc/internal/logic/               DTO 与用例适配
rpc/internal/svc/                 数据库、Redis、system RPC、用例装配
migrations/                      独立 goose SQL 迁移
resources/permissions.yaml       领域权限声明
cmd/member/                      显式迁移命令
```

从仓库根运行 `make init infra-up migrate bootstrap`、`make dev`。`infra-up` 只在 GoSeed 开发 MySQL 中准备 `goseed_member` 数据库与独立账号；生产数据库及权限由部署人员显式配置。启动服务不会自动建表。完整容器启动见[根 README](../../README.md)。

`member-rpc` 本地监听 `127.0.0.1:19002`，Compose 内部监听 `9002`，不映射到宿主机。依赖自己的 MySQL、Redis 和 system-rpc。没有异步业务，因此不创建空的 worker/scheduler。

## 接口与配置

- C 端 HTTP：`/api/v1/auth/register|login|refresh|logout`、`GET/PATCH /api/v1/me`、`POST /api/v1/me/password`。
- 后台 HTTP：`GET /admin-api/v1/member/users`、`GET /users/:id`、`PUT /users/:id/status`，后两项沿用相同 member 前缀。
- RPC 方法及错误见 [RPC 契约](rpc/README.md)，完整调用示例见[会员样板](../../docs/member-module.md)。

配置入口为 `rpc/etc/member.yaml`，环境变量见根 `.env.example` 的 `MEMBER_*`。开发默认注册全局 10 次/分钟、登录全局 120 次/分钟和每账号 10 次/分钟，Redis 故障时拒绝认证尝试。生产应结合入口流量防护配置合适额度。

会员写操作及账号变更事件在本领域同一事务中提交。后台详情展示最近 20 条事件，包含动作、操作主体与 Trace ID；不记录密码、token 或资料正文。当前不向 system 操作日志汇总这些事件。

## 验证

```bash
make test-member        # 独立临时 MySQL 库、并发及故障回滚；结束自动清理
make member-smoke       # 运行中的双 API / 双 RPC，创建并清理唯一命名测试账号
make check test build
```

如何照此新增模块见[模块开发教程](../../docs/new-module-guide.md)。
