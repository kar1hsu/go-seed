# 第一阶段验证记录

日期：2026-10-29。环境：macOS arm64，Go 1.25.11、goctl 1.10.2、protoc 29.1、Node 24.19.0、pnpm 11.15.1，Docker Engine 29.8.2 / Compose 5.5.0。

## 实际执行结果

| 检查 | 结果 |
| --- | --- |
| AI 子模块路径、提交和文件读取 | 两个固定提交可读；已读取 workflow、SKILL.md 和相关 references |
| `make tools` / `make doctor` | 项目级工具安装与版本检查通过 |
| `go build ./...`，`make dev` 中的四个二进制构建 | 通过，三个服务和 probe 可独立构建 |
| `go test ./...` | 通过，覆盖公共错误映射、依赖失败/取消、两个生成 handler 到内存 gRPC 的链路 |
| `go test -race ./pkg/... ./apps/... ./services/...` | 通过；macOS linker 提示 LC_DYSYMTAB warning，未出现 race 失败 |
| `make check` | 格式、包边界、文档链接、vet、前端类型/格式、重复生成检查通过 |
| `make verify-gen` | 重复生成无差异；临时新增 API/RPC 方法后原有手写实现不变且编译通过 |
| `pnpm --dir web/admin build` | 通过，Element Plus 按组件引入，最终构建无大 chunk 告警 |
| 本地 `make infra-up` + `make dev` + `make smoke` | 两条 API → RPC 链路、MySQL/Redis readiness 和 Trace ID 通过 |
| Compose 镜像构建、`up -d --wait`、`make smoke` | 通过，MySQL、Redis、RPC、两个 API 与 Nginx 共六个容器健康 |
| Redis 故障/恢复 | 仅停止 GoSeed Redis：healthz 200、readyz 503、Ping 200；恢复后 readyz 200 |
| RPC 故障/恢复 | 仅停止 GoSeed RPC：两个 API healthz 200、readyz 503、业务探测 HTTP 503/code 100503；恢复后自动重连 |
| 浏览器验收 | 容器静态页面、真实状态、刷新、故障提示和恢复均验证，未使用虚构业务数据 |
| Compose 配置与 Git whitespace 检查 | 通过 |
| [GitHub Actions CI](https://github.com/kar1hsu/go-seed/actions/runs/40882304709) | 第一阶段 `main` 提交通过 |

生成器的 `x-date` 时间戳在文档生成脚本中移除，使输出可重复。Smoke 对启动/重连最多等待 30 秒，不将短暂重连差异当作最终成功或失败；超过上限仍报错。

## 管理端验收

- 管理端采用参考 Vben Admin 浅色主题的布局，保留 Vue 3 + Element Plus。浏览器验证四个页面、侧栏折叠、多页签关闭、实际请求、Trace ID/结果筛选、清空与重查、详情抽屉。
- Compose gateway 的 `8088` 静态页面、同源 API 和 `/monitor` 深层路由刷新验证通过；控制台未记录 error/warn。检查记录来自实际请求，只保留在页面内存中。

## 尚未执行或不属于本阶段

- 未来 Codex 会话的自动技能发现尚需重载会话确认；本次显式读取成功不代表已经验证自动发现。
- 管理员认证、服务身份和认证表迁移的实现范围见 [管理员认证](authentication.md)。RBAC、业务表、Asynq worker/scheduler、文件业务和商品样板留在后续阶段。
- 当前 Compose 是回环端口的开发参考环境，不作为生产安全验收。
