<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import AppModal from '@/components/AppModal.vue';
import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import MenuTreeSelect from '@/components/MenuTreeSelect.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import {
  deleteAdminMenu,
  fetchAdminMenuDetail,
  fetchAdminMenuList,
  fetchAdminMenuOptions,
  saveAdminMenu,
  type AdminMenuForm,
} from '@/api/admin-menu';
import { flattenMenuTree, formatTime, getErrorMessage } from '@/helpers';
import { notifySuccess } from '@/notify';
import type { FlatMenuItem, MenuItem, MenuOption, TableColumn } from '@/types';

const columns: TableColumn[] = [
  { key: 'title', title: '标题' },
  { key: 'name', title: '名称', width: '160px' },
  { key: 'menu_type', title: '类型', width: '100px' },
  { key: 'path', title: '路径', width: '220px' },
  { key: 'permission_code', title: '权限码', width: '200px' },
  { key: 'sort', title: '排序', width: '80px', align: 'center' },
  { key: 'visible', title: '显示', width: '90px', align: 'center' },
  { key: 'status', title: '状态', width: '90px', align: 'center' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '200px', align: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const open = ref(false);
const keyword = ref('');
const errorMessage = ref('');
const menuTree = ref<MenuItem[]>([]);
const menuOptions = ref<MenuOption[]>([]);

const form = reactive<AdminMenuForm>({
  parent_id: 0,
  name: '',
  title: '',
  path: '',
  component: '',
  menu_type: 'menu',
  permission_code: '',
  icon: '',
  sort: 0,
  visible: true,
  status: 1,
});

const modalTitle = computed(() => (form.id ? '编辑菜单' : '新建菜单'));
const flatRows = computed(() => flattenMenuTree(menuTree.value));
const filteredRows = computed(() => {
  const current = keyword.value.trim().toLowerCase();
  if (!current) {
    return flatRows.value;
  }
  return flatRows.value.filter((item) =>
    [item.title, item.name, item.path, item.permission_code].some((field) =>
      String(field || '')
        .toLowerCase()
        .includes(current),
    ),
  );
});
const parentOptions = computed(() => menuOptions.value);
const parentLabel = computed(() => findMenuLabel(parentOptions.value, form.parent_id));

function findMenuLabel(options: MenuOption[], value: number): string {
  if (!value) {
    return '顶级菜单';
  }
  for (const item of options) {
    if (item.value === value) {
      return item.label;
    }
    if (item.children?.length) {
      const found = findMenuLabel(item.children, value);
      if (found) {
        return found;
      }
    }
  }
  return '';
}

function resetForm() {
  form.id = undefined;
  form.parent_id = 0;
  form.name = '';
  form.title = '';
  form.path = '';
  form.component = '';
  form.menu_type = 'menu';
  form.permission_code = '';
  form.icon = '';
  form.sort = 0;
  form.visible = true;
  form.status = 1;
}

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const treeResult = await fetchAdminMenuList();
    menuTree.value = treeResult.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载菜单树失败');
  } finally {
    loading.value = false;
  }
}

async function loadParentOptions() {
  try {
    const optionResult = await fetchAdminMenuOptions();
    menuOptions.value = optionResult.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载父级菜单选项失败');
  }
}

function openCreate(parentID = 0) {
  resetForm();
  form.parent_id = parentID;
  void loadParentOptions();
  open.value = true;
}

async function openEdit(id: number) {
  try {
    await loadParentOptions();
    const detail = await fetchAdminMenuDetail(id);
    form.id = detail.id;
    form.parent_id = detail.parent_id;
    form.name = detail.name;
    form.title = detail.title;
    form.path = detail.path;
    form.component = detail.component;
    form.menu_type = detail.menu_type;
    form.permission_code = detail.permission_code;
    form.icon = detail.icon;
    form.sort = detail.sort;
    form.visible = detail.visible;
    form.status = detail.status;
    open.value = true;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载菜单详情失败');
  }
}

async function submit() {
  saving.value = true;
  errorMessage.value = '';
  try {
    await saveAdminMenu({
      id: form.id,
      parent_id: Number(form.parent_id),
      name: form.name.trim(),
      title: form.title.trim(),
      path: form.path.trim(),
      component: form.component.trim(),
      menu_type: form.menu_type,
      permission_code: form.permission_code.trim(),
      icon: form.icon.trim(),
      sort: Number(form.sort || 0),
      visible: Boolean(form.visible),
      status: Number(form.status),
    });
    notifySuccess(form.id ? '菜单已更新' : '菜单已创建');
    closeModal();
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存菜单失败');
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: FlatMenuItem) {
  if (!window.confirm(`确认删除节点「${row.title}」及其子节点吗？`)) {
    return;
  }
  try {
    await deleteAdminMenu(row.id);
    notifySuccess('菜单节点已删除');
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除菜单失败');
  }
}

function closeModal() {
  open.value = false;
  resetForm();
}

onMounted(load);
</script>

<template>
  <section class="page-stack">
    <article class="card page-card">
      <div class="section-heading">
        <div>
          <h3>菜单管理</h3>
          <p>维护菜单树和按钮节点，按钮权限统一使用 permission_code，并支持父级菜单选择。</p>
        </div>
        <PermissionButton code="admin_menu.save">
          <button class="btn" type="button" @click="openCreate(0)">新建顶级菜单</button>
        </PermissionButton>
      </div>

      <div class="toolbar-row">
        <div class="search-group">
          <input
            v-model="keyword"
            class="input"
            placeholder="搜索标题、名称、路径或权限码"
          />
        </div>
        <span class="text-muted">共 {{ filteredRows.length }} 个节点</span>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <AppTable :columns="columns" :rows="filteredRows" :loading="loading" empty-text="暂无菜单数据">
        <template #cell-title="{ row }">
          <div class="tree-label">
            <span class="tree-indent" :style="{ width: `${row.depth * 20}px` }" />
            <strong>{{ row.title }}</strong>
          </div>
        </template>
        <template #cell-menu_type="{ value }">
          <span class="status-pill" :class="value === 'button' ? 'is-warning' : 'is-info'">
            {{ value === 'button' ? '按钮' : '菜单' }}
          </span>
        </template>
        <template #cell-path="{ value }">
          <code>{{ value || '-' }}</code>
        </template>
        <template #cell-permission_code="{ value }">
          <code>{{ value || '-' }}</code>
        </template>
        <template #cell-visible="{ value }">
          <span class="status-pill" :class="value ? 'is-active' : 'is-muted'">{{ value ? '显示' : '隐藏' }}</span>
        </template>
        <template #cell-status="{ value }">
          <span class="status-pill" :class="Number(value) === 1 ? 'is-active' : 'is-disabled'">
            {{ Number(value) === 1 ? '启用' : '禁用' }}
          </span>
        </template>
        <template #cell-created_at="{ value }">
          {{ formatTime(value) }}
        </template>
        <template #cell-actions="{ row }">
          <div class="table-actions">
            <PermissionButton v-if="row.menu_type === 'menu'" code="admin_menu.save">
              <button class="btn secondary btn-sm" type="button" @click="openCreate(row.id)">新增子项</button>
            </PermissionButton>
            <PermissionButton code="admin_menu.save">
              <button class="btn secondary btn-sm" type="button" @click="openEdit(row.id)">编辑</button>
            </PermissionButton>
            <PermissionButton code="admin_menu.delete">
              <button class="btn danger btn-sm" type="button" @click="removeRow(row)">删除</button>
            </PermissionButton>
          </div>
        </template>
      </AppTable>
    </article>

    <AppModal
      :open="open"
      :title="modalTitle"
      width="860px"
      :mask-closable="!saving"
      :esc-closable="!saving"
      @close="closeModal"
    >
      <div class="form-grid two-columns">
        <FormField label="节点类型" required>
          <select v-model="form.menu_type" class="input">
            <option value="menu">菜单</option>
            <option value="button">按钮</option>
          </select>
        </FormField>
        <FormField label="父级菜单">
          <div class="stack-sm">
            <button class="btn secondary btn-sm align-start" type="button" @click="form.parent_id = 0">
              设为顶级菜单
            </button>
            <span class="text-muted">当前父级：{{ parentLabel || '未选择' }}</span>
          </div>
        </FormField>
        <FormField label="标题" required>
          <input v-model="form.title" class="input" />
        </FormField>
        <FormField label="名称" required hint="建议使用 kebab-case 命名。">
          <input v-model="form.name" class="input" />
        </FormField>
        <FormField label="路径" :required="form.menu_type === 'menu'" hint="菜单节点以 / 开头，按钮可留空。">
          <input v-model="form.path" class="input" placeholder="/system/example" />
        </FormField>
        <FormField label="组件">
          <input v-model="form.component" class="input" placeholder="system/example/index" />
        </FormField>
        <FormField label="权限码" :required="form.menu_type === 'button'" hint="按钮类型必须填写。">
          <input v-model="form.permission_code" class="input" placeholder="admin_menu.save" />
        </FormField>
        <FormField label="图标">
          <input v-model="form.icon" class="input" placeholder="setting" />
        </FormField>
        <FormField label="排序">
          <input v-model="form.sort" class="input" type="number" min="0" />
        </FormField>
        <FormField label="状态" required>
          <select v-model="form.status" class="input">
            <option :value="1">启用</option>
            <option :value="2">禁用</option>
          </select>
        </FormField>
      </div>

      <div class="checkbox-line">
        <label class="checkbox-item">
          <input v-model="form.visible" type="checkbox" />
          <span>前台可见</span>
        </label>
      </div>

      <FormField label="父级菜单树" hint="父级只展示菜单节点，避免将菜单挂到按钮下面。">
        <div class="tree-panel">
          <MenuTreeSelect
            v-model="form.parent_id"
            :options="parentOptions"
            :disabled-values="form.id ? [form.id] : []"
          />
        </div>
      </FormField>

      <template #footer>
        <div class="modal-actions">
          <button class="btn secondary" type="button" @click="closeModal">取消</button>
          <button class="btn" type="button" :disabled="saving" @click="submit">
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </template>
    </AppModal>
  </section>
</template>
