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

## 说明

- 后端接口仅使用 `GET` 和 `POST`
- 后台接口前缀为 `/admin-api`
- 用户端接口前缀为 `/api`
- 权限校验按 `permission_code` 执行
- 模块注册通过 `internal/gen/modules_gen.go`
- `codegen` 当前阶段只提供表结构读取、预览和历史记录，不生成文件
