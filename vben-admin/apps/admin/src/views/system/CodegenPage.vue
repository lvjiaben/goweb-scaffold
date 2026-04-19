<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';

import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import {
  deleteCodegenHistory,
  fetchCodegenHistory,
  fetchCodegenPreview,
  fetchCodegenTableColumns,
  fetchCodegenTables,
  saveCodegenHistory,
  type CodegenPayload,
} from '@/api/codegen';
import { formatTime, getErrorMessage, prettyJSON } from '@/helpers';
import type {
  CodegenColumn,
  CodegenHistoryItem,
  CodegenPreview,
  CodegenTableInfo,
  TableColumn,
} from '@/types';

const columnTable: TableColumn[] = [
  { key: 'column_name', title: '字段名', width: '180px' },
  { key: 'data_type', title: '类型', width: '140px' },
  { key: 'is_nullable', title: '可空', width: '90px', align: 'center' },
  { key: 'is_primary_key', title: '主键', width: '90px', align: 'center' },
  { key: 'column_default', title: '默认值' },
];

const historyTable: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'module_name', title: '模块名', width: '160px' },
  { key: 'table_name', title: '数据表', width: '180px' },
  { key: 'status', title: '状态', width: '120px' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '140px', align: 'right' },
];

const tables = ref<CodegenTableInfo[]>([]);
const columns = ref<CodegenColumn[]>([]);
const historyRows = ref<CodegenHistoryItem[]>([]);
const preview = ref<CodegenPreview | null>(null);
const loadingTables = ref(false);
const loadingColumns = ref(false);
const loadingHistory = ref(false);
const previewing = ref(false);
const saving = ref(false);
const errorMessage = ref('');

const form = reactive<CodegenPayload>({
  module_name: '',
  table_name: '',
  payload: {
    list_fields: [],
    form_fields: [],
    search_fields: [],
  },
});

const selectedTable = ref('');
const canGenerate = computed(() => Boolean(form.module_name.trim() && form.table_name.trim()));

function resetFieldSelections(nextColumns: CodegenColumn[]) {
  const fields = nextColumns
    .map((item) => item.column_name)
    .filter((name) => !['deleted_at'].includes(name));
  const formFields = fields.filter((name) => !['id', 'created_at', 'updated_at'].includes(name));
  const searchFields = nextColumns
    .filter(
      (item) =>
        ['character varying', 'text', 'varchar'].includes(item.data_type) &&
        !['created_at', 'updated_at', 'deleted_at'].includes(item.column_name),
    )
    .slice(0, 3)
    .map((item) => item.column_name);

  form.payload.list_fields = fields;
  form.payload.form_fields = formFields;
  form.payload.search_fields = searchFields;
}

function payloadSnapshot(): CodegenPayload {
  return {
    module_name: form.module_name.trim(),
    table_name: form.table_name.trim(),
    payload: {
      list_fields: [...form.payload.list_fields],
      form_fields: [...form.payload.form_fields],
      search_fields: [...form.payload.search_fields],
    },
  };
}

function toggleField(bucket: 'form_fields' | 'list_fields' | 'search_fields', columnName: string) {
  const current = form.payload[bucket];
  if (current.includes(columnName)) {
    form.payload[bucket] = current.filter((item) => item !== columnName);
    return;
  }
  form.payload[bucket] = [...current, columnName];
}

async function loadTables() {
  loadingTables.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchCodegenTables();
    tables.value = result.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载业务表列表失败');
  } finally {
    loadingTables.value = false;
  }
}

async function loadHistory() {
  loadingHistory.value = true;
  try {
    const result = await fetchCodegenHistory();
    historyRows.value = result.list || [];
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载生成历史失败');
  } finally {
    loadingHistory.value = false;
  }
}

async function loadColumns(tableName: string) {
  if (!tableName) {
    columns.value = [];
    preview.value = null;
    return;
  }
  loadingColumns.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchCodegenTableColumns(tableName);
    columns.value = result.list || [];
    if (!form.module_name) {
      form.module_name = tableName;
    }
    resetFieldSelections(columns.value);
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '读取字段元数据失败');
  } finally {
    loadingColumns.value = false;
  }
}

async function previewCurrent() {
  previewing.value = true;
  errorMessage.value = '';
  try {
    preview.value = await fetchCodegenPreview(payloadSnapshot());
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '生成预览失败');
  } finally {
    previewing.value = false;
  }
}

async function saveCurrent() {
  saving.value = true;
  errorMessage.value = '';
  try {
    await saveCodegenHistory(payloadSnapshot());
    await loadHistory();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '保存生成配置失败');
  } finally {
    saving.value = false;
  }
}

async function removeHistory(row: CodegenHistoryItem) {
  if (!window.confirm(`确认删除生成历史 #${row.id} 吗？`)) {
    return;
  }
  try {
    await deleteCodegenHistory(row.id);
    await loadHistory();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除生成历史失败');
  }
}

function applyTable(item: CodegenTableInfo) {
  const previousTable = form.table_name;
  selectedTable.value = item.table_name;
  form.table_name = item.table_name;
  if (!form.module_name || form.module_name === previousTable) {
    form.module_name = item.table_name;
  }
}

watch(
  () => selectedTable.value,
  async (value) => {
    form.table_name = value;
    await loadColumns(value);
  },
);

onMounted(async () => {
  await Promise.all([loadTables(), loadHistory()]);
});
</script>

<template>
  <section class="page-stack">
    <article class="card page-card">
      <div class="section-heading">
        <div>
          <h3>代码生成</h3>
          <p>当前阶段只做元数据准备：读取 PostgreSQL 表结构、生成预览、保存配置历史，不生成文件。</p>
        </div>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <div class="codegen-layout">
        <section class="card inset-card codegen-sidebar">
          <div class="section-heading compact">
            <div>
              <h4>数据表列表</h4>
              <p>排除了基础表、迁移表和内部表。</p>
            </div>
          </div>
          <div v-if="loadingTables" class="empty-state">加载中...</div>
          <div v-else-if="tables.length" class="table-list">
            <button
              v-for="item in tables"
              :key="item.table_name"
              class="table-list__item"
              :class="{ active: selectedTable === item.table_name }"
              type="button"
              @click="applyTable(item)"
            >
              <strong>{{ item.display_name }}</strong>
              <span>{{ item.table_name }}</span>
            </button>
          </div>
          <div v-else class="empty-state">当前没有可用于 codegen 的业务表。</div>
        </section>

        <section class="page-stack">
          <article class="card page-card">
            <div class="section-heading compact">
              <div>
                <h4>生成配置</h4>
                <p>选择表后维护模块名和字段配置，再获取预览。</p>
              </div>
            </div>

            <div class="form-grid two-columns">
              <FormField label="模块名" required hint="将来用于模块目录、路由和 API 模块名。">
                <input v-model="form.module_name" class="input" placeholder="article" />
              </FormField>
              <FormField label="数据表" required>
                <select v-model="selectedTable" class="input">
                  <option value="">请选择业务表</option>
                  <option v-for="item in tables" :key="item.table_name" :value="item.table_name">
                    {{ item.table_name }}
                  </option>
                </select>
              </FormField>
            </div>

            <div class="field-groups">
              <FormField label="列表字段">
                <div class="check-grid">
                  <label v-for="column in columns" :key="`list-${column.column_name}`" class="check-card">
                    <input
                      type="checkbox"
                      :checked="form.payload.list_fields.includes(column.column_name)"
                      @change="toggleField('list_fields', column.column_name)"
                    />
                    <div>
                      <strong>{{ column.column_name }}</strong>
                      <small>{{ column.data_type }}</small>
                    </div>
                  </label>
                </div>
              </FormField>

              <FormField label="表单字段">
                <div class="check-grid">
                  <label v-for="column in columns" :key="`form-${column.column_name}`" class="check-card">
                    <input
                      type="checkbox"
                      :checked="form.payload.form_fields.includes(column.column_name)"
                      @change="toggleField('form_fields', column.column_name)"
                    />
                    <div>
                      <strong>{{ column.column_name }}</strong>
                      <small>{{ column.data_type }}</small>
                    </div>
                  </label>
                </div>
              </FormField>

              <FormField label="搜索字段">
                <div class="check-grid">
                  <label v-for="column in columns" :key="`search-${column.column_name}`" class="check-card">
                    <input
                      type="checkbox"
                      :checked="form.payload.search_fields.includes(column.column_name)"
                      @change="toggleField('search_fields', column.column_name)"
                    />
                    <div>
                      <strong>{{ column.column_name }}</strong>
                      <small>{{ column.data_type }}</small>
                    </div>
                  </label>
                </div>
              </FormField>
            </div>

            <div class="table-actions align-end">
              <PermissionButton code="codegen.save">
                <button
                  class="btn secondary"
                  type="button"
                  :disabled="previewing || !canGenerate"
                  @click="previewCurrent"
                >
                  {{ previewing ? '预览中...' : '生成预览' }}
                </button>
              </PermissionButton>
              <PermissionButton code="codegen.save">
                <button class="btn" type="button" :disabled="saving || !canGenerate" @click="saveCurrent">
                  {{ saving ? '保存中...' : '保存到历史' }}
                </button>
              </PermissionButton>
            </div>
          </article>

          <article class="card page-card">
            <div class="section-heading compact">
              <div>
                <h4>字段元数据</h4>
                <p>来自 <code>information_schema.columns</code> 的字段结构。</p>
              </div>
            </div>
            <AppTable :columns="columnTable" :rows="columns" :loading="loadingColumns" empty-text="请选择业务表">
              <template #cell-is_nullable="{ value }">
                <span class="status-pill" :class="value ? 'is-muted' : 'is-active'">{{ value ? '是' : '否' }}</span>
              </template>
              <template #cell-is_primary_key="{ value }">
                <span class="status-pill" :class="value ? 'is-warning' : 'is-muted'">{{ value ? '是' : '否' }}</span>
              </template>
              <template #cell-column_default="{ value }">
                <code>{{ value || '-' }}</code>
              </template>
            </AppTable>
          </article>

          <article class="card page-card">
            <div class="section-heading compact">
              <div>
                <h4>预览区</h4>
                <p>当前只返回未来生成所需的元数据，不输出实际文件。</p>
              </div>
            </div>

            <div v-if="preview" class="preview-grid">
              <div class="card inset-card preview-card">
                <strong>页面</strong>
                <span>Route: {{ preview.page.route_path }}</span>
                <span>Page: {{ preview.page.page_name }}</span>
                <span>View: {{ preview.page.view_file }}</span>
              </div>
              <div class="card inset-card preview-card">
                <strong>接口</strong>
                <span v-for="(value, key) in preview.api" :key="key">{{ key }}: {{ value }}</span>
              </div>
              <div class="card inset-card preview-card preview-card--full">
                <strong>Payload</strong>
                <pre class="mini-code">{{ prettyJSON(preview.payload) }}</pre>
              </div>
              <div class="card inset-card preview-card preview-card--full">
                <strong>提示</strong>
                <ul class="note-list">
                  <li v-for="note in preview.notes" :key="note">{{ note }}</li>
                </ul>
              </div>
            </div>
            <div v-else class="empty-state">先选择业务表并点击“生成预览”。</div>
          </article>

          <article class="card page-card">
            <div class="section-heading compact">
              <div>
                <h4>历史记录</h4>
                <p>保存 module_name、table_name 和基础字段配置，便于下一阶段接入真实生成逻辑。</p>
              </div>
            </div>
            <AppTable :columns="historyTable" :rows="historyRows" :loading="loadingHistory" empty-text="暂无历史记录">
              <template #cell-status="{ value }">
                <span class="status-pill is-info">{{ value }}</span>
              </template>
              <template #cell-created_at="{ value }">
                {{ formatTime(value) }}
              </template>
              <template #cell-actions="{ row }">
                <PermissionButton code="codegen.delete">
                  <button class="btn danger btn-sm" type="button" @click="removeHistory(row)">删除</button>
                </PermissionButton>
              </template>
            </AppTable>
          </article>
        </section>
      </div>
    </article>
  </section>
</template>
