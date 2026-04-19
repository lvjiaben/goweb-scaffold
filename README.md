# goweb-scaffold

`goweb-scaffold` 是可直接运行的业务脚手架，运行时依赖 `goweb-core`。

## 目录

- `cmd/server`：HTTP 服务入口
- `configs`：YAML 配置与示例配置
- `migrations`：PostgreSQL 初始化和 seed SQL
- `internal/bootstrap`：运行时装配、认证、RBAC、迁移执行
- `internal/gen/modules_gen.go`：模块注册入口
- `internal/modules`：后台和用户端业务模块
- `storage/uploads`：本地上传目录
- `vben-admin`：`apps/admin` 和 `apps/user` 双应用

## 已含模块

- `admin_auth`
- `admin_user`
- `admin_role`
- `admin_menu`
- `system_config`
- `attachment`
- `app_user_auth`
- `app_user`
- `codegen`

## 运行要求

- Go `1.26+`
- PostgreSQL `14+`
- Node `20+`

## PostgreSQL 初始化

1. 创建数据库：

```sql
CREATE DATABASE goweb_scaffold;
```

2. 复制配置文件并修改 DSN：

```bash
cp configs/config.example.yaml configs/config.yaml
```

3. 启动服务时会自动执行 `migrations/*.sql`。

当前迁移会初始化以下基础表：

- `admin_user`
- `admin_role`
- `admin_menu`
- `admin_user_role`
- `admin_role_menu`
- `admin_login_log`
- `admin_session`
- `app_user`
- `app_user_session`
- `system_config`
- `file_attachment`
- `codegen_history`
- `schema_migrations`

## 默认管理员

- 用户名：`admin`
- 密码：`Admin@123456`

## 启动

后端：

```bash
go run ./cmd/server -config configs/config.yaml
```

admin 前端：

```bash
cd vben-admin/apps/admin
npm install
npm run dev
```

user 前端：

```bash
cd vben-admin/apps/user
npm install
npm run dev
```

## 当前 admin 页面

- `DashboardHomePage`：工作台概览
- `AdminUserPage`：管理员列表、保存、删除、角色绑定
- `AdminRolePage`：角色列表、保存、删除、菜单按钮授权
- `AdminMenuPage`：菜单树、按钮节点、父级菜单维护
- `SystemConfigPage`：系统配置列表与弹窗表单
- `AttachmentPage`：上传、预览、复制 URL、删除
- `CodegenPage`：元数据方案稿、预览、历史记录
- `ForbiddenView`：无权限或不可访问路由兜底页

## 鉴权与路由行为

- 后端接口仅使用 `GET` 和 `POST`
- 后台接口前缀为 `/admin-api`
- 用户端接口前缀为 `/api`
- 权限校验按 `permission_code` 执行
- 模块注册通过 `internal/gen/modules_gen.go`
- 前端 `401` 会统一清理本地 session，并自动跳转 `/login`
- 前端 `403` 会统一提示“当前账号无权执行该操作”
- 用户手输无权访问的后台路由时，不会白屏，会进入 `ForbiddenView`
- 顶栏标题和面包屑优先使用后端返回菜单标题

## 页面专用 options / tree 接口

- `GET /admin-api/admin_role/options`
  用于管理员表单中的角色选项，只对可编辑管理员的账号开放
- `GET /admin-api/admin_menu/tree`
  用于角色授权或菜单编辑场景，返回完整菜单树，包含按钮节点
- `GET /admin-api/admin_menu/options`
  用于菜单父级选择器，只返回菜单节点，不返回按钮节点

## Codegen 当前能力

当前 `codegen` 还不会真正生成文件，但已经进入“方案生成稿”阶段：

- 读取 PostgreSQL 业务表列表
- 读取字段元数据
- 推断字段组件与展示方式
- 输出 `page`、`api`、`inferred_fields`
- 输出 `form_schema`、`list_schema`、`search_schema`
- 保存历史记录并可重新载入当前表单

当前阶段明确不做：

- 不生成真实 Go / Vue 文件
- 不改 user app
- 不生成部署、测试、CI、k8s 等目录

## 说明

- `config.example.yaml` 可以直接复制为 `config.yaml` 作为本地模板
- 仓库不会提交本机私有 PostgreSQL 凭据
- codegen 未来只允许生成 admin 后台 CRUD，不生成 user 页面
