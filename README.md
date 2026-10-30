# GoSeed

[![CI](https://github.com/kar1hsu/go-seed/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/kar1hsu/go-seed/actions/workflows/ci.yaml)

GoSeed 是基于 **go-zero** 的 Go 项目开发脚手架，提供管理端 API、C 端 API、领域 RPC 和 Vue 3 管理后台。采用单仓库、单 Go module，API、RPC、队列消费者和定时调度独立运行，方便在已有基础能力上开发新业务。

后端使用 **GORM + MySQL、Redis、Casbin、Asynq**，管理后台使用 **Vue 3 + TypeScript + Element Plus**。项目包含可运行的会员业务样板、契约生成脚本、Docker Compose 和 CI。

[快速启动](#快速启动) · [本地开发](#本地开发) · [架构与目录](#架构与目录) · [开发新模块](docs/new-module-guide.md) · [文档导航](#文档导航)

## 已实现功能

| 功能 | 内容 |
| --- | --- |
| 管理员认证 | Ed25519 JWT、刷新 Cookie 轮换、退出登录、修改密码与会话撤销 |
| 权限管理 | 管理员增改、启停与角色分配；Casbin RBAC、菜单与按钮授权、服务端权限校验 |
| 操作日志 | 管理员、角色、字典、配置等变更的事务审计；按操作人、对象、日期和 Trace ID 查询 |
| 字典与配置 | 字典分类和子项；文本、整数、布尔、JSON 配置；编辑冲突检测 |
| 私有文件 | 本地或 S3 兼容存储；上传、下载、图片/PDF/文本预览；按配置时区生成日期目录与随机文件名 |
| 队列与定时任务 | Asynq worker/scheduler、事务 outbox、字典 CSV 导出、文件删除、定时清理和失败重试 |
| C 端会员 | 账号密码注册登录、个人资料、改密与会话；后台查询和启停；独立 RPC、数据库账号、密钥与 Cookie |
| 管理后台 | 按权限显示菜单和按钮、可折叠菜单分组、多页签与右键操作 |
| 工程支持 | 显式 SQL 迁移、健康与就绪检查、OpenAPI 和 TypeScript 类型生成、集成测试及容器运行检查 |

会员模块可作为新业务的开发参考。当前提供 C 端 API 和后台会员管理页面，尚未提供 C 端前台页面、短信/邮件登录或模块自动生成器；后续规划见[实施路线](docs/roadmap.md)。

## 快速启动

使用完整容器环境体验项目，需要 Git、make、Python **3.10+**、Docker 和 Compose **v2+**，无需在主机安装 Go 或 Node。

```bash
git clone --recurse-submodules https://github.com/kar1hsu/go-seed.git
cd go-seed
make init                          # 创建 .env，生成独立认证密钥与随机初始密码
make compose-up                    # 构建并启动全部服务及 MySQL / Redis
make compose-migrate               # 执行 system 和 member 数据库迁移
make compose-bootstrap             # 初始化管理员与超级管理员角色
```

打开 [管理后台](http://127.0.0.1:8088)。初始账号为 `admin`，初始密码保存在本地 `.env` 的 `ADMIN_BOOTSTRAP_PASSWORD`，登录后可在账户菜单修改密码。项目没有固定的默认密码。

`make init` 保留已有配置、密钥和密码；没有本机 Go 时会使用固定版本的构建容器生成密钥。已有 checkout 可执行 `git submodule update --init --recursive` 恢复 AI 参考资料。

```bash
make smoke                         # 验证正在运行的 API / RPC / 数据库 / 缓存链路
make compose-down                  # 停止并移除容器，保留命名卷中的数据
```

Compose 提供同源入口，RPC 端口不发布到宿主机。它用于开发和部署参考；生产环境需要按[生产配置](docs/production.md)设置 `GOSEED_MODE=pro`、TLS、Secure Cookie、允许来源和密钥管理。

## 本地开发

先克隆仓库，再准备以下工具。版本以 [tools/versions.env](tools/versions.env)、[.nvmrc](.nvmrc) 和 [package.json](web/admin/package.json) 为准。

| 工具 | 版本 |
| --- | --- |
| Go | 1.25.11 |
| Node.js | 24.19.0 |
| pnpm | 11.15.1 |
| Python | 3.10+ |
| Docker Compose | v2+ |

还需要 Git、curl、unzip 和 make。使用 nvm 时先执行 `nvm install`、`nvm use`；pnpm 按[官方安装说明](https://pnpm.io/installation)安装指定版本。

本地开发与完整容器环境使用相同的 API 端口。如果已启动完整容器环境，先执行 `make compose-down`。

```bash
make init
make tools                         # 固定生成工具安装到项目 .tools/bin
pnpm --dir web/admin install --frozen-lockfile
make doctor
make infra-up                      # 启动项目独立的 MySQL / Redis，准备 member 数据库账号
make migrate                       # 执行 system 和 member 数据库迁移
make bootstrap                     # 幂等初始化管理员，已有密码保持不变
make dev                           # 启动六个 Go 进程和 Vite，Ctrl+C 统一退出
```

打开 [本地管理后台](http://127.0.0.1:5173)，登录凭证与容器模式相同。另一个终端可执行 `make smoke` 检查完整链路。

| 服务 | 本地默认地址 |
| --- | --- |
| 管理后台 | `http://127.0.0.1:5173` |
| admin-api / app-api | `http://127.0.0.1:8081` / `http://127.0.0.1:8080` |
| system-rpc / member-rpc | `127.0.0.1:19001` / `127.0.0.1:19002` |
| worker / scheduler 就绪检查 | `http://127.0.0.1:19201/readyz` / `http://127.0.0.1:19202/readyz` |
| MySQL / Redis | `127.0.0.1:13306` / `127.0.0.1:16379` |

端口和开发凭证在 `.env` 中配置，不要提交该文件。本地服务默认只监听回环地址；`make infra-down` 停止 MySQL 和 Redis，保留数据。

## 架构与目录

请求按 `HTTP API → 领域 RPC → application → repository` 流转。API 负责协议、身份和入口校验，通过 RPC 调用业务；领域内部封装用例、持久化与事务，跨领域通过公开契约协作。

```text
apps/
  admin/api/                  管理端 HTTP 入口
  app/api/                    C 端 HTTP 入口
services/
  system/
    rpc/                      系统领域 RPC、契约与客户端
    internal/                 系统用例、仓储与模型，供 RPC / worker / 命令共享
    worker/                   Asynq 消费、outbox 投递
    scheduler/                定时清理任务投递
    migrations/               系统数据库迁移
  member/
    rpc/                      会员领域 RPC、契约与客户端
    internal/                 会员账号、会话与管理用例
    migrations/               会员数据库迁移
web/admin/                    Vue 管理端，按 feature 组织
pkg/                          公共技术包，不放业务实体
cmd/probe/                    容器健康检查客户端
deploy/                       Dockerfile、Compose 与 Nginx 配置
scripts/                      生成、检查、进程管理与集成验证
docs/                         架构、开发、部署与业务文档
```

- **单 Go module**：`github.com/kar1hsu/go-seed`，六个服务独立构建和运行；子目录不另建 `go.mod`。
- **领域隔离**：system 与 member 使用独立数据库和账号，各进程通过 ServiceContext 显式装配依赖；API 不直接访问业务表。数据库结构通过 SQL 迁移维护，不在启动时执行 AutoMigrate。
- **身份隔离**：管理员与会员使用独立的 JWT 标识、Ed25519 密钥、刷新密钥和 Cookie；领域 RPC 持有签名私钥，API 只持有验签公钥。
- **异步可靠性**：任务经事务 outbox 投递，按至少一次执行设计，通过数据库状态与租约处理幂等和恢复。

更多边界和生成文件归属见[整体架构](docs/architecture.md)与[目录规范](docs/directory-structure.md)。

## 存储与时区

文件默认存放在本地 `.data/files`；Compose 使用 `files-data` 命名卷，也可切换为 S3 兼容存储。文件保持私有，通过鉴权接口访问；对象路径为 `YYYYMMDD/随机文件名`，原始名称保留在元数据中。详见[文件与任务](docs/files-and-tasks.md)。

统一时区由 `.env` 的 `TZ` 配置，默认 `Asia/Shanghai`，支持 IANA 时区名称。日期目录和导出文件名按该时区生成，数据库连接与接口时间仍使用 UTC。修改后重启本地进程，或执行 `make compose-up` 更新容器；已有文件路径不变。

Go 运行镜像采用固定 digest 的 **Distroless Debian 13 nonroot**，包含 CA 和时区数据；健康检查与领域命令直接执行二进制。

## 常用命令

| 命令 | 用途 |
| --- | --- |
| `make check` | 格式、依赖边界、文档链接、vet、前端类型/格式和生成一致性检查 |
| `make test` / `make build` | 无外部服务依赖的测试 / 六个服务、probe 与管理端构建 |
| `make smoke` | 检查已运行服务的 API、RPC、MySQL、Redis 链路 |
| `make test-auth` / `make auth-smoke` | MySQL 认证并发测试 / HTTP 会话链路 |
| `make test-rbac` / `make rbac-smoke` | 权限并发与 worker 退出测试 / HTTP 管理链路 |
| `make test-member` / `make member-smoke` | 会员并发、回滚测试 / 双 API、双 RPC 会话与管理链路 |
| `make test-storage` | 本地存储与 S3 协议检查，可选独立 S3 实服务测试 |
| `make test-runtime` | Distroless 用户、CA、时区、文件卷和二进制执行检查；需要 Go 和 Docker |
| `make sync-resources` | 同步权限目录，保留角色授权和成员关系 |
| `make gen` | 生成 API、RPC、OpenAPI 3、管理端类型和菜单资源 |
| `make gen-api APP=admin` | 从 `.api` 生成指定入口，保留手写实现 |
| `make gen-rpc DOMAIN=system` | 从 `.proto` 生成指定领域 RPC |
| `make gen-web` | 校验权限和菜单声明，更新 OpenAPI、前端类型与菜单资源 |
| `make verify-gen` | 在临时副本检查重复生成一致性，以及新增契约对手写实现的保留 |

数据库集成测试需要本项目的 MySQL，HTTP smoke 需要对应服务已经启动；具体准备步骤见[开发约定](docs/development-guide.md)。构建产物位于 `bin/` 和 `web/admin/dist/`，契约、生成文件及 lockfile 进入版本控制。CI 执行检查、构建、race、数据库集成测试、HTTP smoke 和运行镜像验证。

## AI 开发

官方 [ai-context](.ai-context/go-zero/00-instructions.md) 与 [zero-skills](.agents/skills/zero-skills/SKILL.md) 已通过固定提交的 Git 子模块接入。AI 助手先读 [AGENTS.md](AGENTS.md)，再按任务查阅工作流与实现参考；支持终端的助手直接运行项目固定版本的 goctl，无需启用 mcp-zero。

这些资料不是运行时依赖。版本、加载方式和更新流程见 [AI 开发配置](docs/ai-development.md)。

## 文档导航

| 主题 | 文档 |
| --- | --- |
| 架构与开发 | [整体架构](docs/architecture.md) · [目录规范](docs/directory-structure.md) · [开发约定](docs/development-guide.md) · [实施路线](docs/roadmap.md) |
| 开发新业务 | [新增模块教程](docs/new-module-guide.md) · [会员业务样板](docs/member-module.md) |
| 身份与权限 | [管理员认证](docs/authentication.md) · [角色授权](docs/authorization.md) · [操作日志](docs/operation-audit.md) |
| 配置与任务 | [字典与系统配置](docs/system-settings.md) · [文件与任务](docs/files-and-tasks.md) |
| 部署与工具 | [生产配置](docs/production.md) · [AI 开发配置](docs/ai-development.md) |
| HTTP 接口 | [管理 API](apps/admin/api/README.md) · [业务 API](apps/app/api/README.md) · [管理端 OpenAPI](docs/openapi/admin.openapi.json) · [C 端 OpenAPI](docs/openapi/app.openapi.json) |
| 服务与前端 | [系统服务](services/system/README.md) · [会员服务](services/member/README.md) · [管理端](web/admin/README.md) |

<details>
<summary>阶段验证记录</summary>

- [可靠性修复验证](docs/reliability-validation.md)
- [会员样板验证](docs/member-validation.md)
- [文件与任务验证](docs/phase-three-validation.md)
- [权限阶段验证](docs/phase-two-validation.md)
- [第一阶段验证](docs/phase-one-validation.md)

</details>
