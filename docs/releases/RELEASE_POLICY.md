# Release Policy

## 当前策略

- 当前仓库版本：`v0.9.0-rc.1`
- 当前模板版本：`v7`
- 当前状态：release candidate

## 发布边界

`goweb-scaffold` 负责：

- 可直接运行的后台脚手架
- admin / user 双应用
- RBAC
- 基础系统模块
- codegen 生命周期

`goweb-scaffold` 不负责：

- `goweb-core` 运行时内核版本治理
- 最终业务试跑结论
- user 页面代码生成

## 版本变更规则

- repo version 使用 semver
- template version 独立推进
- breaking change 必须先通过 `check-breaking` 能被识别，再考虑发布
- 每次 release candidate 必须同步更新：
  - `VERSION`
  - `CHANGELOG.md`
  - `docs/versioning.md`
  - `docs/compatibility.md`
  - `docs/upgrade/*`
  - `docs/releases/*`
