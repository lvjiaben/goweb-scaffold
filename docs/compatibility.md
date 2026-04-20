# Compatibility

## 当前兼容矩阵

| item | current | compatible with | notes |
| --- | --- | --- | --- |
| `goweb-scaffold` | `v0.9.0-rc.1` | `goweb-core v0.9.0-rc.1` / `v0.9.x` | 当前业务脚手架版本 |
| `goweb-core` | `v0.9.0-rc.1` | `goweb-scaffold v0.9.0-rc.1` / `v0.9.x` | 运行时内核 |
| `template_version` | `v7` | legacy migrate: `v5` / `v6` / `v7` | 只影响 scaffold codegen |

## codegen source 兼容

当前支持：

- `v5 -> v6 -> v7`
- `v6 -> v7`
- `v7 -> v7`

推荐流程：

1. 如果 source 很旧，先 `migrate-source`
2. 然后 `check-breaking`
3. 确认结果后再 `generate` 或 `regenerate`

## release candidate 注意

当前版本仍然是 `rc`：

- 允许继续完善发布治理和升级说明
- 不应在文档中伪装成 final stable
- 第十一阶段真实业务试跑前，不承诺 `1.0`
