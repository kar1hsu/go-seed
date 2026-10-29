# GoSeed AI 开发配置

GoSeed 适合采用“项目规则 + 官方工作流 + 按需参考 + goctl 命令行”的 AI 开发方式。当前 Codex 可以直接执行终端命令，因此不需要额外部署 mcp-zero 来完成代码生成。

本文依据 2026 年 10 月 29 日读取的官方资料。项目 `AGENTS.md` 与两个官方参考库均已接入；资料采用 Git 子模块固定提交，没有修改全局 Codex 配置或启用 MCP。

## 三个官方项目的取舍

| 项目 | 作用 | GoSeed 的选择 |
| --- | --- | --- |
| [ai-context](https://github.com/zeromicro/ai-context) | 定义契约优先、代码生成、实现及验证的工作流程 | 已项目级接入，由 AGENTS.md 引用，保留 GoSeed 自己的架构规则 |
| [zero-skills](https://github.com/zeromicro/zero-skills) | REST、RPC、数据库、错误处理等参考范例，按任务读取 | 已安装到项目 `.agents/skills/zero-skills`，固定版本 |
| [mcp-zero](https://github.com/zeromicro/mcp-zero) | 将生成、分析等能力暴露为 MCP 工具，仍需 goctl | 当前不需要；将来使用不能运行终端但支持 MCP 的客户端时再评估 |

ai-context 的官方流程直接在终端调用 goctl；仅下载文档不会激活规则，还需要对应编辑器的入口。[ai-context 安装与验证](https://github.com/zeromicro/ai-context#manual-setup)

zero-skills 的 Codex 指南采用 `.agents/skills` 原生发现。它与 `AGENTS.md` 职责不同：前者提供可复用技术资料，后者记录本项目的边界和命令。[zero-skills Codex 指南](https://github.com/zeromicro/zero-skills/blob/main/getting-started/codex-guide.md)

## 项目级配置结构

```text
go-seed/
├── AGENTS.md                         # 已创建，简明的项目规则
├── docs/ai-development.md            # 本文
├── .ai-context/
│   └── go-zero/                      # ai-context 固定提交
│       ├── 00-instructions.md
│       ├── workflows.md
│       ├── tools.md
│       └── patterns.md
└── .agents/skills/
    └── zero-skills/                   # 官方 skill 固定提交
        ├── SKILL.md
        └── references/
```

不要同时把 zero-skills 安装到 `.ai-context/zero-skills` 和 `.agents/skills/zero-skills`。GoSeed 采用后一条路径，所有项目引用随之调整。

Codex 从仓库根目录到工作目录读取 AGENTS.md，并在项目层发现 `.agents/skills` 中的 skills。这个结构无需额外添加 `.codex/config.toml` 或修改全局指令。[OpenAI AGENTS.md 文档](https://developers.openai.com/codex/guides/agents-md/)、[OpenAI Skills 文档](https://developers.openai.com/codex/skills/)

## 已完成的接入

已从仓库根目录执行以下命令：

```bash
git submodule add https://github.com/zeromicro/ai-context.git .ai-context/go-zero
git submodule add https://github.com/zeromicro/zero-skills.git .agents/skills/zero-skills
```

使用 submodule 固定到当次审查的提交。提交时将 `.gitmodules`、两个 gitlink 与 AGENTS.md 一起纳入版本控制；不要只提交路径或忽略子模块配置。其他开发者克隆后执行：

```bash
git submodule update --init --recursive
```

固定提交：ai-context `6ad8a5f1ae0c99ea8e3511994f93749fd0acba5e`；zero-skills `f1efca09deff5e9c360f9eb115f63769495931b1`。本仓库 AGENTS.md 已引用这两个实际路径。不要直接用官方适配文件覆盖项目 AGENTS.md，否则会丢失 API/RPC 数据边界、GORM 选择及生成文件所有权等项目约定。

升级时显式选择并审查新的上游提交，再更新子模块指针；日常开发不自动跟随 main，也不把更新官方文档作为构建业务服务的步骤。CI 构建运行时代码不应依赖外网读取 AI 资料。

## 版本与项目约定

当前 ai-context README 声明的验证基线是 Go 1.26.3、goctl/go-zero 1.9.2；读取到的 zero-skills SKILL.md 则声明 go-zero 1.10.3 为稳定目标。它们不是同一套版本锁定配置，也不是要求本项目自动升级或降级。[ai-context Requirements](https://github.com/zeromicro/ai-context#requirements)、[zero-skills 版本说明](https://github.com/zeromicro/zero-skills/blob/main/SKILL.md)

第一阶段已固定并验证 Go 1.25.11、go-zero/goctl 1.10.2、protoc 29.1、protoc-gen-go 1.36.5、protoc-gen-go-grpc 1.5.1。运行时在 go.mod，工具在 tools/versions.env，生成模板随仓库保存。

上游示例提供通用模式，GoSeed 有以下具体约定：

- API 只做客户端协议适配与编排，业务数据访问属于对应 RPC 领域。
- 同一领域的 RPC 和 worker 共用 application；不把可复用业务全部塞进 RPC logic。
- 默认 GORM，不因为参考文档展示 `goctl model` 就切换到 sqlx 或生成第二套模型。
- 生成文件、可手写骨架和派生前端代码按[目录规范](directory-structure.md)区分。
- GoSeed 的响应、会话撤销、权限和审计规则优先；示例不自动满足这些规则。

对任何从资料复制的 API 签名、配置字段或生成参数，都先核对固定版本和本机 `--help`，再进行最小生成或编译验证。框架资料不能代替项目测试。

## 验证方式

1. 检查项目 AGENTS.md 与文档路径可以读取。
2. 接入子模块后，确认 `00-instructions.md` 与 `SKILL.md` 位于预期位置，且 gitlink 固定到可审查提交。
3. 开启一个新的 Codex 会话，核对项目规则的发现情况。CLI/IDE 可使用 `/skills` 查看 `zero-skills`；桌面端可在 Skills 界面或技能选择器中核对，未出现时重启对应客户端。
4. 明确调用 `$zero-skills` 并读取一个相关 reference，确认文件可访问；显式读取成功并不等于已经证明自动发现。
5. 在临时目录用固定工具链生成最小 API/RPC，确认模块路径、生成目录与重复生成行为，然后再用于项目代码。

已完成子模块路径/提交核对、实际读取 workflow 与 REST/RPC/goctl references、工程生成与编译。自动发现仍需在下一会话或技能选择器确认；本次显式读取成功不等同于验证未来会话自动加载。未启动 MCP 服务。

## 其他编辑器

后续确实使用 Claude Code、Cursor 或 Copilot 时，再增加相应的薄入口文件，共享同一套项目规则和参考目录。保持一份 GoSeed 规则来源，不在每个编辑器配置中复制完整架构。当前不预设其他编辑器或写入它们的全局配置。
