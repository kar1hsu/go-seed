# 生产配置

生产进程设置 `GOSEED_MODE=pro`；本地 `.env` 和 Compose 默认为 `dev`。pro 使用 go-zero 生产降载配置，system-rpc 不注册 reflection。公开健康接口仍仅暴露存活/就绪状态；业务 RPC 继续逐方法验证服务凭证和用户身份。

两个 API 在 pro 启动时强制要求对应 `ADMIN_COOKIE_SECURE=true` / `MEMBER_COOKIE_SECURE=true`，允许来源必须是完整 HTTPS Origin。HTTP 由外层可信 HTTPS 网关终止 TLS；网关到 API 应置于受控网络。禁止把任意私网网段整体当作可信代理。`TRUSTED_PROXY_CIDRS` 只填写真实直接代理的 IP/CIDR；解析从右向左越过已信任跳点，遇到首个不信任地址即停止。

RPC 的配置：

| 变量 | 用途 |
| --- | --- |
| `RPC_TLS_CA` | API、member-rpc 及 probe 信任的 PEM CA 文件路径 |
| `SYSTEM_RPC_TLS_NAME` / `MEMBER_RPC_TLS_NAME` | 证书 SAN 中对应服务的主机名 |
| `SYSTEM_RPC_TLS_CERT` / `SYSTEM_RPC_TLS_KEY` | system-rpc 服务端 PEM 证书链和私钥路径 |
| `MEMBER_RPC_TLS_CERT` / `MEMBER_RPC_TLS_KEY` | member-rpc 服务端 PEM 证书链和私钥路径 |

TLS 最低 1.2，客户端验证 CA 与主机名，不提供跳过验证选项；pro 禁止明文 RPC。TLS 保护传输，服务凭证白名单仍负责调用方授权。证书由部署系统签发/轮换并只读挂载；参考 Compose 透传配置，但不捆绑证书、HTTPS 域名或生产密钥。使用独立 Compose override 或编排平台挂载证书到上表路径，API 只挂 CA，RPC 挂各自证书和私钥。pro 的 probe 也使用相同 CA 和对应服务名。

示例配置（路径为容器内路径，证书由部署系统提供）：

```dotenv
GOSEED_MODE=pro
ADMIN_COOKIE_SECURE=true
MEMBER_COOKIE_SECURE=true
ADMIN_ALLOWED_ORIGINS=https://admin.example.com
MEMBER_ALLOWED_ORIGINS=https://app.example.com
RPC_TLS_CA=/run/secrets/rpc-ca.pem
SYSTEM_RPC_TLS_NAME=system-rpc
MEMBER_RPC_TLS_NAME=member-rpc
SYSTEM_RPC_TLS_CERT=/run/secrets/system.crt
SYSTEM_RPC_TLS_KEY=/run/secrets/system.key
MEMBER_RPC_TLS_CERT=/run/secrets/member.crt
MEMBER_RPC_TLS_KEY=/run/secrets/member.key
```

Compose 的 gateway 使用独立 ingress 网络 `172.30.86.0/28`、固定地址 `172.30.86.2`（动态地址池限定为 `172.30.86.8/29`，避免 API 抢占 gateway 地址），API 只信任这个地址。部署网段冲突时同时调整网络和 API 信任配置。gateway 不与数据库共享网络。直接暴露 API 端口仍只绑定回环地址。

Nginx 参考配置包含 CSP、nosniff、Referrer-Policy 和禁止被嵌入的响应头。Element Plus 运行时样式需要 `style-src 'unsafe-inline'`；脚本仍限定同源。仅在真实 HTTPS 入口添加 `Strict-Transport-Security`，例如 `max-age=31536000`；本地 HTTP 示例不发送 HSTS。多级代理应在边缘覆盖不可信的转发头，并明确每一跳的信任关系。

首次 `make init` 为 MySQL system/root/member 账号、认证与 RPC 生成随机凭证；保留已有 `.env`，不会自动更换正在使用的数据库密码。worker 不需要管理员 JWT 密钥。生产使用独立密码/服务密钥、数据库最小权限、受控 Redis 与私有对象存储。

升级本次可靠性变更需先应用 system `00006`、`00007` 和 member `00002` 迁移，再更新服务。迁移进程使用独立连接，取消 2 秒读超时，由 10 分钟命令上下文限定 DDL；不自动回滚/截断审计历史。SQL 的 Down 仅用于可容纳原有列宽的数据，生产回滚前须检查 64 字符以上的审计名称。

`AUDIT_RETENTION_DAYS=0` 保留所有历史。设置非零值前先制定归档策略；系统不会替运维选择保留期。迁移、部署、备份和证书管理仍由发布流程负责。

## 服务时区

所有容器和 Go 进程统一读取 `TZ`（默认 `Asia/Shanghai`），使用 IANA 名称并随环境配置发布。Go 服务、worker、scheduler、probe 和领域命令均在启动时校验，空值和无效时区直接报错；不运行中修改 `time.Local`。Distroless 服务镜像包含 tzdata，Go 二进制也内嵌 IANA 数据，避免精简系统缺少时区。Compose 同时向 Nginx、MySQL 和 Redis 传递该值；固定版本的 MySQL/Redis 镜像自带所需时区数据。

业务按日分组（文件目录、导出文件名）使用配置的日历时区。MySQL 系统时区跟随容器，应用及迁移连接显式设置 `time_zone=+00:00`，保持数据库的 NOW()、时间默认值、驱动读写和 UTC 接口契约一致。小时清理任务继续以 UTC 时间桶作为幂等标识。变更 TZ 后统一重建容器/重启进程；已有文件对象键和存储时间不重写。

## 身份密钥

所有值采用不带 padding 的 base64url。Ed25519 公钥为 32 字节、私钥为 64 字节（Go 的 seed + public 格式），刷新恢复 HMAC 密钥为独立的 32 字节。启动时拒绝错误长度、编码、不配对的私钥/公钥，以及拿公钥或私钥 seed 充当刷新密钥的配置。

| 进程 | JWT 配置 | 刷新恢复配置 |
| --- | --- | --- |
| system-rpc | `ADMIN_TOKEN_PRIVATE_KEY`、`ADMIN_TOKEN_PUBLIC_KEY` | `ADMIN_REFRESH_KEY` |
| admin-api | `ADMIN_TOKEN_PUBLIC_KEY` | 无 |
| member-rpc | `MEMBER_TOKEN_PRIVATE_KEY`、`MEMBER_TOKEN_PUBLIC_KEY` | `MEMBER_REFRESH_KEY` |
| app-api | `MEMBER_TOKEN_PUBLIC_KEY` | 无 |
| worker、scheduler、领域管理命令、管理端 Vite、Nginx | 无 | 无 |

`make init` 生成缺失的两套密钥，校验并保留已有值；不能只保留公钥再静默替换私钥。初始化还会拒绝跨域共用 JWT 密钥、共用刷新密钥，以及把任一身份域的公钥或私钥 seed 填入刷新密钥的情况；生产自行注入时也应保持这些值独立。初始化程序使用标准库 `crypto/ed25519`，本机无 Go 时使用固定 Go 构建容器，敏感值只经 stdin/stdout 捕获，不出现在命令行或初始化日志。`.env` 原子写入且权限为 0600。已有 `.env` 中的 `ADMIN_TOKEN_SECRET` / `MEMBER_TOKEN_SECRET` 仅作为旧配置保留，服务已不再读取，也不会传给本地子进程或 Compose 服务。

本地 `make dev` 按各服务 YAML 的变量引用筛选环境，Vite 仅接收端口等必要配置；Compose 分别注入上表密钥。生产应由部署系统注入并限制配置读取权限，不把完整 `.env` 注入所有容器。

密钥更新采用协调切换：先准备配对密钥，再同步更新签发 RPC 与对应 API 并重启；当前只有单个有效公钥，不提供多 `kid` 过渡或 HS256 降级兼容。只轮换 JWT 密钥时保留 `*_REFRESH_KEY`；有效刷新 Cookie 可换取新 access token。刷新密钥另行轮换会影响正在恢复的刷新请求，应避开恢复窗口或接受这些会话重新登录。

## Go 运行镜像

六个 Go 服务使用 `gcr.io/distroless/static-debian13:nonroot`，Dockerfile 固定多架构 manifest digest，保留 UID/GID 65532。构建阶段继续使用固定 Go 版本并 `CGO_ENABLED=0`；运行镜像只复制服务、probe、领域命令、配置文件和预设权限的数据目录，没有 shell、包管理器或编译工具。Nginx、MySQL、Redis 镜像不在此次替换范围内。

CA 证书与 tzdata 由 static 镜像提供，`TZ` 及 Go 内嵌时区数据继续生效。健康检查通过 exec 形式调用 `/app/probe`，`make compose-migrate`、`make compose-bootstrap` 直接执行领域二进制，不依赖 shell。上传目录 `/app/data/files` 预设为 65532 所有，命名卷沿用该权限；生产 bind mount 也需授予该 UID 写权限。

`make test-runtime` 构建 system-rpc 作为共享 Dockerfile 的代表镜像，在隔离网络、只读根文件系统和临时文件卷中检查 UID/GID、CA、三个时区、文件读写，以及服务、probe、领域命令的直接执行。该检查已接入 CI；完整业务联调仍使用 Compose 和各 smoke 命令。

诊断使用容器日志、probe 或单独的调试容器；运行镜像不支持 `docker exec ... sh/date/ls`。更新基础镜像时显式更新 digest，并重跑容器健康、迁移、认证、时区和文件写入检查。官方资料：[运行镜像内容](https://github.com/GoogleContainerTools/distroless/blob/main/base/README.md)、[维护中的镜像](https://github.com/GoogleContainerTools/distroless#what-images-are-available)。
