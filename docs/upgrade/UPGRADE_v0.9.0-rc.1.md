# Upgrade v0.9.0-rc.1

## 适用范围

适用于把旧的 lock/export/source 迁移到当前 `goweb-scaffold v0.9.0-rc.1` + `template v7`。

## 推荐顺序

1. 先确认当前仓库版本与 core 版本
2. 对旧 source 执行 `migrate-source`
3. 对目标模块执行 `check-breaking`
4. 如果结果是 `same` 或 `non_breaking`，再执行 `regenerate`
5. 如果结果是 `breaking`，先审查 route / permission / api / schema 变化，再决定是否落地

## 什么时候先 `migrate-source`

- lock/export 仍是 `v5` 或 `v6`
- 你准备把旧 source 导入当前仓库
- 你需要先看迁移后的 `snapshot` 结构

## 什么时候先 `check-breaking`

- 模板已经升级
- 模块已有稳定 lock
- 你担心 regenerate 会把旧模块弄坏

## 什么时候直接 `regenerate`

- 当前 lock 已经是 `v7`
- `check-breaking` 返回 `same`
- 或你明确接受 `non_breaking`

## release candidate 注意事项

- 当前版本仍是 `rc`
- 第十一阶段真实业务试跑前，不要把这版当成 `1.0`
- 如果历史 source 很旧，优先保留 export 备份，再迁移
