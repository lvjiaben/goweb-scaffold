# Frontend Standard

## 基线

前端基线固定为当前 `vben-admin`：

- 后台：`vben-admin/apps/backend`
- 用户端：`vben-admin/apps/user`
- workspace、packages、internal 按 vben-admin monorepo 保持。

不允许把后台主线改成 `apps/admin`。

## CRUD 三文件结构

所有后台 CRUD 页面必须是：

```text
vben-admin/apps/backend/src/views/<module>/
├── list.vue
├── data.ts
└── modules/
    └── form-drawer.vue
```

职责：

- `list.vue`：列表、搜索、批量操作、删除、打开 drawer。
- `data.ts`：VXE columns、search schema、options、`searchFormFields`。
- `modules/form-drawer.vue`：新建和编辑表单。

## CRUD 交互规范

- drawer 打开使用 `formDrawerApi.setData(row).open()`。
- toolbar 新增/批量操作放 `#toolbar-actions`。
- 模糊搜索放 `#toolbar-tools`。
- 操作列使用 `<VbenButtonGroup border>` 和 `variant="icon"`。
- 列表必须设置稳定 `rowConfig.keyField = 'id'`。
- VXE checkbox 不允许一选全选，生成页默认 `checkboxConfig.reserve=true`。
- 搜索表单应支持 URL query 自动填充。

## 菜单页面规范

- 菜单类型支持 `menu`、`button`、`iframe`、`link`。
- 菜单字段支持 `enname`。
- 页面组件字段必须使用 `componentKeys` 自动下拉，不退化成普通文本输入。
- 角色权限树必须有复选框并提交 `menu_ids`。

## 参考来源

- CRUD 页面参考旧仓库 `/var/golang/admin/vben-admin/apps/backend/src/views/user`。
- 菜单表单参考 `/var/golang/admin/vben-admin/apps/backend/src/views/admin/menu/modules/form.vue`。
- 非 CRUD 通用写法参考 vben 官方 playground。
