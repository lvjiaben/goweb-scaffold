<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import AppModal from '@/components/AppModal.vue';
import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import { adminState } from '@/auth';
import {
  deleteAdminUser,
  fetchAdminUserDetail,
  fetchAdminUsers,
  fetchRoleOptions,
  saveAdminUser,
  type AdminUserForm,
} from '@/api/admin-user';
import { formatTime, getErrorMessage } from '@/helpers';
import type { AdminUserItem, RoleOption, TableColumn } from '@/types';

const columns: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'username', title: '用户名', width: '180px' },
  { key: 'nickname', title: '昵称', width: '160px' },
  { key: 'status', title: '状态', width: '110px' },
  { key: 'is_super', title: '超级管理员', width: '130px' },
  { key: 'role_names', title: '角色' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '220px', align: 'right' },
];

const loading = ref(false);
const saving = ref(false);
const open = ref(false);
const keyword = ref('');
const errorMessage = ref('');
const rows = ref<AdminUserItem[]>([]);
const roleOptions = ref<RoleOption[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

const form = reactive<AdminUserForm>({
  username: '',
  password: '',
  nickname: '',
  status: 1,
  is_super: false,
  role_ids: [],
});

const modalTitle = computed(() => (form.id ? '编辑管理员' : '新建管理员'));
const totalPage = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));
const isEditingSelf = computed(() => Number(form.id || 0) === Number(adminState.user?.id || 0));

function resetForm() {
  form.id = undefined;
  form.username = '';
  form.password = '';
  form.nickname = '';
  form.status = 1;
  form.is_super = false;
  form.role_ids = [];
}

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchAdminUsers({
      keyword: keyword.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载管理员失败');
  } finally {
    loading.value = false;
  }
}

async function loadRoleOptions() {
  try {
    const result = await fetchRoleOptions();
    roleOptions.value = result.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载角色选项失败');
  }
}

function openCreate() {
  resetForm();
  open.value = true;
}

async function openEdit(id: number) {
  errorMessage.value = '';
  try {
    const detail = await fetchAdminUserDetail(id);
    form.id = detail.id;
    form.username = detail.username;
    form.password = '';
    form.nickname = detail.nickname;
    form.status = detail.status;
    form.is_super = detail.is_super;
    form.role_ids = [...detail.role_ids];
    open.value = true;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载管理员详情失败');
  }
}

async function submit() {
  saving.value = true;
  errorMessage.value = '';
  try {
    await saveAdminUser({
      id: form.id,
      username: form.username.trim(),
      password: form.password || undefined,
      nickname: form.nickname.trim(),
      status: Number(form.status),
      is_super: form.is_super,
      role_ids: [...form.role_ids],
    });
    open.value = false;
    resetForm();
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存管理员失败');
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: AdminUserItem) {
  if (!window.confirm(`确认删除管理员「${row.username}」吗？`)) {
    return;
  }
  try {
    await deleteAdminUser(row.id);
    if (rows.value.length === 1 && page.value > 1) {
      page.value -= 1;
    }
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除管理员失败');
  }
}

function toggleRole(roleID: number) {
  if (form.role_ids.includes(roleID)) {
    form.role_ids = form.role_ids.filter((item) => item !== roleID);
    return;
  }
  form.role_ids = [...form.role_ids, roleID];
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

onMounted(async () => {
  await Promise.all([load(), loadRoleOptions()]);
});
</script>

<template>
  <section class="page-stack">
    <article class="card page-card">
      <div class="section-heading">
        <div>
          <h3>管理员管理</h3>
          <p>支持关键词搜索、角色绑定、超级管理员标记和账号状态维护。</p>
        </div>
        <PermissionButton code="admin_user.save">
          <button class="btn" type="button" @click="openCreate">新建管理员</button>
        </PermissionButton>
      </div>

      <div class="toolbar-row">
        <div class="search-group">
          <input v-model="keyword" class="input" placeholder="搜索用户名或昵称" @keyup.enter="search" />
          <button class="btn secondary" type="button" @click="search">搜索</button>
        </div>
        <span class="text-muted">共 {{ total }} 条</span>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <AppTable :columns="columns" :rows="rows" :loading="loading" empty-text="暂无管理员数据">
        <template #cell-status="{ value }">
          <span class="status-pill" :class="Number(value) === 1 ? 'is-active' : 'is-disabled'">
            {{ Number(value) === 1 ? '启用' : '禁用' }}
          </span>
        </template>
        <template #cell-is_super="{ value }">
          <span class="status-pill" :class="value ? 'is-active' : 'is-muted'">
            {{ value ? '是' : '否' }}
          </span>
        </template>
        <template #cell-role_names="{ value }">
          <div class="tag-list">
            <span v-for="name in value || []" :key="name" class="tag-chip">{{ name }}</span>
            <span v-if="!(value || []).length" class="text-muted">未分配</span>
          </div>
        </template>
        <template #cell-created_at="{ value }">
          {{ formatTime(value) }}
        </template>
        <template #cell-actions="{ row }">
          <div class="table-actions">
            <PermissionButton code="admin_user.save">
              <button class="btn secondary btn-sm" type="button" @click="openEdit(row.id)">编辑</button>
            </PermissionButton>
            <PermissionButton code="admin_user.delete">
              <button
                class="btn danger btn-sm"
                type="button"
                :disabled="row.id === 1"
                @click="removeRow(row)"
              >
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

    <AppModal :open="open" :title="modalTitle" width="760px" @close="closeModal">
      <div class="form-grid two-columns">
        <FormField label="用户名" required>
          <input v-model="form.username" class="input" />
        </FormField>
        <FormField
          :label="form.id ? '密码' : '密码（新建必填）'"
          :required="!form.id"
          hint="编辑时留空表示不修改密码。"
        >
          <input v-model="form.password" class="input" type="password" placeholder="请输入密码" />
        </FormField>
        <FormField label="昵称" required>
          <input v-model="form.nickname" class="input" />
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
          <input v-model="form.is_super" type="checkbox" />
          <span>超级管理员</span>
        </label>
      </div>

      <FormField label="角色" hint="可多选；超级管理员也可以保留角色绑定。">
        <div class="check-grid">
          <label v-for="role in roleOptions" :key="role.value" class="check-card">
            <input
              type="checkbox"
              :checked="form.role_ids.includes(role.value)"
              @change="toggleRole(role.value)"
            />
            <div>
              <strong>{{ role.label }}</strong>
              <small>{{ role.code }}</small>
            </div>
          </label>
        </div>
      </FormField>

      <p v-if="isEditingSelf" class="hint-banner">当前正在编辑自己的账号，允许修改昵称和密码。</p>

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
