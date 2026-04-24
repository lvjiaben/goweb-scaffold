# AI Guide

本文档给后续 AI/Codex 会话快速建立项目边界，避免偏离当前架构。

## 两仓关系

- `goweb-core` 是运行时内核，只提供 app/config/db/httpx/auth/rbac/files/validate/errorsx/logx 等基础能力。
- `goweb-scaffold` 是业务脚手架，包含后端业务模块、migrations、codegen、`vben-admin` 前端。

## 后端边界

- 后台前缀：`/backend`。
- 用户端前缀：`/api`。
- 响应结构：`code/message/data/request_id`。
- 成功：HTTP 200 且 `code=200`。
- 登录失效：HTTP 200 且 `code=401`。
- 普通错误仍按 HTTP 状态码返回。

## 前端边界

- 当前前端基线是旧仓库完整 `vben-admin`，后台 app 是 `apps/backend`，用户端 app 是 `apps/user`。
- 不允许把主线改回 `apps/admin`。
- CRUD 页面必须使用三文件结构：`list.vue`、`data.ts`、`modules/form-drawer.vue`。
- 页面样式、组件、抽屉、表格、菜单优先参考旧仓库和 vben 官方 playground。

## Codegen 边界

- 只生成 admin 后台 CRUD。
- 后端生成到 `internal/modules/app/<module>`。
- 前端生成到 `vben-admin/apps/backend/src/views/<module>`。
- 生成注册只更新 `internal/gen/modules_gen.go` 与 `vben-admin/apps/backend/src/router/routes/modules/generated.ts`。
- 不允许使用 magic comment 修改中央路由文件。

## 禁止事项

- 禁止引入 Gin、Chi、Casbin、RabbitMQ。
- 禁止新增 `/api/v1`。
- 禁止重写 `vben-admin` 主结构。
- 禁止把 GORM 查询写进 handler。
- 禁止把生成页面改成自定义 HTML 壳子。
