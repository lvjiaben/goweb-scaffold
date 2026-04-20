# Versioning

## 当前版本

- repo version: `v0.9.0-rc.1`
- template version: `v7`
- release status: release candidate

## 三层版本

### 1. repo version

`goweb-scaffold` 仓库版本遵循 semver，目前是 `v0.9.0-rc.1`。

影响范围：

- 目录结构
- CLI 能力
- 后端模块边界
- admin/user 应用整体基线

### 2. template version

codegen 模板版本当前是 `v7`。

影响范围：

- lock / export 内的 `template_version`
- preview/schema/render 输出
- snapshot / check-breaking 兼容性判断

当前支持迁移的 source version：

- `v5`
- `v6`
- `v7`

### 3. codegen source version

lock / export / plan 里的 `format`、`version`、`template_version` 共同决定 source 的兼容面。

- `format=codegen-export` / `version=v1`：portable export
- `template_version=v7`：当前模板视图
- 老版本 source 会先做内存迁移，再进入 `preview` / `diff` / `generate` / `regenerate` / `check-breaking`

## semver 与内部版本的边界

遵循 semver：

- `goweb-scaffold` 仓库版本
- `goweb-core` 依赖版本

不直接遵循 semver、而在内部 source 中体现：

- `template_version`
- `codegen-export format/version`
- `codegen.lock.json` 的 snapshot 结构

## 当前发布判断

`v0.9.0-rc.1` 的含义：

- 已具备正式发布治理能力
- 已具备 codegen 生命周期闭环
- 仍需第十一阶段真实业务试跑验证
- 因此不能宣称 `v1.0.0`
