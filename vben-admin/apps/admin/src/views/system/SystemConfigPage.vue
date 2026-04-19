<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import AppModal from '@/components/AppModal.vue';
import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import {
  fetchSystemConfigDetail,
  fetchSystemConfigs,
  saveSystemConfig,
  type SystemConfigForm,
} from '@/api/system-config';
import { formatTime, getErrorMessage, prettyJSON } from '@/helpers';
import type { SystemConfigItem, TableColumn } from '@/types';

const columns: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'config_key', title: '配置键', width: '200px' },
  { key: 'config_name', title: '配置名称', width: '180px' },
  { key: 'config_value', title: '配置值' },
  { key: 'remark', title: '备注', width: '180px' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '140px', align: 'right' },
];

const rows = ref<SystemConfigItem[]>([]);
const loading = ref(false);
const saving = ref(false);
const open = ref(false);
const keyword = ref('');
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const errorMessage = ref('');
const configText = ref('{\n  \n}');

const form = reactive<SystemConfigForm>({
  config_key: '',
  config_name: '',
  config_value: {},
  remark: '',
});

const modalTitle = computed(() => (form.id ? '编辑配置' : '新建配置'));
const totalPage = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)));

function resetForm() {
  form.id = undefined;
  form.config_key = '';
  form.config_name = '';
  form.config_value = {};
  form.remark = '';
  configText.value = '{\n  \n}';
}

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchSystemConfigs({
      keyword: keyword.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载系统配置失败');
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  resetForm();
  open.value = true;
}

async function openEdit(id: number) {
  try {
    const detail = await fetchSystemConfigDetail(id);
    form.id = detail.id;
    form.config_key = detail.config_key;
    form.config_name = detail.config_name;
    form.config_value = detail.config_value;
    form.remark = detail.remark;
    configText.value = prettyJSON(detail.config_value);
    open.value = true;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载配置详情失败');
  }
}

async function submit() {
  saving.value = true;
  errorMessage.value = '';
  try {
    form.config_value = JSON.parse(configText.value || '{}');
    await saveSystemConfig({
      id: form.id,
      config_key: form.config_key.trim(),
      config_name: form.config_name.trim(),
      config_value: form.config_value,
      remark: form.remark.trim(),
    });
    closeModal();
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存系统配置失败');
  } finally {
    saving.value = false;
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
          <h3>系统配置</h3>
          <p>使用表格和弹窗维护配置项，配置值暂时保留 JSON 文本编辑，但外层是正式表单。</p>
        </div>
        <PermissionButton code="system_config.save">
          <button class="btn" type="button" @click="openCreate">新建配置</button>
        </PermissionButton>
      </div>

      <div class="toolbar-row">
        <div class="search-group">
          <input v-model="keyword" class="input" placeholder="搜索配置键或配置名称" @keyup.enter="search" />
          <button class="btn secondary" type="button" @click="search">搜索</button>
        </div>
        <span class="text-muted">共 {{ total }} 条</span>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <AppTable :columns="columns" :rows="rows" :loading="loading" empty-text="暂无配置数据">
        <template #cell-config_value="{ value }">
          <pre class="mini-code">{{ prettyJSON(value) }}</pre>
        </template>
        <template #cell-created_at="{ value }">
          {{ formatTime(value) }}
        </template>
        <template #cell-actions="{ row }">
          <div class="table-actions">
            <PermissionButton code="system_config.save">
              <button class="btn secondary btn-sm" type="button" @click="openEdit(row.id)">编辑</button>
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

    <AppModal :open="open" :title="modalTitle" width="840px" @close="closeModal">
      <div class="form-grid two-columns">
        <FormField label="配置键" required hint="config_key 必须唯一。">
          <input v-model="form.config_key" class="input" />
        </FormField>
        <FormField label="配置名称" required>
          <input v-model="form.config_name" class="input" />
        </FormField>
      </div>

      <FormField label="配置值 JSON">
        <textarea v-model="configText" class="input textarea" rows="10" />
      </FormField>

      <FormField label="备注">
        <textarea v-model="form.remark" class="input textarea" rows="4" />
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
