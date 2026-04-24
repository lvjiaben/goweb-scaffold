# Goweb Framework 约束

## 两仓关系

- `goweb-core` 是运行时内核，只放通用能力：`app`、`auth`、`config`、`db`、`errorsx`、`files`、`httpx`、`logx`、`rbac`、`validate`。
- `goweb-scaffold` 是业务脚手架，依赖 `goweb-core`，包含后端模块、codegen、migrations、配置和 `vben-admin` 前端。

## scaffold 目录

- `cmd/server`：后端启动入口。
- `cmd/codegen`：无头 codegen CLI。
- `configs`：本地配置模板。
- `migrations/0001_init.sql`：唯一初始化 SQL。
- `internal/bootstrap`：运行时、配置、鉴权、中间件辅助。
- `internal/modules/admin`：后台管理员、角色、菜单、后台认证。
- `internal/modules/app`：业务模块与用户模块，codegen 生成模块也放这里。
- `internal/modules/system`：系统配置、附件、codegen、公共接口、dashboard。
- `internal/gen`：codegen service、模板、writer、registry。
- `internal/shared/model`：共享基础模型。
- `internal/shared/query`：统一查询 builder。
- `internal/platform`：未来 queue、cron、cache、events 等基础设施。
- `internal/http/middleware`：未来 HTTP 中间件。
- `vben-admin/apps/backend`：后台前端。
- `vben-admin/apps/user`：用户端前端。

## 后端规范

- 业务接口只用 `GET` 和 `POST`。
- 后台前缀固定 `/backend`，用户端前缀固定 `/api`，禁止新增 `/api/v1`。
- 响应结构固定 `code/message/data/request_id`。
- 成功 `code=200`；认证失效 HTTP 200 + `code=401`；普通参数错误仍可 HTTP 400。
- 模块必须按 `handler/service/repo/dto/model` 分层，handler 不写 GORM。
- 新业务模块如果同时有后台和用户端接口，必须在同一个模块内拆 `backend_handler.go`、`api_handler.go`、`backend_dto.go`、`api_dto.go`。
- `service.go` 承载业务规则、事务编排、校验；`repo.go` 承载 GORM 查询；`model.go` 只放本模块模型。
- 手写模块注册在 `internal/modules/manual_gen.go`。
- codegen 生成模块注册在 `internal/gen/modules_gen.go`，生成器只重建这个文件。
- RBAC 按 `permission_code` 校验，禁止按 URL path 模糊匹配。
- 禁止引入 Gin、Chi、Casbin、RabbitMQ。

## 用户模块规则

- `internal/modules/app/user` 统一承载用户相关 backend 和 api。
- `backend_handler.go` 提供后台用户 CRUD、余额调整、积分调整、日志查看。
- `api_handler.go` 提供用户登录、注册、退出、profile、密码修改。
- `app_user` 包含 `pid`、`tid`、`money`、`score`、`code`、`avatar`、`status_text`、`wechat_unionid`、`wechat_openid`、`version`。
- 余额变更必须写 `user_money_log`；积分变更必须写 `user_score_log`。
- money/score 更新必须使用事务：锁定用户、计算 before/after、更新用户表、插入日志。

## Codegen 规范

- codegen 只生成 backend 管理端 CRUD，不生成 user 前端页面。
- 生成后端路径固定 `internal/modules/app/<module>/`。
- 生成后端文件固定：`module.go`、`backend_handler.go`、`api_handler.go`、`service.go`、`repo.go`、`backend_dto.go`、`api_dto.go`、`model.go`、`meta.go`、`codegen.lock.json`。
- `api_handler.go` 和 `api_dto.go` 默认只生成空骨架。
- 生成前端路径固定：`vben-admin/apps/backend/src/views/<module>/list.vue`、`data.ts`、`modules/form-drawer.vue`。
- sort/weight/weigh 默认 `DESC` 排序。
- 表格必须设置稳定 `rowConfig.keyField = 'id'`，避免 checkbox 一选全选。
- 组件配置 JSON 示例不能放进 i18n message。
- lock/regenerate/remove/check-breaking/batch 能力保留，但示例 demo 不提交。

## vben-admin 规范

- 前端基线固定 `vben-admin/apps/backend` 和 `vben-admin/apps/user`，禁止改回 `apps/admin`。
- CRUD 页面只能采用三文件结构：`list.vue`、`data.ts`、`modules/form-drawer.vue`。
- `toolbar-actions` 放新增按钮和批量按钮；`toolbar-tools` 放搜索。
- 批量按钮组使用 `VbenButtonGroup border`，只保留 `X` 图标删除选中。
- 操作列使用 `VbenButtonGroup border` 和 `variant="icon"`。
- 新增 drawer 用 `formDrawerApi.setData({}).open()`；编辑用 `formDrawerApi.setData(row).open()`；新增子级只传必要字段如 `{ pid: row.id }`。
- 默认值放在 `form-drawer.vue` 内，不在 list 页塞完整默认对象。
- 搜索表单支持 URL query 回填。
- 菜单管理必须是树表，组件下拉来自页面组件列表，权限树必须有 checkbox。
- `index.html` 标题变量使用 `%VITE_APP_TITLE%`。

## 禁止事项

- 不新增 tests、docs、examples、demo、release、workflows。
- 不保留多份 md 文档；新增规范只能追加本文件。
- 不把 GORM 写进 handler。
- 不重写 `vben-admin` 主结构。
- 不生成 user 前端页面。
- 不把 backend/api handler 混在一个文件。
- 不把全局业务模型塞进某个模块的 `model.go`。

## 新 AI 接手规则

新对话或 AI 接手时，第一步必须读取 `FRAMEWORK.md`。后续所有代码修改必须遵守本文件。新增规范只能追加到 `FRAMEWORK.md`，不能再新建 `docs`、`README`、`AI_GUIDE`、`BACKEND_STANDARD`、`FRONTEND_STANDARD`、`CODEGEN_STANDARD` 等文档文件。
