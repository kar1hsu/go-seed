# 照着 Member 开发新模块

Member 是完整样板：两个 HTTP 入口、独立 RPC、领域应用层、GORM、迁移、跨域授权、权限菜单、前端 client、测试和部署都已连通。新增功能时选择真实需要的部分，不复制会员认证到每个领域。

## 先确定数据归属

如果只是扩展会员资料，在 `services/member` 内增加字段和用例，不创建新 RPC。如果是生命周期和数据所有权独立的新领域，才建立 `services/<domain>`。C 端身份仍由 member 管理，管理员权限仍由 system 管理；其他领域不得直接读取这两个库。

## 文件对照

| 开发内容 | 可参考的实际文件 |
| --- | --- |
| HTTP 契约与公开/管理 DTO 区分 | [C 端契约](../apps/app/api/desc/member/auth.api)、[后台契约](../apps/admin/api/desc/member/users.api) |
| RPC 用例契约 | [member.proto](../services/member/rpc/member.proto) |
| 业务规则、授权与事务 | [会员用例](../services/member/internal/application/account/profile.go)、[后台用例](../services/member/internal/application/management/service.go) |
| GORM 查询、锁、分页快照 | [repository](../services/member/internal/repository/users.go) |
| 版本迁移 | [SQL](../services/member/migrations/00001_member_users.sql)、[显式命令](../services/member/cmd/member/main.go) |
| 依赖装配与跨域适配 | [ServiceContext](../services/member/rpc/internal/svc/service_context.go) |
| 领域权限与接口映射 | [权限](../services/member/resources/permissions.yaml)、[管理端映射](../apps/admin/api/resources/endpoints.yaml)、[C 端映射](../apps/app/api/resources/endpoints.yaml) |
| 后台 feature | [页面](../web/admin/src/features/member/users/MembersPage.vue)、[请求适配](../web/admin/src/features/member/users/api.ts) |
| C 端独立 client | [member-client.ts](../web/admin/src/api/member-client.ts) |
| 并发/故障回滚与真实接口测试 | [领域测试](../services/member/internal/application/account/service_integration_test.go)、[HTTP 测试](../services/member/internal/application/account/http_integration_test.go) |

## 推荐开发顺序

1. 写出数据所有者、允许操作、状态变化和幂等规则。明确管理员操作与 C 端操作的身份来源、数据范围；不能信任请求 body 的 user ID 或 permission。
2. 新领域只建立需要的 `rpc/internal`、领域 `internal/application|repository|model`、`migrations` 和 `resources`。保留根目录一个 go.mod，不创建空 worker、scheduler 或通用 CRUD 基类。
3. 先写 `.api` / `.proto`，分别定义公开 DTO 与管理 DTO。把 HTTP 契约导入入口 `.api`，在对应 endpoints.yaml 声明每个路由的权限或身份要求。ID 用字符串；revision 用不透明版本值；不要直接返回 GORM 实体。
4. 给 `scripts/gen.sh` 的显式 domain 列表加入新领域，再执行 `make gen-api`、`make gen-rpc DOMAIN=<domain>`。初始 logic、handler、ServiceContext 可以实现，pb/client/server/routes/types 通过生成维护。
5. 实现应用层和仓储。管理操作由用例固定权限键，经 system CheckAccess 获取可信 actor ID。system 的 `DomainCallers` 为新服务配置独立凭证与精确权限前缀，不能授予 `system.` 范围。C 端其他领域的用户会话查询需先设计 member 公开授权契约，不能借用管理员 CheckAccess。
6. 在本领域事务内修改业务数据和审计/可靠任务。参考 member 的单行锁与 revision；需要队列时参考 [system outbox](files-and-tasks.md)，不把 MySQL 事务跨 RPC 传递。
7. 添加独立数据库账号、显式迁移命令、启动配置与健康检查。更新 `.env.example`、初始化脚本、Makefile、dev.py、Compose 和 Dockerfile。本地配置只追加缺省值，不覆盖已有密码。
8. 添加领域 permissions.yaml。`make gen-web` 会聚合全部领域权限到 system 的生成目录；更新菜单声明和组件白名单后，运行 `make sync-resources`。不能直接编辑 permissions.generated.yaml 或自动给普通角色授权。
9. 前端按 `features/<domain>/<resource>` 组织，只依赖共享组件/工具与生成的 HTTP 类型。按钮权限只控制交互，后端继续执行最终授权。复制列表时保留请求序号和卸载防护，详情和提交错误可重试，冲突要重新读取。
10. 添加有意义的越权、并发和故障测试；运行 `make check test build`、对应领域集成测试及 HTTP smoke。`make verify-gen` 验证新增契约不会覆盖手写实现。最后补齐服务 README、调用示例和当前限制。

新字段只需新增迁移并修改契约/实体/用例/页面，不改已经执行过的迁移。领域命名、环境变量、端口和服务凭证逐项检查，禁止在原目录上进行全仓库字符串替换。

## 交付检查

- API 没有业务数据库连接；跨领域只依赖公开 RPC 契约。
- 输入校验实际执行，未授权用户不能调用写操作；直接 RPC 调用也需要服务身份。
- 用户隔离、唯一性、事务回滚和重复请求语义有测试，公开响应没有内部字段。
- SQL、环境样例、启动/迁移步骤、生成文件与类型一致。
- 新代码重新生成无差异，已有手写代码保留。

当前提供可运行样板和复制教程，尚未提供自动生成完整业务模块或重写品牌/module 的项目生成器。
