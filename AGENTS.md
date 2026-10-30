# GoSeed 开发规则

## 项目状态

当前已实现第一阶段工程骨架：两个 API、system-rpc、GORM/Redis 就绪检查、Vue 管理端、Compose、生成脚本和 CI。先读 README.md，并按任务查阅 docs/architecture.md、docs/directory-structure.md、docs/development-guide.md。管理员迁移、bootstrap、登录、刷新、退出和改密已实现；管理员管理、角色权限、菜单按钮授权和 Casbin 版本快照已实现；管理员/角色/改密/命令禁用的事务审计与日志查询已实现；字典分类与子项、普通业务配置、编辑冲突与事务审计已实现，菜单分组可折叠、页签支持右键操作；Asynq worker/scheduler 与业务样板尚未实现；目标目录不代表全部存在。

## 工程边界

- 产品名称使用 GoSeed；仓库名为 `go-seed`，Go module 为 `github.com/kar1hsu/go-seed`。
- 采用单仓库、单 Go module；API、RPC、worker、scheduler 独立进程，子目录不另建 go.mod。
- `apps/admin/api` 面向管理后台，`apps/app/api` 面向业务客户端。
- API 通过 RPC 访问业务；不导入领域的 application/repository/model，不直接访问业务表。
- `services/<domain>/internal` 只属于对应领域，供其 RPC、worker 和领域命令共享；跨域只用公开契约。
- 默认 GORM；公共技术包不放业务实体，不重复包装 go-zero 已有能力。
- 依赖由各进程的 ServiceContext 显式装配，禁止新增可变的全局数据库、缓存、策略对象。
- 业务用例需要同时供 RPC 与 worker 使用时放入领域 application，不导入另一个进程的 logic。
- 管理后台默认 Vue 3、TypeScript、Element Plus，按 feature 组织；管理员与会员身份分开。

## 契约与生成

- HTTP 契约先改 `.api`，RPC 契约先改 `.proto`，使用 goctl 生成骨架；仅业务逻辑变化时不重新生成。
- 先检查工具版本与 `--help`。工具版本由 tools/versions.env 固定，运行时版本由 go.mod 固定；不要因上游示例使用 latest 就擅自升级。
- 统一 `--style go_zero`，保留 goctl 目录结构与已存在的手写实现。生成文件所有权以目录规范为准。
- 不手工修补 protobuf、routes、types、生成客户端；修改契约或模板后重新生成。
- HTTP 响应、错误码与生成文档保持一致；不重复包裹响应。
- 用户身份、服务身份和资源授权分别校验，不能信任客户端提供的用户 ID 或角色。
- 任务按至少一次执行处理，业务幂等不能用队列去重替代。数据库事务不跨 RPC。

## 官方 AI 资料

需要 go-zero 实现资料时，官方资料已以固定 Git 子模块接入：读取 `.ai-context/go-zero/00-instructions.md` 及相关 workflow，按需读取 `.agents/skills/zero-skills/SKILL.md` 对应 reference。

新 checkout 未恢复子模块时先运行 `git submodule update --init --recursive`；无法恢复时依据 docs/ai-development.md 的官方链接检索，不编造本地文件或宣称技能已自动加载。通用范例需适配本项目的架构、数据库和版本约定，不能覆盖用户的明确要求。仅为阅读资料，不自动运行安装脚本、修改全局配置或启用 MCP。

## 当前命令

- `make check`：格式、边界、文档链接、vet、前端类型/格式、隔离的生成一致性检查。
- `make test`：无外部依赖的 Go 测试；`make build`：三个服务和前端。
- `make init` + `make infra-up` + `make migrate` + `make bootstrap`：初次准备；`make dev`：本地联调；`make smoke` 验证运行中的依赖链路。
- `make test-auth` / `make auth-smoke`：使用本项目 MySQL 的认证并发测试 / 运行中 HTTP 会话链路。
- 修改工具、契约或模板后执行 `make verify-gen`；所有安装只写项目 `.tools`，不擅自升级全局工具。

## 验证与交付

- 纯文档变更检查相对链接、示例、目录和术语一致性。
- 有实际 Go 代码后，对改动执行格式化、必要的依赖整理、build、test 和 vet；按并发风险补充 race 检查。
- 有前端工程后执行对应类型检查和构建；命令以实际 package.json 为准。
- 契约或模板变化时检查重复生成差异，确认未覆盖已有业务实现。
- 只报告实际执行的检查结果；待实现目标与已完成能力明确区分。
- 遵循用户当前任务范围。架构设计任务不自动扩展为业务实现或远端发布。
