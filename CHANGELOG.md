# Changelog

## v0.9.0-rc.1

Release candidate，仍未经过第十一阶段的真实业务试跑，因此不是 final stable。

已完成：

- admin/user 双端基础脚手架
- 自定义 RBAC，按 `permission_code` 校验
- 系统模块：`admin_auth`、`admin_user`、`admin_role`、`admin_menu`、`system_config`
- `attachment` 上传与管理
- `app_user_auth`、`app_user`
- codegen 生命周期：
  - preview
  - diff
  - generate
  - regenerate
  - remove
  - export
  - import
  - templates
  - migrate-source
  - batch
  - check-breaking
- `codegen.lock.json`、portable export、plan file、snapshot、breaking change 检测

明确不包含：

- user 端代码生成
- 第十一阶段的真实业务试跑结论
- 最终稳定版承诺

文档入口：

- `docs/versioning.md`
- `docs/compatibility.md`
- `docs/upgrade/UPGRADE_v0.9.0-rc.1.md`
- `docs/releases/v0.9.0-rc.1.md`
