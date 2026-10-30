# 会员样板验证记录

2026-10-30，本地 MySQL 8.4.6、Redis 8.2.1、Go 1.25.11、Node 24.19.0、pnpm 11.15.1。

## 已执行并通过

| 检查 | 结果 |
| --- | --- |
| `go mod tidy` | 无新增运行时依赖 |
| `make check test build` | 格式、文档链接、边界、vet、前端类型/格式、生成一致性、Go/前端测试、六个服务及前端构建通过 |
| `go test -race ./pkg/... ./apps/... ./services/...` | 无外部依赖测试及竞态检查通过 |
| `make test-member` | 临时数据库会员并发、撤销与审计回滚测试通过，75.213 秒 |
| `make test-auth` | 原管理员认证并发回归通过，97.296 秒 |
| `make test-rbac` | 原管理、权限、字典与文件任务集成回归通过，221.391 秒 |
| `make init infra-up migrate sync-resources` | 保留已有配置，新增独立会员库与账号，迁移与权限目录同步通过 |
| `make smoke` | 双 API、RPC 依赖链路与 worker/scheduler 就绪检查通过 |
| `make auth-smoke` | 管理员 HTTP 会话回归通过，11.883 秒 |
| `make rbac-smoke` | 管理 HTTP 回归通过，34.196 秒 |
| `make member-smoke` | 真实双 API / 双 RPC 会员链路通过，5.419 秒 |
| Compose | 当前 9 个容器健康，member RPC 端口不对宿主机开放 |

生成一致性检查包含两个 RPC 的新增方法，确认原有手写 logic、ServiceContext、配置与入口保持不变。

## 会员边界

- 重复注册在两个实例并发执行时只创建一个账号和一条注册事件。
- 并行消费相同 refresh token 触发重放保护，返回过的新 access token 也无法继续使用。
- 并发编辑同一 revision 仅一个成功；旧资料更新返回冲突。
- 退出、改密、停用使旧会话失效；重新启用不恢复旧会话。
- 权限拒绝、会话无效、授权依赖失败均拒绝后台操作；领域不可通过 CheckAccess 请求其他权限前缀。
- 审计写入失败时，注册、资料变更、密码变更和启停全部回滚。
- 真实 HTTP 验证会员与管理员 JWT/Cookie 互相拒绝；普通无授权管理员无法查询或启停会员。
- 公开 DTO 不包含 hash、token version、账号启停或 refresh token；Cookie 有独立名称和 Path，HttpOnly / SameSite=Strict。
- member 数据库账号不能读取 system 管理员表，system 账号不能读取 member 表。

## 浏览器验收

通过同源入口 `http://127.0.0.1:8088/member/users` 验证会员列表、详情、注册事件、停用确认与状态刷新。浏览器停用后，该会员原 access token 请求 `/api/v1/me` 返回 401。页面沿用现有表格、筛选和抽屉样式；浏览器 warn/error 日志为空。

唯一命名的 HTTP、数据库和浏览器测试账号均已清理，包括会话与事件；现有用户配置保留。此前 HTTP 测试曾将管理员创建状态误写为 200，修正为实际契约的 201 后通过；该次留下的唯一测试管理员已按创建 Trace ID 清理。

当前未执行真实短信/邮件、第三方登录或生产 TLS 验证；这些不属于首版实现。当前无 C 端页面，交付的是 C 端 API、独立 client 示例与后台会员页。自动模块/新项目生成器尚未实现。
