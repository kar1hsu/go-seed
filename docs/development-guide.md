# GoSeed 开发约定

开发从契约和数据归属开始，按 HTTP、RPC、业务用例、数据访问、后台页面逐层实现。当前 Makefile 已提供生成、检查、构建、迁移和认证验证入口；实际命令及后续计划在下文分别列出。

## 以会员样板新增模块

当前可运行样板为 `member`：C 端注册登录和个人资料，后台查询与启停，独立 RPC、数据库账号和跨领域授权。完整文件对照与复制步骤见[新增模块教程](new-module-guide.md)，接口示例见[会员模块](member-module.md)。

同一领域新增资源时，直接增加用例、仓储和契约，不新建微服务。新领域以数据归属和生命周期为依据，按需创建目录；只有身份领域维护会话，不能把会员登录复制到每一个业务服务。

## HTTP 契约

| 项目 | 约定 |
| --- | --- |
| 管理端前缀 | `/admin-api/v1` |
| 业务端前缀 | `/api/v1` |
| 列表查询 | `GET /member/users?page=1&pageSize=20`，pageSize 上限默认 100 |
| 详情与更新 | `GET/PATCH /member/users/:id`，路径以 go-zero 契约语法声明 |
| 创建与删除 | `POST /member/users`、`DELETE /member/users/:id` |
| 业务动作 | `PUT /member/users/:id/status`，明确动作语义 |
| ID | HTTP 与 RPC 中统一传十进制字符串，避免 JavaScript 大整数精度损失 |
| 时间 | HTTP 使用 UTC RFC3339，数据库统一 UTC；内部 RPC 使用明确的时间类型 |
| 金额 | 整数最小单位或十进制定点字符串，契约注明币种和单位 |
| 列表结果 | `items`、`total`、`page`、`pageSize`，空列表是 `[]` |

成功响应形状：

```json
{
  "code": 0,
  "message": "ok",
  "data": { "id": "10001", "nickname": "示例会员" },
  "traceId": "example-trace-id"
}
```

失败使用对应 HTTP 状态，并返回相同顶层结构，`data` 为 null。常规查询返回 200，创建返回 201，异步受理返回 202；需要响应体的删除返回 200，不向 204 写入 JSON。

`.api` 必须表达最终 JSON 的完整结构，包括响应包裹层。模板和 DTO 适配器统一填写基础字段；不在 logic 和全局响应处理器中重复包裹。OpenAPI、前端客户端及实际响应以契约测试对齐。文件下载、流式响应和健康检查明确采用各自协议，不套 JSON 包裹层。

### 错误处理

| 情况 | HTTP | gRPC |
| --- | --- | --- |
| 参数不合法 | 400 | InvalidArgument |
| 未认证、会话撤销 | 401 | Unauthenticated |
| 没有权限 | 403 | PermissionDenied |
| 资源不存在 | 404 | NotFound |
| 重复创建、状态冲突 | 409 | AlreadyExists / FailedPrecondition / Aborted |
| 限流 | 429 | ResourceExhausted |
| 依赖不可用 | 503 | Unavailable |
| 超时 | 504 | DeadlineExceeded |
| 请求取消 | 408 | Canceled |
| 未知内部错误 | 500 | Internal |

业务错误码由 `pkg/xerr` 约定公共范围，各领域分配独立范围，例如通用 100xxx、system 200xxx、member 300xxx。第一阶段只有公共错误映射，由 gRPC status code 映射稳定 100xxx 错误码；后续领域业务错误通过 status 与结构化 details 传递稳定错误标识，API 只映射已知标识，未知错误统一返回通用消息并记录完整内部日志。

参数校验分两层：API 检查类型、必填、长度和范围；application 检查状态、唯一性、权限与数据归属。若使用 validator 标签，必须接入实际执行路径，不能只写标签。幂等键、唯一约束和事务共同处理重复写请求，禁止仅凭前端按钮禁用保证幂等。

## RPC 与依赖装配

一个 proto 初期对应一个领域 service，例如 `System`，package 为 `goseed.system.v1`。消息字段编号不重用，删除字段使用 reserved；新增字段优先保持向后兼容。跨服务滚动升级先兼容旧调用方，再切换调用方。

RPC 方法按业务用例命名，避免 `Execute(action, json)` 形式。只暴露用例需要的数据，不返回 GORM 结构。API 与 worker 传入原有 context，取消与 deadline 继续向数据库、Redis 和下游传递。

装配在每个进程的 `NewServiceContext` 完成，可以调用共享的资源构造函数；业务代码只接收依赖，不负责打开连接。测试用小范围接口或 fake 替换依赖，不引入运行时服务定位器。

## 代码生成

生成器版本、模板和参数必须固定。项目的 go-zero 版本在 `go.mod`，工具版本在 `tools/versions.env`，前端包管理器版本在 `package.json`，依赖由 lockfile 固定。上游文档中的 `@latest` 不能作为可重复构建配置。

当前已有根 `go.mod`、契约和生成脚本，统一使用 `make gen` / `make gen-api` / `make gen-rpc` / `make gen-web`。下面是核心参数说明，实际脚本还固定 `--home templates/goctl`。

```bash
# 从仓库根目录运行
goctl api validate --api apps/admin/api/desc/admin.api
goctl api go --api apps/admin/api/desc/admin.api --dir apps/admin/api --style go_zero

goctl rpc protoc services/system/rpc/system.proto \
  --proto_path=services/system/rpc \
  --go_out=. --go_opt=module=github.com/kar1hsu/go-seed \
  --go-grpc_out=. --go-grpc_opt=module=github.com/kar1hsu/go-seed \
  --zrpc_out=services/system/rpc --style=go_zero --name-from-filename
```

对应 proto 的关键头部：

```protobuf
syntax = "proto3";
package goseed.system.v1;
option go_package = "github.com/kar1hsu/go-seed/services/system/rpc/pb";
```

采用上述布局时，单 service 名为 `System`，goctl 生成的 client 位于 `services/system/rpc/system/`。这些路径已由固定 goctl 1.10.2 生成并编译验证。`go_package` 由路径推断包名 pb；当前生成器对 `;pb` 显式别名处理不一致，故不使用该写法。`make verify-gen` 检查重复生成及新增契约保留手写实现。

API 按 `.api` 的 `@server` group 分目录。自定义模板先解决统一响应、认证、日志和错误映射，再批量生成服务；只在契约变化时重新生成，业务逻辑修改无需重新生成。[goctl API](https://go-zero.dev/reference/cli-guide/api/)、[goctl RPC](https://github.com/zeromicro/go-zero/blob/master/tools/goctl/rpc/README.md)

生成结果进入版本控制。CI 在干净副本中运行相同命令，要求已跟踪文件无差异且没有遗漏的新生成文件；同时检查新增接口不会覆盖已有 logic、handler 和装配代码。

## 数据库与任务

迁移采用时间序号加含义命名，一经发布不能修改已执行的迁移，修复通过新迁移完成。默认主键使用领域内唯一的 BIGINT；若业务需要跨域全局 ID，再选择专门方案。

repository 接口接收 context，分页、排序字段使用白名单。软删除按业务决定，不给所有表强制增加软删除。事务由 application 持有，外部 RPC 和慢任务不在长时间数据库事务中执行。

任务消费者必须验证 payload 版本和必要字段，对可重试和永久失败分类；任务成功的依据是业务状态已经完成，而不只是 handler 返回 nil。重试耗尽的任务保留状态与原因，并提供带权限与审计的重试途径。创建外部副作用时，使用稳定幂等键或业务状态机。

## 前端约定

API client 统一处理基础 URL、超时、错误码、request ID 和 refresh 并发；feature 中不散落裸 fetch/axios 调用。自动重试仅用于明确可重试的请求，写入请求需要幂等语义。

管理后台 access token 默认存内存，refresh token 使用 HttpOnly、Secure cookie，并限制 Path 与 SameSite。刷新、注销等 cookie 接口校验 Origin 与 CSRF 防护；生产建议通过同源反向代理部署。跨域部署需要单独配置允许来源与凭证，不能使用通配来源加凭证。

页面的列表筛选、加载、空状态、错误状态、确认交互采用共享组件和 composables；业务字段、校验及操作留在 feature 内。菜单路由必须从编译时组件白名单解析，权限变更后刷新菜单和按钮状态。

## 开发命令

当前已实现 doctor、infra-up、dev、gen-api、gen-rpc、gen-web、check、test、build、migrate、bootstrap，完整清单与实际行为见根 README。下表也保留后续阶段的设计接口；sync-resources 已实现；new-module 尚未实现，不应提前调用。

| 命令 | 行为 |
| --- | --- |
| `make doctor` | 检查 Go、goctl、protoc、插件、Node、pnpm 及固定版本 |
| `make infra-up` | 启动本地 MySQL 与 Redis |
| `make migrate` | 运行 system 领域 SQL 迁移 |
| `make bootstrap` | 幂等创建初始管理员，不覆盖已有账号 |
| `make dev` | 启动 API、RPC、管理端开发服务并统一处理退出 |
| `make gen-api APP=admin` | 校验并生成选定 API |
| `make gen-rpc DOMAIN=system` | 生成指定领域 RPC |
| `make gen-web` | 从契约生成 OpenAPI 与 TypeScript client |
| `make sync-resources` | 显式同步 system 权限目录，保留普通角色授权；菜单由 gen-web 校验和编译 |
| `make check` | 格式、静态检查、类型检查及契约一致性 |
| `make test` | 不依赖外部服务的测试，包含内存 gRPC；真实依赖单独用 `make smoke` |
| `make build` | 构建各进程与前端 |
| `make new-module DOMAIN=<domain> RESOURCE=<resource>` | 样板稳定后增加；默认不覆盖已有文件 |

首版先实现少量脚本和 Makefile，不先开发一个复杂脚手架 CLI。新建项目时替换 module、镜像名、环境前缀与品牌配置，再运行契约生成及检查，避免全仓库无差别字符串替换。

## 验证要求

纯文档变更检查链接、示例和设计一致性。实现阶段按改动范围执行 Go 格式化、必要的依赖整理、build、test、vet 及前端类型检查；不把尚不存在的测试命令报告为通过。

第一版重点覆盖：登录与刷新轮换、会话撤销、角色授权变更、多实例策略加载、公开接口字段过滤、业务事务回滚、重复任务、迁移重复运行，以及从空环境启动完整链路。并发相关代码补充 race 检查，生成器升级补充重复生成对比。

当前生成链路为 `.api` → goctl Swagger 2 → OpenAPI 3（合并公共错误 schema）→ openapi-typescript；运行时使用 openapi-fetch。Health 路由独立注册，框架级未匹配路由/方法不属于业务 envelope。模板只对解析错误映射做小范围定制，避免把业务实现放进模板。

## 管理员认证开发

当前认证实现、初始化命令、HTTP Cookie 与会话撤销规则见 [管理员认证](authentication.md)。修改认证行为时执行 `make test-auth`；对运行中的完整链路执行 `make auth-smoke`。这两个命令使用本项目开发 MySQL，创建随机测试账号并仅删除自身账号。认证令牌、密码和 Cookie 不写入日志或提交的测试产物。
