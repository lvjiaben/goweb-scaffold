# Codegen Standard

## 生成范围

codegen 只生成 admin 后台 CRUD，不生成 user 页面。

## 后端生成目录

后端生成模块固定在：

```text
internal/modules/app/<module>/
├── module.go
├── handler.go
├── service.go
├── repo.go
├── dto.go
├── model.go
├── meta.go
└── codegen.lock.json
```

生成模块必须满足：

- handler 无 GORM。
- service 承担业务判断。
- repo 承担所有 DB 查询。
- model 只定义当前模块 Entity。

## 前端生成目录

前端生成模块固定在：

```text
vben-admin/apps/backend/src/views/<module>/
├── list.vue
├── data.ts
└── modules/
    └── form-drawer.vue
```

API 文件固定：

```text
vben-admin/apps/backend/src/api/<module>.ts
```

## 字段推断

字段推断必须在后端 codegen service 层完成，页面 preview、CLI preview、generate、batch、regenerate 共用同一套规则。

推断要求：

- `id`：列表显示、不可表单编辑、可排序。
- `created_at/updated_at`：列表显示、可排序、搜索可用时间范围。
- `deleted_at`：默认隐藏。
- `sort/weight/weigh`：数字输入、列表显示、可排序，默认 `DESC`。
- `status/state/enabled/is_*`：select/radio/switch，表格 tag，自动 options。
- `*_id`：关联 ID，优先 TableSelect/Select。
- `*_ids`：多关联，优先 multiple。
- `varchar`：Input，默认列表显示，可模糊搜索。
- `text`：Textarea，默认表单显示，长文本默认不进列表。
- `json/jsonb`：JSON/textarea，默认不搜索、不列表显示。

## 排序规则

所有带 `sort`、`weight`、`weigh` 的列表默认按值越大越靠前：

```sql
ORDER BY sort DESC, id DESC
```

## 复选框规则

生成列表必须设置：

- `rowConfig.keyField = 'id'`
- checkbox 使用稳定唯一 row key
- 批量删除只读 `getCheckboxRecords()` 的选中行

## 生命周期能力

必须保持以下能力可用：

- lock
- export/import
- migrate-source
- diff
- generate
- regenerate
- remove
- batch
- check-breaking
- snapshot
