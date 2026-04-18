# goweb-scaffold

`goweb-scaffold` 是可直接跑的业务脚手架，依赖 `goweb-core` 作为运行时内核。

## 目录

- `cmd/server`：服务入口
- `configs`：YAML 配置
- `migrations`：PostgreSQL 初始化与种子数据
- `internal/bootstrap`：运行时装配、认证、RBAC、迁移
- `internal/modules`：后台与用户端模块
- `internal/gen/modules_gen.go`：显式模块注册，不修改中央路由文件
- `storage/uploads`：本地上传目录
- `vben-admin`：前端 admin/user 双应用底板

## 已含模块

- `admin_auth`
- `admin_user`
- `admin_role`
- `admin_menu`
- `system_config`
- `attachment`
- `app_user_auth`
- `app_user`
- `codegen`（占位骨架，不含生成逻辑）

## 运行要求

- Go 1.26+
- PostgreSQL 14+
- Node 20+（前端）

## 默认管理员

- 账号：`admin`
- 密码：`Admin@123456`
