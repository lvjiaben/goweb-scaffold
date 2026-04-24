# Goweb Scaffold Codex Instructions

本仓库必须按现有框架约束开发，不允许重新发明前后端结构。

## 总规则

- 后端只用 Go、net/http、PostgreSQL、GORM。
- 后台接口前缀固定为 `/backend`，用户端接口前缀固定为 `/api`。
- 响应结构固定为 `code/message/data/request_id`，成功 `code=200`。
- 认证失效响应使用 HTTP 200 且 body `code=401`，普通参数错误仍使用 HTTP 400。
- 后端模块固定分域：`internal/modules/admin`、`internal/modules/app`、`internal/modules/system`。
- handler 只做 HTTP 参数解析和响应，业务逻辑在 service，数据库操作在 repo。
- 手写模块注册只维护 `internal/modules/manual_gen.go`。
- codegen 生成模块只放 `internal/modules/app/<module>`，注册只重建 `internal/gen/modules_gen.go`。
- 前端基线是 `vben-admin/apps/backend` 和 `vben-admin/apps/user`，不要改回 `apps/admin`。
- 后台 CRUD 页面固定为 `list.vue`、`data.ts`、`modules/form-drawer.vue`。
- CRUD 页面结构和交互优先参考旧仓库 `/var/golang/admin/vben-admin/apps/backend` 与 vben 官方 playground。
- codegen 只生成 admin 后台 CRUD，不生成 user 页面。
- 不准引入 Gin、Chi、Casbin、RabbitMQ。
- 不准新增 `/api/v1`。

## 开发前必读

- `docs/AI_GUIDE.md`
- `docs/BACKEND_STANDARD.md`
- `docs/FRONTEND_STANDARD.md`
- `docs/CODEGEN_STANDARD.md`
