<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import AppModal from '@/components/AppModal.vue';
import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import MenuTreeSelect from '@/components/MenuTreeSelect.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import {
  deleteAdminRole,
  fetchAdminRoleDetail,
  fetchAdminRoles,
  fetchMenuTreeOptions,
  saveAdminRole,
  type AdminRoleForm,
} from '@/api/admin-role';
import { formatTime, getErrorMessage } from '@/helpers';
import { notifySuccess } from '@/notify';
import type { MenuOption, RoleItem, TableColumn } from '@/types';

const columns: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'name', title: '名称', width: '180px' },
  { key: 'code', title: '编码', width: '220px' },
  { key: 'status', title: '状态', width: '120px' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '200px', align: 'right' },
];

const rows = ref<RoleItem[]>([]);
const menuOptions = ref<MenuOption[]>([]);
const loading = ref(false);
const saving = ref(false);
const open = ref(false);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const keyword = ref('');
const errorMessage = ref('');

const form = reactive<AdminRoleForm>({
  name: '',
  code: '',
  status: 1,
  menu_ids: [],
});

const modalTitle = computed(() => (form.id ? '编辑角色' : '新建角色'));
const totalPage = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));

function resetForm() {
  form.id = undefined;
  form.name = '';
  form.code = '';
  form.status = 1;
  form.menu_ids = [];
}

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchAdminRoles({
      keyword: keyword.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载角色失败');
  } finally {
    loading.value = false;
  }
}

async function loadMenuOptions() {
  try {
    const result = await fetchMenuTreeOptions();
    menuOptions.value = result.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载菜单树失败');
  }
}

function openCreate() {
  resetForm();
  void loadMenuOptions();
  open.value = true;
}

async function openEdit(id: number) {
  try {
    await loadMenuOptions();
    const detail = await fetchAdminRoleDetail(id);
    form.id = detail.id;
    form.name = detail.name;
    form.code = detail.code;
    form.status = detail.status;
    form.menu_ids = [...detail.menu_ids];
    open.value = true;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载角色详情失败');
  }
}

async function submit() {
  saving.value = true;
  errorMessage.value = '';
  try {
    await saveAdminRole({
      id: form.id,
      name: form.name.trim(),
      code: form.code.trim(),
      status: Number(form.status),
      menu_ids: [...form.menu_ids],
    });
    notifySuccess(form.id ? '角色已更新' : '角色已创建');
    closeModal();
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存角色失败');
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: RoleItem) {
  if (!window.confirm(`确认删除角色「${row.name}」吗？`)) {
    return;
  }
  try {
    await deleteAdminRole(row.id);
    if (rows.value.length === 1 && page.value > 1) {
      page.value -= 1;
    }
    notifySuccess('角色已删除');
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除角色失败');
  }
}

function closeModal() {
  open.value = false;
  resetForm();
}

async function search() {
  page.value = 1;
  await load();
}

async function goPage(next: number) {
  if (next < 1 || next > totalPage.value || next === page.value) {
    return;
  }
  page.value = next;
  await load();
}

onMounted(load);
</script>

<template>
  <section class="page-stack">
    <article class="card page-card">
      <div class="section-heading">
        <div>
          <h3>角色管理</h3>
          <p>角色保存时同步 role-menu 关联，授权节点按菜单树和按钮 permission_code 维护。</p>
        </div>
        <PermissionButton code="admin_role.save">
          <button class="btn" type="button" @click="openCreate">新建角色</button>
        </PermissionButton>
      </div>

      <div class="toolbar-row">
        <div class="search-group">
          <input v-model="keyword" class="input" placeholder="搜索角色名称或编码" @keyup.enter="search" />
          <button class="btn secondary" type="button" @click="search">搜索</button>
        </div>
        <span class="text-muted">共 {{ total }} 条</span>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <AppTable :columns="columns" :rows="rows" :loading="loading" empty-text="暂无角色数据">
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
            <PermissionButton code="admin_role.save">
              <button class="btn secondary btn-sm" type="button" @click="openEdit(row.id)">编辑</button>
            </PermissionButton>
            <PermissionButton code="admin_role.delete">
              <button class="btn danger btn-sm" type="button" :disabled="row.id === 1" @click="removeRow(row)">
                删除
              </button>
            </PermissionButton>
          </div>
        </template>
      </AppTable>

      <div class="table-footer">
        <span>第 {{ page }} / {{ totalPage }} 页</span>
        <div class="table-actions">
          <button class="btn secondary btn-sm" type="button" :disabled="page <= 1" @click="goPage(page - 1)">
            上一页
          </button>
          <button
            class="btn secondary btn-sm"
            type="button"
            :disabled="page >= totalPage"
            @click="goPage(page + 1)"
          >
            下一页
          </button>
        </div>
      </div>
    </article>

    <AppModal
      :open="open"
      :title="modalTitle"
      width="840px"
      :mask-closable="!saving"
      :esc-closable="!saving"
      @close="closeModal"
    >
      <div class="form-grid two-columns">
        <FormField label="名称" required>
          <input v-model="form.name" class="input" />
        </FormField>
        <FormField label="编码" required hint="角色编码需唯一，例如 content_editor。">
          <input v-model="form.code" class="input" />
        </FormField>
        <FormField label="状态" required>
          <select v-model="form.status" class="input">
            <option :value="1">启用</option>
            <option :value="2">禁用</option>
          </select>
        </FormField>
      </div>

      <FormField label="菜单权限" required hint="包含菜单和按钮，按钮节点会携带 permission_code。">
        <div class="tree-panel">
          <MenuTreeSelect v-model="form.menu_ids" :options="menuOptions" multiple />
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
