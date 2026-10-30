# GoSeed

GoSeed 是基于 go-zero 的项目开发底座：单仓库、单 Go module，管理端 API、业务端 API 与领域 RPC 独立运行，管理后台采用 Vue 3。

**当前已完成工程骨架、管理员认证、角色授权、操作日志、字典与系统配置。** 提供管理员列表/新增/编辑/启停/角色分配，角色管理与菜单按钮授权，版本化 SQL 迁移，API → RPC 探测，GORM/MySQL 与 Redis 就绪检查。管理员/角色变更与审计日志在同一事务提交，可按操作人、对象、类型、日期和 Trace ID 查询。字典支持分类与字典项维护，系统配置支持文本、整数、布尔及 JSON 参数。业务样板及 Asynq worker/scheduler 按后续阶段接入。

## 本地开发

准备 Go **1.25.11**、Node **24.19.0**（见 `.nvmrc`）、pnpm **11.15.1**、Python **3.10+**、Docker 和 Compose v2+，以及 curl、unzip、make。首次安装 pnpm 可使用 Corepack，或按 [pnpm 安装文档](https://pnpm.io/installation) 安装指定版本。

```bash
git clone --recurse-submodules git@github.com:kar1hsu/go-seed.git
cd go-seed
# 已有 checkout 时恢复固定版本的 AI 参考资料
git submodule update --init --recursive
make init                          # 创建 .env，并生成本地认证密钥和初始密码
# 使用 nvm 的开发者先执行 nvm install && nvm use
make tools                         # 固定 Go 生成工具安装到 .tools/bin；protoc 按需下载
pnpm --dir web/admin install --frozen-lockfile
make doctor
make infra-up                      # 仅启动 GoSeed 自己的 MySQL / Redis
make migrate                       # 显式执行 system SQL 迁移
make bootstrap                     # 初始化 admin 与超级管理员角色；已有密码保持不变
make dev                           # 编译并启动三个 Go 进程和 Vite；Ctrl+C 统一退出
```

访问 [管理工作台](http://127.0.0.1:5173)。初始账号为 `admin`，初始密码由 `make init` 随机生成，保存在本地 `.env` 的 `ADMIN_BOOTSTRAP_PASSWORD`。登录后可在账户菜单中修改密码。另一个终端执行 `make smoke`，验证两个 API 的健康检查、RPC 调用、响应结构和 Trace ID。

| 服务 | 本地默认地址 |
| --- | --- |
| 管理后台 | `http://127.0.0.1:5173` |
| admin-api | `http://127.0.0.1:8081` |
| app-api | `http://127.0.0.1:8080` |
| system-rpc | `127.0.0.1:19001`（Compose 内部为 `9001`） |
| MySQL | `127.0.0.1:13306` |
| Redis | `127.0.0.1:16379` |

端口和开发凭证在 `.env` 中配置；不要提交 `.env`。本地只监听回环地址。MySQL 与 Redis 使用独立容器和命名卷，不复用其他项目的实例。`make infra-down` 停止基础设施并保留数据。

```bash
curl http://127.0.0.1:8081/admin-api/v1/system/ping
curl http://127.0.0.1:8080/api/v1/system/ping
curl http://127.0.0.1:8081/healthz     # 进程存活
curl http://127.0.0.1:8081/readyz      # RPC + MySQL + Redis 就绪
```

## 完整容器运行

需要 Docker/Compose 与 Python 3，无需主机 Go/Node 工具链：

```bash
make init                          # 保留已有配置及凭证
make compose-up
make compose-migrate               # 使用容器内的领域命令执行迁移
make compose-bootstrap             # 从本地私有配置初始化管理员
# 打开 http://127.0.0.1:8088
make smoke                         # 主机有 Python 时运行
make compose-down                  # 停容器，保留命名卷
```

容器模式与 `make dev` 使用相同的 API 端口，切换前先停止另一种运行方式。Compose 是开发/参考配置；生产交付需配置 TLS、Secure Cookie、准确的允许来源及密钥管理。管理员认证、逐方法 RPC 服务身份校验和角色权限已接入。RPC 端口不发布到容器宿主机，Nginx 提供同源入口。

## 工程命令

| 命令 | 当前行为 |
| --- | --- |
| `make init` | 创建本地配置和随机认证凭证，保留已有值 |
| `make migrate` / `make bootstrap` | 迁移 system 数据库 / 幂等创建初始管理员 |
| `make test-auth` / `make auth-smoke` | MySQL 认证并发测试 / 运行中 HTTP 登录链路 |
| `make sync-resources` | 同步权限目录，保留角色授权和成员关系 |
| `make test-rbac` / `make rbac-smoke` | 独立临时库权限并发测试 / 运行中 HTTP 管理链路 |
| `make gen` | 生成两份 API、RPC、OpenAPI 3 与管理端 TypeScript 类型 |
| `make gen-api APP=admin` | 从 `.api` 生成指定入口，保留手写骨架 |
| `make gen-rpc DOMAIN=system` | 从 `.proto` 生成 system RPC |
| `make gen-web` | 校验权限/菜单声明，更新 OpenAPI、前端类型及菜单资源 |
| `make verify-gen` | 在临时副本验证重复生成一致、新增契约保留手写实现且可编译 |
| `make check` | 格式、依赖边界、文档链接、vet、前端类型/格式及生成一致性 |
| `make test` | 不依赖外部服务的 Go 测试（包含内存 gRPC）及前端会话/权限竞态测试 |
| `make build` | 分别构建三个服务、probe 与管理端静态资源 |
| `make smoke` | 验证已经运行的真实 API/RPC/MySQL/Redis 链路 |

二进制输出到 `bin/`，前端输出到 `web/admin/dist/`。契约、生成文件与 lockfile 进入版本控制。CI 执行检查、构建、race 测试及真实依赖 smoke；AI 参考库不是运行时构建依赖。

## 目录与技术决策

```text
apps/admin/api/          管理端 HTTP 入口，调用 RPC
apps/app/api/            业务客户端 HTTP 入口，调用 RPC
services/system/rpc/    系统领域契约、客户端与服务
services/system/internal/application/  领域用例，后续与 worker 共享
pkg/                    通用错误、HTTP 与数据库连接技术代码
cmd/probe/              容器健康检查客户端
web/admin/              Vue 3 + TypeScript + Element Plus，按 feature 组织
scripts/                生成、检查、进程监督与 smoke
```

Go module 固定为 `github.com/kar1hsu/go-seed`。运行时使用 go-zero **1.10.2**、GORM **1.31.1**，工具版本见 [versions.env](tools/versions.env)。API 不连接业务数据库，GORM 由 system 的 ServiceContext 装配，不在启动时 AutoMigrate。队列和定时任务保留 **Asynq** 方案，第三阶段按真实任务接入，不预先引入空 worker 或未使用的运行时依赖。

[ai-context](.ai-context/go-zero/00-instructions.md) 与 [zero-skills](.agents/skills/zero-skills/SKILL.md) 已按固定提交接入项目，入口为 [AGENTS.md](AGENTS.md)，不需要 mcp-zero。

| 文档 | 内容 |
| --- | --- |
| [整体架构](docs/architecture.md) | 服务、数据、权限与异步任务的目标设计 |
| [目录规范](docs/directory-structure.md) | 依赖边界与生成文件所有权 |
| [开发约定](docs/development-guide.md) | 契约、错误、开发与检查流程 |
| [实施路线](docs/roadmap.md) | 当前进度与后续阶段 |
| [字典与系统配置](docs/system-settings.md) | 参数类型、字典项、权限、审计与编辑冲突 |
| [操作日志](docs/operation-audit.md) | 日志范围、事务一致性、查询与脱敏 |
| [角色授权](docs/authorization.md) | 管理员、角色、菜单按钮、版本一致性与验证 |
| [管理员认证](docs/authentication.md) | 初始化、会话、Cookie、接口和验收 |
| [权限阶段验证](docs/phase-two-validation.md) | 本批检查、权限边界与浏览器验收 |
| [第一阶段验证](docs/phase-one-validation.md) | 已执行的检查与当前边界 |
| [AI 开发配置](docs/ai-development.md) | 官方资料版本、发现与升级方式 |
| [管理 API](apps/admin/api/README.md) / [业务 API](apps/app/api/README.md) / [系统服务](services/system/README.md) | 运行和接口说明 |
| [管理端](web/admin/README.md) | 页面组织和前端命令 |
