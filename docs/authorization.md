# 管理员与角色授权

当前实现管理员列表/新增/编辑/启停/角色分配、角色列表/新增/编辑/启停/授权/删除，以及菜单和按钮权限。管理页面延续 Vben 风格的浅色布局，采用筛选、分页表格和编辑抽屉。API 经 RPC 调用 system application，数据库和 Casbin 只由 system 持有。

## 初始化与升级

```bash
make migrate               # 包含 00002_admin_rbac.sql
make bootstrap             # 同步目录，首次授予初始管理员 super_admin
make sync-resources        # 后续发布：显式同步权限目录
# 容器模式对应 compose-migrate / compose-bootstrap / compose-sync-resources
```

bootstrap 不修改已有密码；`policy_states.bootstrapped` 记录首次超级管理员初始化，后续执行不会恢复被移除的角色。没有可用超级管理员时应排查变更，通过受控运维恢复，不能依赖重复 bootstrap 提权。

普通角色仅拥有明确勾选的权限，同步目录不会自动追加授权。内置 `super_admin` 是例外：始终拥有全部有效目录权限。它不可改名、停用、删除或编辑权限；新增超级管理员只允许现有超级管理员分配此角色。不能停用自己，也不能移除或停用最后一个启用的超级管理员，领域禁用命令遵守同样规则。

## 权限与契约

| 权限键 | 页面/操作 |
| --- | --- |
| `system.user.read` | 管理员菜单、列表 |
| `system.user.create` | 新增管理员 |
| `system.user.update` | 修改姓名、启用/停用 |
| `system.user.assign_role` | 分配角色，同时需要 `system.role.read` |
| `system.role.read` | 角色菜单、角色列表、权限目录 |
| `system.role.create` | 新增角色 |
| `system.role.update` | 修改名称/说明、启用/停用 |
| `system.role.grant` | 设置菜单、按钮权限 |
| `system.role.delete` | 删除未分配给成员的角色 |

按钮权限依赖对应菜单读取权限，前端自动补齐依赖，服务端拒绝缺失依赖的授权请求。普通管理员不能授予自己没有的权限，也不能操作权限超过自己的角色或账户；禁用角色的潜在权限仍参与范围判断。普通管理员不能修改自己的角色关联。页面中的 `manageable`、`assignable` 只辅助交互，保存时服务端重新检查。

权限目录：[permissions.yaml](../services/system/resources/permissions.yaml)；路由声明：[endpoints.yaml](../apps/admin/api/resources/endpoints.yaml)；菜单声明：[menu.yaml](../web/admin/src/router/menu.yaml)。`make gen-web` 和 `make check` 检查重复项、权限引用、组件白名单和 HTTP 路由覆盖。运行中的同步命令只写权限目录，不写菜单、角色授权或成员关系。菜单编译为 `resources.generated.ts`，浏览器用本地白名单映射页面，不接收动态脚本或组件路径。

HTTP 前缀 `/admin-api/v1/system`，所有管理接口要求管理员 Bearer token；写操作另要求 `X-GoSeed-CSRF: 1` 和同源策略。

| 方法/路径 | 行为 |
| --- | --- |
| GET / POST `/administrators` | 筛选分页 / 新增，创建成功为 201 |
| PATCH `/administrators/{id}` | 姓名与状态 |
| PUT `/administrators/{id}/roles` | 替换关联角色 |
| GET / POST `/roles` | 筛选分页 / 新增，创建成功为 201 |
| PATCH `/roles/{id}` | 名称、说明与状态 |
| PUT `/roles/{id}/permissions` | 替换角色权限 |
| DELETE `/roles/{id}?revision=...` | 删除未使用的角色 |
| GET `/permissions` | 有效权限目录与依赖 |

ID 为十进制字符串；列表默认 `page=1&pageSize=20&enabled=all`，每页最多 100。`enabled` 支持 all/enabled/disabled，关键词最多 64 字。更新、授权、删除必须携带列表记录的 `revision`；冲突返回 409，重新加载后再操作，不自动覆盖其他人的变更。goctl 对 DELETE 的 form 标签输出 formData，生成脚本按 httpx 实际查询参数行为归一化为 OpenAPI query。

账号为 3–32 位小写 ASCII 字符，以字母开头，可含数字、点、下划线和短横线；初始密码为 12–72 字节。新账号默认无角色，创建后显式分配。此批不提供管理员硬删除和替其他人重置密码，用户可以在账户安全页自行改密。

## 一致性与撤销

关系表 `roles / permissions / admin_roles / role_permissions` 是授权事实来源，没有第二套 casbin_rule。授权器每次从主库检查账号、会话、token version 与 `policy_states.version`，再使用匹配版本的不可变 Casbin 快照。关系和版本从同一个可重复读事务快照加载，多实例无需依赖通知才发现更新。主库或策略加载失败返回不可用，不用旧缓存放行。

所有角色和权限写操作先锁定策略状态行，在同一事务更新关系与版本；并发变更因此串行化，也保护最后一个超级管理员。禁用账户撤销全部会话，重新启用不会恢复旧 token。撤销提交后启动的新授权检查立即生效，已通过检查的在途读取不追溯取消。

`GET /auth/me` 返回 permissions、roles、policyVersion、superAdmin。前端在导航、窗口重新激活、可见页面每 30 秒，以及本浏览器授权保存通知后更新菜单和按钮；403 后刷新权限。UI 更新间隔不影响后端即时拒绝请求。写操作失败后不自动重放。

当前未接入 Redis watcher、远程跨领域 `CheckAccess`、可视化菜单编辑和业务审计表。watcher 可作为提前刷新优化，不能替代权威版本确认。字典、配置和审计按实施路线继续接入。

## 验证

- `make test-rbac`：本地 MySQL root 连接创建随机 `goseed_rbac_test_*` 临时库，结束只删除该次创建的库。覆盖范围授权、两实例撤销、停用角色、编辑冲突、最后超级管理员并发保护、重复同步/初始化、存储失败拒绝授权，带 race。
- `make rbac-smoke`：已运行 API 上验证 201、分页筛选、CSRF、伪造身份头、403、409、角色分配/撤销和删除。先运行 migrate/bootstrap；测试只创建并清理本次随机命名的账户/角色。
- `make test-auth`、`make auth-smoke`：认证回归。
- `make check test build`：格式、边界、文档链接、vet、前端类型/构建、重复生成，以及退出/切换账号/并发轮询时旧响应不能恢复会话或权限的前端回归。CI 同样运行上述集成与 HTTP 检查。
