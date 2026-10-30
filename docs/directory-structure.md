# GoSeed 目录规范

仓库按“客户端入口、业务领域、公共技术能力、前端”组织。以下为完整目标目录；第一阶段已落地入口服务、健康用例、技术包、前端与工具，认证授权、文件与任务相关 application、repository、model、迁移及领域命令已接入，worker/scheduler 已运行；其余业务目录按需创建。

## 顶层目录

```text
go-seed/
├── README.md
├── AGENTS.md                         # AI 开发入口与本项目边界
├── go.mod                            # 唯一 Go module
├── go.sum
├── Makefile                          # 统一开发命令
├── .env.example                      # 本地环境变量清单，无真实凭证
├── .gitignore
├── .github/workflows/                # 检查、构建与发布
├── apps/                             # 面向客户端的 HTTP 入口
│   ├── admin/api/                    # 管理端 API
│   └── app/api/                      # 网站、App、小程序 API
├── cmd/probe/                        # 容器健康检查客户端
├── services/                         # 按数据所有权划分的领域
│   ├── system/                       # 管理基础能力
│   └── catalog/                      # 后续加入的商品开发样板
├── pkg/                              # 无业务所有权的共享技术代码
├── web/
│   └── admin/                        # Vue 管理后台
├── deploy/
│   ├── compose/                      # 开发与小规模交付
│   ├── docker/                       # 各进程 Dockerfile
│   └── nginx/                        # 静态页面与入口代理
├── scripts/                          # 生成、迁移、校验、项目初始化
├── tools/
│   └── versions.env                  # goctl、protoc、插件等工具版本
├── templates/
│   ├── goctl/                        # 固定版本的 API/RPC 模板
│   └── module/                       # 验证样板后再提取的业务模板
├── tests/
│   ├── integration/                  # 数据库、Redis、RPC 集成测试
│   ├── e2e/                          # API 与后台业务链路
│   └── fixtures/                     # 非敏感测试数据
└── docs/
    ├── architecture.md
    ├── directory-structure.md
    ├── development-guide.md
    ├── roadmap.md
    ├── ai-development.md
    └── openapi/                      # 实现阶段生成的接口文档
```

AI 官方参考资料已按固定子模块接入，目录见 [AI 开发配置](ai-development.md)，与服务运行时依赖分开。

不使用根目录 `internal/app` 装配所有服务，也不设置根目录共享的业务 `model`、`repository`。`pkg` 不是所有代码都能随意放入的公共仓库。

## HTTP API 目录

```text
apps/admin/api/
├── admin.go                          # main，保持 goctl 入口位置
├── desc/
│   ├── admin.api                     # 契约入口，导入子文件
│   ├── base.api                      # 本 API 的通用响应与分页类型
│   ├── auth/auth.api                 # 当前登录与会话契约
│   └── system/
│       ├── user.api
│       ├── role.api
│       └── menu.api
├── etc/
│   └── admin-api.yaml                # 使用环境变量占位的可提交配置
├── resources/
│   └── endpoints.yaml                # 路由模板到权限的映射声明
└── internal/
    ├── config/config.go
    ├── handler/
    │   ├── routes.go                 # 可重新生成
    │   └── system/user/              # .api 的 group 控制分组
    ├── logic/
    │   └── system/user/              # DTO 转换、调用 RPC、接口聚合
    ├── middleware/                   # 本 API 的认证与请求处理
    ├── svc/service_context.go        # 只装配本进程依赖
    └── types/types.go                # 从 .api 生成
```

`apps/app/api` 采用相同结构，入口为 `app.go`，契约为 `desc/app.api`，配置为 `etc/app-api.yaml`。两个 API 不互相 import 各自的 `internal`，也不调用对方的 HTTP 接口来复用业务。

handler 与 logic 使用 goctl 的 `go_zero` 文件命名风格，例如 `create_user_handler.go`、`create_user_logic.go`。保留 goctl 识别的固定目录和生成文件名，不对生成文件做二次批量搬运。

## 业务领域目录

```text
services/system/
├── README.md                         # 职责、依赖、运行、检查说明
├── rpc/
│   ├── system.go                     # RPC main
│   ├── system.proto                  # 领域公开契约
│   ├── etc/system.yaml
│   ├── pb/
│   │   ├── system.pb.go              # protoc 生成
│   │   └── system_grpc.pb.go         # protoc 生成
│   ├── system/
│   │   └── system.go                 # goctl 生成的 RPC client
│   └── internal/
│       ├── config/config.go
│       ├── server/system_server.go   # RPC 转接层
│       ├── logic/                    # RPC 方法适配
│       └── svc/service_context.go
├── internal/                         # 仅 system 子树内可引用
│   ├── application/
│   │   ├── auth/                     # 登录、刷新、会话撤销
│   │   ├── user/
│   │   ├── role/
│   │   ├── menu/
│   │   ├── setting/
│   │   ├── file/
│   │   └── audit/
│   ├── model/                        # GORM 数据实体
│   ├── repository/                   # 领域持久化实现与本地事务
│   ├── authorization/                # Casbin、策略快照与版本同步
│   ├── session/                      # token、refresh 与 session 规则
│   └── task/                         # 本领域任务类型与 payload
├── worker/
│   ├── main.go
│   ├── etc/system-worker.yaml
│   └── internal/
│       ├── config/
│       ├── handler/                  # Asynq handler 调用 application
│       └── svc/
├── scheduler/
│   ├── main.go
│   ├── etc/system-scheduler.yaml
│   └── internal/
│       ├── config/
│       ├── schedule/                 # Cron 定义与投递
│       └── svc/
├── cmd/
│   ├── migrate/main.go              # 独立运行迁移
│   └── bootstrap/main.go            # 初始管理员与基础数据
├── migrations/mysql/
│   └── 202610290001_create_users.sql # 示例名称，goose Up/Down 格式
└── resources/
    ├── permissions.yaml              # 领域权限目录
    └── casbin_model.conf             # 权限模型
```

`rpc/internal` 仅供 RPC 进程使用，`services/system/internal` 供同一领域的 RPC、worker、迁移程序共享。业务用例放在后者，worker 不通过导入 RPC logic 来复用功能。Go 的 `internal` 规则可以阻止其他领域导入这些实现。[Go 模块组织](https://go.dev/doc/modules/layout)

`rpc/system/` 是 `service System` 在当前 goctl 单 service 模式下的客户端输出目录；使用方可用 `systemclient` 作为 import 别名，避免与领域名称混淆。未来一个 proto 内增加多个 service 时，要显式启用 multiple 模式并同步调整生成目录规范，不能混用两种布局。

新增 `catalog`、`order` 等领域时复制此结构的必要部分。没有异步业务就不创建 worker/scheduler，没有表就不创建 migrations。

## 依赖约束

| 来源 | 允许依赖 | 禁止依赖 |
| --- | --- | --- |
| `apps/*/api` | RPC 的 pb/client、`pkg`、自身 internal | 领域 model/repository/application、其他 API 的 internal |
| 领域 RPC | 自己的 internal、其他领域公开 RPC client、`pkg` | 其他领域内部实现与数据库 |
| 领域 worker | 自己的 application/repository、必要 RPC client、`pkg` | RPC handler/logic、其他领域私有任务 payload |
| 领域 application | 本领域数据访问、抽象依赖、`pkg` | HTTP request/response、API DTO、RPC handler |
| `pkg` | 标准库与通用第三方库 | `apps`、`services`、具体用户或订单模型 |
| `web/admin` | HTTP 契约、自己的共享组件 | 后端数据库模型、内部 RPC 类型 |

跨领域任务不能通过导入其他领域私有 payload、直接写其队列绕过领域入口；先调用目标领域 RPC，由目标领域生成任务。公开事件另设版本化契约，按真实需求引入。

第一版共享包按需要建立：

| 包 | 职责 |
| --- | --- |
| `pkg/xerr` | 通用错误类型、稳定错误码与 RPC 映射 |
| `pkg/xhttp` | HTTP 响应、分页和参数错误处理 |
| `pkg/identity` | 主体类型、上下文访问接口；不含用户表查询 |
| `pkg/rpcx` | 服务身份、metadata 与错误适配的共用逻辑 |
| `pkg/dbx` | 数据库连接构建、事务辅助；不含共享业务实体 |
| `pkg/taskx` | 队列连接、生命周期、Trace 传播 |
| `pkg/storage` | 本地及 S3 兼容存储的技术接口 |

go-zero 已提供的日志、配置、熔断等直接使用，不为了统一名字再套一层空包装。依赖边界先通过 `internal` 和代码审查保证，实施阶段增加 import 规则检查。

## 管理后台目录

```text
web/admin/
├── package.json
├── pnpm-lock.yaml
├── vite.config.ts
├── .env.example
├── public/
└── src/
    ├── main.ts
    ├── app/                          # 应用入口、插件、路由初始化
    ├── layouts/                      # 登录布局、后台布局
    ├── router/                       # 基础路由、权限守卫、组件白名单
    ├── stores/                       # session、菜单、主题等全局状态
    ├── api/generated/                # 从契约生成的类型及基础客户端
    ├── components/                   # 可跨业务复用的 UI 组件
    ├── composables/                  # useTable、usePermission 等
    ├── lib/                          # HTTP、日期、格式化等工具
    ├── styles/
    └── features/
        ├── auth/                     # 登录、刷新与退出交互
        ├── dashboard/
        ├── system/
        │   ├── user/
        │   ├── role/
        │   ├── menu/
        │   ├── setting/
        │   ├── file/
        │   └── audit/
        └── catalog/product/          # 业务样板实现时加入
            ├── pages/
            ├── components/
            ├── api.ts                # 基于 generated client 的业务适配
            ├── routes.ts
            └── menu.yaml             # 菜单和按钮的声明
```

当前系统管理 feature 位于 `web/admin/src/features/system`；菜单集中在 `src/router/menu.yaml`，由 `make gen-web` 编译为 `resources.generated.ts`。上图的按模块分散声明是后续扩展方向。

feature 独有状态留在 feature；不要把所有列表筛选和弹窗状态都塞进全局 store。前端构建时汇总 `menu.yaml`，验证路由、组件及权限引用；运行时只加载白名单中的页面。

## 命名与配置

- 产品展示名为 GoSeed，仓库名为 `go-seed`，Go module 固定为 `github.com/kar1hsu/go-seed`。
- Go 包名采用小写简短单词；手写文件使用 `snake_case.go`，类型与方法使用 Go 常规命名。
- 进程、容器名使用 `admin-api`、`system-rpc` 等短横线形式；目录按 `apps/admin/api`、`services/system/rpc` 表达层级。
- HTTP 资源使用复数名词，前缀为 `/admin-api/v1` 与 `/api/v1`；JSON 使用 camelCase。
- Protobuf package 使用 `goseed.<domain>.v1`，字段使用 snake_case，明确完整 `go_package`。
- SQL 表使用复数 snake_case，迁移由领域拥有；Redis key 同时包含项目、环境、领域。
- 可提交 YAML 使用环境变量占位；`*.local.yaml`、真实 `.env`、证书、构建产物、日志不提交。

每个进程从自身配置构造依赖，不从工作目录隐式寻找一个根级“万能配置”。所有脚本以仓库根目录为入口并使用明确路径，部署命令显式传入配置文件。

## 生成文件的所有权

| 文件类型 | 所有权 | 变更方式 |
| --- | --- | --- |
| `.api`、`.proto` | 人工维护的契约 | 先改契约，再生成 |
| `types.go`、`routes.go`、`*.pb.go`、RPC client/server 转接文件 | 生成器 | 修改契约或模板，重新生成 |
| 初始 handler、logic、config、ServiceContext、main | 生成骨架后由项目维护 | 保持生成器约定；新增方法交由 goctl 补充 |
| application、repository、迁移、前端 feature | 业务开发者 | 正常修改和验证 |
| OpenAPI 与前端 generated client | 派生产物 | 从契约统一生成，不手工修补类型 |

具体保留或覆盖行为取决于固定的 goctl 版本和模板。生成脚本必须验证重复生成不会覆盖手写内容；升级工具或模板时先在临时目录比较结果。删除或重命名接口后，审查并清理遗留骨架文件，不能只运行生成命令就认为清理完成。
