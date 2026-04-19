<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

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
import { notifySuccess } from '@/notify';
import type {
  CodegenColumn,
  CodegenHistoryItem,
  CodegenPreview,
  CodegenSchemaItem,
  CodegenTableInfo,
  TableColumn,
} from '@/types';

const columnTable: TableColumn[] = [
  { key: 'column_name', title: '字段名', width: '180px' },
  { key: 'data_type', title: '类型', width: '150px' },
  { key: 'is_nullable', title: '可空', width: '90px', align: 'center' },
  { key: 'is_primary_key', title: '主键', width: '90px', align: 'center' },
  { key: 'column_default', title: '默认值' },
];

const inferredFieldTable: TableColumn[] = [
  { key: 'column_name', title: '字段名', width: '160px' },
  { key: 'guessed_form_component', title: '表单组件', width: '150px' },
  { key: 'guessed_list_display', title: '列表展示', width: '130px' },
  { key: 'guessed_searchable', title: '可搜索', width: '90px', align: 'center' },
  { key: 'guessed_sortable', title: '可排序', width: '90px', align: 'center' },
];

const schemaTable: TableColumn[] = [
  { key: 'field', title: '字段', width: '160px' },
  { key: 'label', title: '标签', width: '180px' },
  { key: 'component', title: '组件', width: '160px' },
  { key: 'display', title: '展示' },
  { key: 'operator', title: '搜索操作符', width: '120px' },
  { key: 'flags', title: '标记', width: '220px' },
];

const historyTable: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'module_name', title: '模块名', width: '160px' },
  { key: 'table_name', title: '数据表', width: '180px' },
  { key: 'status', title: '状态', width: '120px' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '220px', align: 'right' },
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
const tableHint = computed(() =>
  preview.value
    ? `API 模块：${preview.value.api.module_code}，页面：${preview.value.page.page_name}`
    : '选择业务表后生成预览，查看未来 CRUD 方案稿。',
);

function parsePayload(value: unknown): CodegenPayload['payload'] {
  const payload =
    value && typeof value === 'object'
      ? (value as Partial<CodegenPayload['payload']>)
      : {};
  return {
    list_fields: Array.isArray(payload.list_fields) ? [...payload.list_fields] : [],
    form_fields: Array.isArray(payload.form_fields) ? [...payload.form_fields] : [],
    search_fields: Array.isArray(payload.search_fields) ? [...payload.search_fields] : [],
  };
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

function resetFieldSelections(nextColumns: CodegenColumn[]) {
  const fields = nextColumns
    .map((item) => item.column_name)
    .filter((name) => !['deleted_at'].includes(name));
  const formFields = fields.filter((name) => !['id', 'created_at', 'updated_at'].includes(name));
  const searchFields = nextColumns
    .filter(
      (item) =>
        ['character varying', 'text', 'varchar', 'timestamp with time zone'].includes(item.data_type) &&
        !['created_at', 'updated_at', 'deleted_at'].includes(item.column_name),
    )
    .slice(0, 3)
    .map((item) => item.column_name);

  form.payload.list_fields = fields;
  form.payload.form_fields = formFields;
  form.payload.search_fields = searchFields;
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
    form.table_name = '';
    return;
  }

  loadingColumns.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchCodegenTableColumns(tableName);
    columns.value = result.list || [];
    form.table_name = tableName;
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

async function changeTable(tableName: string) {
  const previousTable = form.table_name;
  selectedTable.value = tableName;
  if (!form.module_name || form.module_name === previousTable) {
    form.module_name = tableName;
  }
  await loadColumns(tableName);
}

async function previewCurrent() {
  previewing.value = true;
  errorMessage.value = '';
  try {
    preview.value = await fetchCodegenPreview(payloadSnapshot());
    notifySuccess('方案预览已更新');
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
    notifySuccess('生成配置已保存到历史');
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
    notifySuccess('生成历史已删除');
    await loadHistory();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除生成历史失败');
  }
}

async function applyHistory(row: CodegenHistoryItem) {
  const payload = parsePayload(row.payload);
  form.module_name = row.module_name;
  await changeTable(row.table_name);
  form.payload = {
    list_fields: [...payload.list_fields],
    form_fields: [...payload.form_fields],
    search_fields: [...payload.search_fields],
  };
  preview.value = null;
  notifySuccess(`已载入历史配置 #${row.id}`);
}

function schemaFlags(row: CodegenSchemaItem) {
  const flags = [];
  if (row.required) flags.push('required');
  if (row.readonly) flags.push('readonly');
  if (row.hidden) flags.push('hidden');
  if (row.searchable) flags.push('searchable');
  if (row.sortable) flags.push('sortable');
  return flags;
}

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
          <p>当前阶段产出完整方案稿：字段推断、表单 schema、列表 schema、搜索 schema 和页面/API 设计信息。</p>
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
              @click="changeTable(item.table_name)"
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
                <p>{{ tableHint }}</p>
              </div>
            </div>

            <div class="form-grid two-columns">
              <FormField label="模块名" required hint="将来用于模块目录、路由和 API 模块名。">
                <input v-model="form.module_name" class="input" placeholder="article" />
              </FormField>
              <FormField label="数据表" required>
                <select v-model="selectedTable" class="input" @change="changeTable(selectedTable)">
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
                  {{ previewing ? '预览中...' : '生成方案稿' }}
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
                <p>来自 PostgreSQL `information_schema.columns`。</p>
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
                <h4>方案预览</h4>
                <p>不落真实文件，但已经给出下一阶段 codegen 可直接消费的结构化方案。</p>
              </div>
            </div>

            <div v-if="preview" class="page-stack">
              <div class="preview-grid">
                <div class="card inset-card preview-card">
                  <strong>页面元信息</strong>
                  <span>Route: {{ preview.page.route_path }}</span>
                  <span>Page: {{ preview.page.page_name }}</span>
                  <span>View: {{ preview.page.view_file }}</span>
                </div>
                <div class="card inset-card preview-card">
                  <strong>API 元信息</strong>
                  <span>Module: {{ preview.api.module_code }}</span>
                  <span>List: {{ preview.api.list }}</span>
                  <span>Save: {{ preview.api.save }}</span>
                </div>
              </div>

              <article class="card inset-card preview-card preview-card--full">
                <strong>字段推断结果</strong>
                <AppTable
                  :columns="inferredFieldTable"
                  :rows="preview.inferred_fields"
                  empty-text="暂无推断结果"
                >
                  <template #cell-guessed_searchable="{ value }">
                    <span class="status-pill" :class="value ? 'is-active' : 'is-muted'">{{ value ? '是' : '否' }}</span>
                  </template>
                  <template #cell-guessed_sortable="{ value }">
                    <span class="status-pill" :class="value ? 'is-active' : 'is-muted'">{{ value ? '是' : '否' }}</span>
                  </template>
                </AppTable>
              </article>

              <article class="card inset-card preview-card preview-card--full">
                <strong>列表方案</strong>
                <AppTable :columns="schemaTable" :rows="preview.list_schema" empty-text="暂无列表 schema">
                  <template #cell-flags="{ row }">
                    <div class="tag-list">
                      <span v-for="flag in schemaFlags(row)" :key="flag" class="tag-chip">{{ flag }}</span>
                    </div>
                  </template>
                </AppTable>
              </article>

              <article class="card inset-card preview-card preview-card--full">
                <strong>表单方案</strong>
                <AppTable :columns="schemaTable" :rows="preview.form_schema" empty-text="暂无表单 schema">
                  <template #cell-flags="{ row }">
                    <div class="tag-list">
                      <span v-for="flag in schemaFlags(row)" :key="flag" class="tag-chip">{{ flag }}</span>
                    </div>
                  </template>
                </AppTable>
              </article>

              <article class="card inset-card preview-card preview-card--full">
                <strong>搜索方案</strong>
                <AppTable :columns="schemaTable" :rows="preview.search_schema" empty-text="暂无搜索 schema">
                  <template #cell-flags="{ row }">
                    <div class="tag-list">
                      <span v-for="flag in schemaFlags(row)" :key="flag" class="tag-chip">{{ flag }}</span>
                    </div>
                  </template>
                </AppTable>
              </article>

              <article class="card inset-card preview-card preview-card--full">
                <strong>配置快照</strong>
                <pre class="mini-code">{{ prettyJSON(preview.payload) }}</pre>
              </article>

              <article class="card inset-card preview-card preview-card--full">
                <strong>说明</strong>
                <ul class="note-list">
                  <li v-for="note in preview.notes" :key="note">{{ note }}</li>
                </ul>
              </article>
            </div>
            <div v-else class="empty-state">先选择业务表并点击“生成方案稿”。</div>
          </article>

          <article class="card page-card">
            <div class="section-heading compact">
              <div>
                <h4>历史记录</h4>
                <p>可以把历史配置重新载入到当前表单，继续修改并重做预览。</p>
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
                <div class="table-actions">
                  <PermissionButton code="codegen.save">
                    <button class="btn secondary btn-sm" type="button" @click="applyHistory(row)">载入配置</button>
                  </PermissionButton>
                  <PermissionButton code="codegen.delete">
                    <button class="btn danger btn-sm" type="button" @click="removeHistory(row)">删除</button>
                  </PermissionButton>
                </div>
              </template>
            </AppTable>
          </article>
        </section>
      </div>
    </article>
  </section>
</template>
