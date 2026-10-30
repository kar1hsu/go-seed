# 字典管理与系统配置

系统管理提供字典分类、字典项及普通业务参数的管理接口。后台菜单按分组展开/收起；页面标签支持右键刷新、关闭当前、其他、右侧及全部标签，工作台固定保留。右上角的标签操作按钮提供相同入口，键盘可在标签上使用 Shift+F10。

## 字典

每个字典包含不可变的 `code`、名称、说明、启停状态以及最多 200 个字典项。字典项的 `value` 在同一分类中唯一且区分大小写，`label` 为显示文字，`sort` 为 0–9999 的整数，读取时按 sort、value 排序。名称、值和标签最多 64 字，说明最多 256 字。

字典分类与所有字典项在一次本地事务中保存。更新请求提交完整的 items 数组，空数组表示清空；服务端校验分类 revision，再整体替换子项。另一编辑者先保存后，旧 revision 请求返回 409，不能静默覆盖。修改字典值等同于移除旧选项并新增选项，接入业务后需考虑已使用的值。

分类启停与字典项启停分别保存，管理接口可查看停用记录。分类停用后才允许删除，删除同时移除子项并保留历史审计。当前没有业务引用该数据；后续接入业务时由用例校验引用与有效状态，不将管理列表直接作为公开选项接口。

## 系统配置

配置键 `code` 与值类型创建后不可修改；键为最长 64 位的小写标识，允许点分段，如 `site.title`、`order.page_size`。配置值经 RPC 再次校验：

| 类型 | 存储和输入约定 |
| --- | --- |
| text | UTF-8 文本，允许空字符串及换行 |
| integer | 规范十进制的有符号 int64 字符串，不经 JavaScript Number 转换 |
| boolean | 字符串 `true` 或 `false` |
| json | 合法 JSON；移除无意义空白，保留大整数原始数值精度 |

所有配置值上限为 4096 字节。配置支持关键词、启停筛选与分页；编辑、删除携带 revision，删除前必须停用。配置读写不修改 Casbin 策略版本。

此模块用于普通业务参数，不是密钥存储。数据库密码、JWT secret、服务凭证等继续通过部署环境/密钥管理传入，不能保存到配置值或说明中；配置值会显示给有读取权限的管理员，并记录到审计历史。

当前实现管理与持久化，**不会自动覆盖进程 YAML、环境变量或现有页面行为**，也不向 app-api 公开配置。业务实际使用某个键时，应定义专门的读取契约、默认值、停用行为及更新生效规则，再按需求决定缓存。

## 接口与权限

所有 HTTP 接口以前缀 `/admin-api/v1/system` 开头，API 仅通过 system RPC 调用领域用例。

| 方法 / 路径 | 权限 | 行为 |
| --- | --- | --- |
| GET /dictionaries | system.dictionary.read | 分页查询分类、字典项数 |
| GET /dictionaries/:id | system.dictionary.read | 分类详情及字典项 |
| POST /dictionaries | system.dictionary.create | 创建分类与字典项，201 |
| PUT /dictionaries/:id | system.dictionary.update | 更新名称、状态及完整字典项集合 |
| DELETE /dictionaries/:id?revision=... | system.dictionary.delete | 删除已停用分类及子项 |
| GET /settings | system.setting.read | 分页查询配置 |
| POST /settings | system.setting.create | 创建配置，201 |
| PUT /settings/:id | system.setting.update | 校验原有类型后更新值与状态 |
| DELETE /settings/:id?revision=... | system.setting.delete | 删除已停用配置 |

写权限依赖相应的读取权限。RPC 方法在 admin-api 服务身份白名单中；用例校验原始 token 的有效会话与最新权限，不能通过用户 ID、角色请求头伪造身份。写事务与授权变更共享策略行锁，避免撤权和写入之间的竞态，但不会无故增加策略版本。

每次成功写入同时保存操作人、对象、修改前后字段及 Trace ID；审计写入失败时，分类、子项、配置及删除操作一起回滚。查看审计需要独立的 `system.audit.read` 权限，包含配置历史值。

新增字典请求示例：

```json
{
  "code": "order.status",
  "name": "订单状态",
  "description": "订单流程选项",
  "enabled": true,
  "items": [
    { "value": "pending", "label": "待处理", "sort": 10, "enabled": true }
  ]
}
```

完整响应与生成类型见 [OpenAPI](openapi/admin.json)。列表沿用 page、pageSize、keyword、enabled；详情返回 `data.record` 与 `data.items`。非法参数 400，未认证 401，无权限 403，不存在 404，重复键/过期 revision/删除启用记录 409，数据库或审计失败 503。

## 启用与验证

```bash
make migrate                    # 00004_system_settings.sql，显式建表
make sync-resources             # 新增权限，保留普通角色授权
make test-rbac                  # 独立临时库，包含并发和审计回滚
make check test build
make rbac-smoke                 # 运行中的 HTTP → RPC → MySQL 全链路
```

容器运行使用 `make compose-migrate compose-sync-resources`，在更新服务镜像后执行。首次初始化 `make bootstrap` 已包含权限同步。普通角色需要显式授予新权限，超级管理员自动拥有有效目录中的新权限。

集成测试包括分类/配置 CRUD、重复值与类型校验、停用删除、字典项回收、独立权限与撤销、两个应用实例的 revision 冲突、六种写操作的审计原子性及 HTTP CSRF/身份伪造拒绝。测试数据库和 HTTP fixture 使用随机名称，清理仅作用于本次测试数据。
