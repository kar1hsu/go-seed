# admin-api 接口

契约：[desc/admin.api](desc/admin.api)。生成文档：[OpenAPI 3](../../../docs/openapi/admin.openapi.json)。

## GET /admin-api/v1/system/ping

无参数、无认证要求，仅用于第一阶段非敏感通信探测。调用 `goseed.system.v1.System/Ping`，不会访问业务表。

```bash
curl -i http://127.0.0.1:8081/admin-api/v1/system/ping
```

成功示例（time 和 traceId 每次不同）：

```json
{"code":0,"message":"ok","data":{"service":"system-rpc","version":"0.1.0","time":"2026-10-29T00:00:00Z"},"traceId":"736dc7b9d89893a1e9ad965b6d7e65e4"}
```

依赖连接失败返回 HTTP 503 / code 100503；RPC 超时返回 HTTP 504 / code 100504；未知内部失败返回 HTTP 500 / code 100500。错误响应 `data` 为 null，完整内部原因只记日志。通用状态映射见[开发约定](../../../docs/development-guide.md)。

```json
{"code":100503,"message":"service unavailable","data":null,"traceId":"736dc7b9d89893a1e9ad965b6d7e65e4"}
```

业务 handler 的完整成功类型由 `.api` 定义，错误 schema 由 `web/admin/scripts/generate-api.mjs` 合并进 OpenAPI。框架级未匹配路由/方法与健康检查采用各自 HTTP 协议；健康探测不属于业务 OpenAPI。

Ping 成功只表示 API/RPC 可达，数据库状态请检查 `GET /readyz`。未来敏感接口必须完成用户身份、服务身份和资源授权后才开放。

## 管理员认证

认证接口前缀 `/admin-api/v1/auth`。详细请求、Cookie 和撤销语义见 [管理员认证](../../../docs/authentication.md)，OpenAPI 见 [admin.openapi.json](../../../docs/openapi/admin.openapi.json)。

| 方法 | 路径 | 输入 / 输出 |
| --- | --- | --- |
| POST | `/login` | username/password；返回 access token、到期时间和管理员信息，设置 refresh Cookie |
| POST | `/refresh` | refresh Cookie；原子轮换 Cookie 与 access token |
| POST | `/logout` | refresh Cookie；撤销当前会话并清除 Cookie |
| GET | `/me` | Bearer access token；返回经过会话校验的管理员信息 |
| POST | `/password` | Bearer access token、currentPassword/newPassword；改密并撤销全部会话 |

所有认证 POST 要求 `X-GoSeed-CSRF: 1`。浏览器 Origin 必须在配置白名单内。身份失败返回 HTTP 401/code 100401；字段或密码规则不符合要求返回 400/100400；来源拒绝返回 403/100403；登录频率超限返回 429/100429。认证响应使用 `Cache-Control: no-store`。
