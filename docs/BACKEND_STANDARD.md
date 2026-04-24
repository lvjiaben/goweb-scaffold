# Backend Standard

## 目录分域

后端模块只允许放在：

- `internal/modules/admin/*`：后台管理员、角色、菜单、认证等。
- `internal/modules/app/*`：用户端模块和 codegen 生成业务模块。
- `internal/modules/system/*`：系统配置、附件、codegen、公共接口、dashboard 等。

基础设施不放进 modules：

- `internal/platform/queue`
- `internal/platform/cron`
- `internal/platform/cache`
- `internal/platform/events`
- `internal/http/middleware`

## 模块分层

每个模块至少按以下文件组织：

- `module.go`：只注册模块和路由。
- `handler.go`：只解析 query/body、调用 service、返回响应。
- `service.go`：业务规则、事务编排、参数校验、调用 repo。
- `repo.go`：所有 GORM 查询、创建、更新、删除。
- `dto.go`：request/response/list/delete DTO。
- `model.go`：当前模块自己的 GORM model。

handler 禁止出现：

- `runtime.DB`
- `.Model(`
- `.Where(`
- `.Count(`
- `.Find(`
- `.First(`
- `.Create(`
- `.Save(`
- `.Delete(`
- `.Transaction(`
- `.Begin(`

## 注册职责

- `internal/modules/manual_gen.go` 只注册手写模块。
- `internal/gen/modules_gen.go` 只注册 codegen 生成模块。
- codegen 只能重建 `internal/gen/modules_gen.go`。

## 路由与响应

- 后台接口固定 `/backend`。
- 用户端接口固定 `/api`。
- 响应字段固定 `code/message/data/request_id`。
- 成功 `code=200`。
- 登录失效 HTTP 200 + `code=401`，前端据此跳转登录页。

## Query Builder

通用查询能力放在 `internal/shared/query`，模块 repo 应优先复用它处理：

- 分页
- search 模糊搜索
- filter JSON
- 精确匹配
- 模糊匹配
- 时间范围
- sort_by/sort_order
- `sort/weight/weigh DESC` 默认排序

## Migration

- PostgreSQL only。
- 主键 bigint。
- 时间字段 timestamptz。
- 扩展字段 jsonb。
- 新字段必须通过 migrations 落库，不允许只改 model。
