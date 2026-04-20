<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import {
  deleteCodegenHistory,
  fetchCodegenExport,
  fetchCodegenDiff,
  fetchCodegenHistory,
  fetchCodegenModules,
  fetchCodegenPreview,
  fetchCodegenTableColumns,
  fetchCodegenTables,
  generateCodegenFiles,
  regenerateCodegenFiles,
  removeCodegenModule,
  saveCodegenHistory,
  type CodegenPayload,
  type CodegenPayloadBody,
} from '@/api/codegen';
import AppModal from '@/components/AppModal.vue';
import AppTable from '@/components/AppTable.vue';
import FormField from '@/components/FormField.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import { formatTime, getErrorMessage, prettyJSON } from '@/helpers';
import { notifyError, notifyInfo, notifySuccess } from '@/notify';
import type {
  CodegenColumn,
  CodegenDiffResult,
  CodegenExportFile,
  CodegenFieldOption,
  CodegenFieldOverride,
  CodegenGenerateResult,
  CodegenHistoryItem,
  CodegenManagedModule,
  CodegenMenuRecord,
  CodegenPreview,
  CodegenRemoveResult,
  CodegenSchemaItem,
  CodegenSourceKind,
  CodegenTableInfo,
  TableColumn,
} from '@/types';

type OverrideBoolValue = '' | 'false' | 'true';

type FieldConfigRow = {
  column_name: string;
  data_type: string;
  column_comment: string;
  in_list: boolean;
  in_form: boolean;
  in_search: boolean;
  auto_label: string;
  auto_component: string;
  auto_placeholder: string;
  auto_required: boolean;
  auto_readonly: boolean;
  auto_hidden: boolean;
  auto_searchable: boolean;
  auto_sortable: boolean;
  auto_width: string;
  auto_options: CodegenFieldOption[];
  auto_default_value: unknown;
  label: string;
  component: string;
  placeholder: string;
  required: OverrideBoolValue;
  readonly: OverrideBoolValue;
  hidden: OverrideBoolValue;
  searchable: OverrideBoolValue;
  sortable: OverrideBoolValue;
  width: string;
  options_text: string;
  default_value_text: string;
};

const registryFiles = [
  'internal/gen/modules_gen.go',
  'vben-admin/apps/admin/src/generated/routes.ts',
];

const componentOptions = [
  { label: '自动推断', value: '' },
  { label: '文本输入', value: 'text-input' },
  { label: '多行文本', value: 'textarea' },
  { label: '数字输入', value: 'number-input' },
  { label: '下拉选择', value: 'select' },
  { label: '单选组', value: 'radio' },
  { label: '开关', value: 'switch' },
  { label: '时间选择', value: 'datetime-picker' },
  { label: 'JSON 文本', value: 'json-editor' },
  { label: '只读文本', value: 'readonly-text' },
  { label: '只读时间', value: 'readonly-datetime' },
  { label: '隐藏字段', value: 'hidden' },
];

const columnTable: TableColumn[] = [
  { key: 'column_name', title: '字段名', width: '160px' },
  { key: 'data_type', title: '类型', width: '140px' },
  { key: 'column_comment', title: '列注释', width: '180px' },
  { key: 'is_nullable', title: '可空', width: '80px', align: 'center' },
  { key: 'is_primary_key', title: '主键', width: '80px', align: 'center' },
  { key: 'column_default', title: '默认值' },
];

const historyTable: TableColumn[] = [
  { key: 'id', title: 'ID', width: '80px' },
  { key: 'module_name', title: '模块名', width: '160px' },
  { key: 'table_name', title: '数据表', width: '180px' },
  { key: 'status', title: '状态', width: '120px' },
  { key: 'remark', title: '备注', width: '220px' },
  { key: 'created_at', title: '创建时间', width: '180px' },
  { key: 'actions', title: '操作', width: '420px', align: 'right' },
];

const diffTable: TableColumn[] = [
  { key: 'path', title: '文件路径', width: '360px' },
  { key: 'status', title: '状态', width: '100px' },
  { key: 'hashes', title: '内容摘要', width: '220px' },
  { key: 'changed_sections', title: '变更摘要' },
];

const tables = ref<CodegenTableInfo[]>([]);
const modules = ref<CodegenManagedModule[]>([]);
const columns = ref<CodegenColumn[]>([]);
const historyRows = ref<CodegenHistoryItem[]>([]);
const fieldConfigs = ref<FieldConfigRow[]>([]);
const preview = ref<CodegenPreview | null>(null);
const diffResult = ref<CodegenDiffResult | null>(null);
const generateResult = ref<CodegenGenerateResult | null>(null);
const removeResult = ref<CodegenRemoveResult | null>(null);
const loadingTables = ref(false);
const loadingModules = ref(false);
const loadingColumns = ref(false);
const loadingHistory = ref(false);
const previewing = ref(false);
const diffing = ref(false);
const saving = ref(false);
const generating = ref(false);
const removing = ref(false);
const regeneratingHistoryId = ref<number | null>(null);
const regeneratingModuleName = ref('');
const selectedManagedModuleName = ref('');
const errorMessage = ref('');
const loadedSourceKind = ref<CodegenSourceKind>('direct');
const exportFileInput = ref<HTMLInputElement | null>(null);

const sourceKindLabels: Record<CodegenSourceKind, string> = {
  direct: '直接选表',
  payload: 'Payload 文件',
  export: 'Export 文件',
  lock: 'Lock 文件',
  history: '历史记录',
};

const form = reactive<{
  module_name: string;
  table_name: string;
  payload: {
    title: string;
    list_fields: string[];
    form_fields: string[];
    search_fields: string[];
  };
}>({
  module_name: '',
  table_name: '',
  payload: {
    title: '',
    list_fields: [],
    form_fields: [],
    search_fields: [],
  },
});

const generateOptions = reactive({
  overwrite: true,
  register_module: true,
  upsert_menu: true,
});

const removeModal = reactive({
  open: false,
  module_name: '',
  remove_files: true,
  unregister_module: true,
  remove_menu: true,
  remove_history: false,
  remove_lock: true,
});

const selectedTableInfo = computed(
  () => tables.value.find((item) => item.table_name === form.table_name) || null,
);

const selectedManagedModule = computed(
  () => modules.value.find((item) => item.module_name === selectedManagedModuleName.value) || null,
);

const removeTargetModule = computed(
  () => modules.value.find((item) => item.module_name === removeModal.module_name) || null,
);

const effectiveTitle = computed(() => {
  return preview.value?.payload.title || selectedTableInfo.value?.display_name || form.module_name || '未命名模块';
});

const loadedSourceKindLabel = computed(() => sourceKindLabels[loadedSourceKind.value] || loadedSourceKind.value);

const tableHint = computed(() => {
  if (!preview.value) {
    return '先做 preview，再看 diff，最后 generate、regenerate 或 remove。';
  }
  return `路由 ${preview.value.page.route_path}，API 模块 ${preview.value.api.module_code}。`;
});

const canOperate = computed(() => Boolean(form.module_name.trim() && form.table_name.trim()));

const selectedOverrideEntries = computed(() => {
  const overrides = selectedManagedModule.value?.payload.field_overrides || {};
  return Object.entries(overrides).map(([field, config]) => ({
    field,
    config,
  }));
});

const removeTargetFiles = computed(() => {
  const target = removeTargetModule.value;
  if (!target) {
    return {
      moduleFiles: [] as string[],
      registry: [] as string[],
    };
  }
  const moduleFiles = target.files.filter((item) => !registryFiles.includes(item));
  const registry = target.files.filter((item) => registryFiles.includes(item));
  return { moduleFiles, registry };
});

const removeTargetMenuRecords = computed(() => {
  const target = removeTargetModule.value;
  if (!target) {
    return [] as CodegenMenuRecord[];
  }
  return [
    {
      id: 0,
      name: target.module_name,
      title: target.payload.title || target.preview_summary.table_comment || target.module_name,
      path: target.route_path,
      menu_type: 'menu',
    },
    ...target.permission_codes.map((code) => ({
      id: 0,
      name: code,
      title: code,
      menu_type: 'button',
      permission_code: code,
    })),
  ];
});

function setActionError(error: unknown, fallback: string) {
  const message = getErrorMessage(error, fallback);
  errorMessage.value = message;
  notifyError(message);
}

function boolOverrideToInput(value: boolean | undefined): OverrideBoolValue {
  if (typeof value !== 'boolean') {
    return '';
  }
  return value ? 'true' : 'false';
}

function parseBoolOverride(value: OverrideBoolValue): boolean | undefined {
  if (value === '') {
    return undefined;
  }
  return value === 'true';
}

function stringifyEditableValue(value: unknown) {
  if (value === undefined) {
    return '';
  }
  if (typeof value === 'string') {
    return value;
  }
  return prettyJSON(value);
}

function parseEditableValue(value: string): unknown {
  const trimmed = value.trim();
  if (!trimmed) {
    return undefined;
  }
  try {
    return JSON.parse(trimmed);
  } catch {
    return trimmed;
  }
}

function normalizeFieldOptions(value: unknown): CodegenFieldOption[] {
  if (!Array.isArray(value)) {
    throw new Error('options 必须是 JSON 数组');
  }
  return value
    .map((item) => {
      if (!item || typeof item !== 'object') {
        return null;
      }
      const row = item as Record<string, unknown>;
      const label = String(row.label || '').trim();
      if (!label) {
        return null;
      }
      return {
        label,
        value: row.value,
      };
    })
    .filter((item): item is CodegenFieldOption => Boolean(item));
}

function normalizePayload(value: unknown): CodegenPayloadBody {
  const payload = value && typeof value === 'object' ? (value as Partial<CodegenPayloadBody>) : {};
  return {
    title: typeof payload.title === 'string' ? payload.title : '',
    list_fields: Array.isArray(payload.list_fields) ? [...payload.list_fields] : [],
    form_fields: Array.isArray(payload.form_fields) ? [...payload.form_fields] : [],
    search_fields: Array.isArray(payload.search_fields) ? [...payload.search_fields] : [],
    field_overrides:
      payload.field_overrides && typeof payload.field_overrides === 'object'
        ? { ...payload.field_overrides }
        : {},
  };
}

function schemaIndex(items: CodegenSchemaItem[]) {
  return new Map(items.map((item) => [item.field, item]));
}

function inferredIndex(nextPreview: CodegenPreview) {
  return new Map(nextPreview.inferred_fields.map((item) => [item.column_name, item]));
}

function buildFieldRows(nextPreview: CodegenPreview): FieldConfigRow[] {
  const listIndex = schemaIndex(nextPreview.list_schema);
  const formIndex = schemaIndex(nextPreview.form_schema);
  const searchIndex = schemaIndex(nextPreview.search_schema);
  const inferred = inferredIndex(nextPreview);
  const overrides = nextPreview.payload.field_overrides || {};

  return columns.value.map((column) => {
    const inferredField = inferred.get(column.column_name);
    const listField = listIndex.get(column.column_name);
    const formField = formIndex.get(column.column_name);
    const searchField = searchIndex.get(column.column_name);
    const override = overrides[column.column_name] || {};

    const autoLabel = inferredField?.guessed_label || formField?.label || listField?.label || column.column_name;
    const autoComponent =
      inferredField?.guessed_form_component || formField?.component || searchField?.component || '';
    const autoPlaceholder = formField?.placeholder || searchField?.placeholder || '';
    const autoOptions = formField?.options || searchField?.options || listField?.options || [];
    const autoDefaultValue = formField?.default_value;

    return {
      column_name: column.column_name,
      data_type: column.data_type,
      column_comment: column.column_comment || '',
      in_list: nextPreview.payload.list_fields.includes(column.column_name),
      in_form: nextPreview.payload.form_fields.includes(column.column_name),
      in_search: nextPreview.payload.search_fields.includes(column.column_name),
      auto_label: autoLabel,
      auto_component: autoComponent,
      auto_placeholder: autoPlaceholder,
      auto_required: Boolean(formField?.required),
      auto_readonly: Boolean(formField?.readonly),
      auto_hidden: Boolean(formField?.hidden),
      auto_searchable: Boolean(searchField?.searchable ?? listField?.searchable),
      auto_sortable: Boolean(listField?.sortable),
      auto_width: listField?.width || formField?.width || searchField?.width || '',
      auto_options: autoOptions,
      auto_default_value: autoDefaultValue,
      label: override.label || '',
      component: override.component || '',
      placeholder: override.placeholder || '',
      required: boolOverrideToInput(override.required),
      readonly: boolOverrideToInput(override.readonly),
      hidden: boolOverrideToInput(override.hidden),
      searchable: boolOverrideToInput(override.searchable),
      sortable: boolOverrideToInput(override.sortable),
      width: override.width || '',
      options_text: override.options?.length ? prettyJSON(override.options) : '',
      default_value_text: stringifyEditableValue(override.default_value),
    };
  });
}

function syncPreview(nextPreview: CodegenPreview) {
  preview.value = nextPreview;
  form.payload.list_fields = [...nextPreview.payload.list_fields];
  form.payload.form_fields = [...nextPreview.payload.form_fields];
  form.payload.search_fields = [...nextPreview.payload.search_fields];
  fieldConfigs.value = buildFieldRows(nextPreview);
}

function buildFieldOverrides() {
  const overrides: Record<string, CodegenFieldOverride> = {};
  for (const row of fieldConfigs.value) {
    const next: CodegenFieldOverride = {};
    const label = row.label.trim();
    const component = row.component.trim();
    const placeholder = row.placeholder.trim();
    const width = row.width.trim();

    if (label) {
      next.label = label;
    }
    if (component) {
      next.component = component;
    }
    if (placeholder) {
      next.placeholder = placeholder;
    }

    const required = parseBoolOverride(row.required);
    const readonly = parseBoolOverride(row.readonly);
    const hidden = parseBoolOverride(row.hidden);
    const searchable = parseBoolOverride(row.searchable);
    const sortable = parseBoolOverride(row.sortable);

    if (required !== undefined) next.required = required;
    if (readonly !== undefined) next.readonly = readonly;
    if (hidden !== undefined) next.hidden = hidden;
    if (searchable !== undefined) next.searchable = searchable;
    if (sortable !== undefined) next.sortable = sortable;
    if (width) next.width = width;

    const optionsText = row.options_text.trim();
    if (optionsText) {
      next.options = normalizeFieldOptions(parseEditableValue(optionsText));
    }

    const defaultValueText = row.default_value_text.trim();
    if (defaultValueText) {
      next.default_value = parseEditableValue(defaultValueText);
    }

    if (Object.keys(next).length > 0) {
      overrides[row.column_name] = next;
    }
  }
  return overrides;
}

function payloadSnapshot(): CodegenPayload {
  return {
    module_name: form.module_name.trim(),
    table_name: form.table_name.trim(),
    payload: {
      title: form.payload.title.trim() || undefined,
      list_fields: fieldConfigs.value.filter((item) => item.in_list).map((item) => item.column_name),
      form_fields: fieldConfigs.value.filter((item) => item.in_form).map((item) => item.column_name),
      search_fields: fieldConfigs.value.filter((item) => item.in_search).map((item) => item.column_name),
      field_overrides: buildFieldOverrides(),
    },
  };
}

function changeSelection(row: FieldConfigRow, bucket: 'list' | 'form' | 'search', checked: boolean) {
  if (bucket === 'list') {
    row.in_list = checked;
    return;
  }
  if (bucket === 'form') {
    row.in_form = checked;
    return;
  }
  row.in_search = checked;
}

function ensureManagedModuleSelection(items: CodegenManagedModule[]) {
  if (!items.length) {
    selectedManagedModuleName.value = '';
    return;
  }
  const current = items.find((item) => item.module_name === selectedManagedModuleName.value);
  if (!current) {
    selectedManagedModuleName.value = items[0].module_name;
  }
}

function downloadJSONFile(fileName: string, value: unknown) {
  const blob = new Blob([`${prettyJSON(value)}\n`], { type: 'application/json;charset=utf-8' });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = fileName;
  link.click();
  window.URL.revokeObjectURL(url);
}

async function loadTables() {
  loadingTables.value = true;
  try {
    const result = await fetchCodegenTables();
    tables.value = result.list || [];
  } catch (error) {
    setActionError(error, '加载业务表列表失败');
  } finally {
    loadingTables.value = false;
  }
}

async function loadModules() {
  loadingModules.value = true;
  try {
    const result = await fetchCodegenModules();
    modules.value = result.list || [];
    ensureManagedModuleSelection(modules.value);
  } catch (error) {
    setActionError(error, '加载已生成模块失败');
  } finally {
    loadingModules.value = false;
  }
}

async function loadHistory() {
  loadingHistory.value = true;
  try {
    const result = await fetchCodegenHistory();
    historyRows.value = result.list || [];
  } catch (error) {
    setActionError(error, '加载生成历史失败');
  } finally {
    loadingHistory.value = false;
  }
}

async function refreshLifecycleData() {
  await Promise.all([loadModules(), loadHistory()]);
}

async function loadColumns(tableName: string) {
  if (!tableName) {
    columns.value = [];
    fieldConfigs.value = [];
    preview.value = null;
    diffResult.value = null;
    generateResult.value = null;
    removeResult.value = null;
    form.table_name = '';
    loadedSourceKind.value = 'direct';
    return;
  }

  loadingColumns.value = true;
  try {
    const result = await fetchCodegenTableColumns(tableName);
    columns.value = result.list || [];
    form.table_name = tableName;
    preview.value = null;
    diffResult.value = null;
    generateResult.value = null;
    removeResult.value = null;
    fieldConfigs.value = [];
  } catch (error) {
    setActionError(error, '读取字段元数据失败');
  } finally {
    loadingColumns.value = false;
  }
}

async function changeTable(tableName: string, payload?: CodegenPayloadBody, sourceKind: CodegenSourceKind = 'direct') {
  const previousTable = form.table_name;
  if (!form.module_name || form.module_name === previousTable) {
    form.module_name = tableName;
  }
  await loadColumns(tableName);

  const normalized = payload ? normalizePayload(payload) : null;
  form.payload.list_fields = normalized ? [...normalized.list_fields] : [];
  form.payload.form_fields = normalized ? [...normalized.form_fields] : [];
  form.payload.search_fields = normalized ? [...normalized.search_fields] : [];
  form.payload.title = normalized?.title || '';
  loadedSourceKind.value = sourceKind;

  if (tableName) {
    await previewCurrent(false);
  }
}

async function previewCurrent(showNotice = true) {
  if (!canOperate.value) {
    return null;
  }
  previewing.value = true;
  errorMessage.value = '';
  try {
    const nextPreview = await fetchCodegenPreview(payloadSnapshot());
    syncPreview(nextPreview);
    diffResult.value = null;
    generateResult.value = null;
    removeResult.value = null;
    if (showNotice) {
      notifySuccess('方案预览已更新');
    }
    return nextPreview;
  } catch (error) {
    setActionError(error, '生成预览失败');
    return null;
  } finally {
    previewing.value = false;
  }
}

async function diffCurrent() {
  if (!canOperate.value) {
    return;
  }
  diffing.value = true;
  errorMessage.value = '';
  try {
    const nextPreview = await previewCurrent(false);
    if (!nextPreview) {
      return;
    }
    diffResult.value = await fetchCodegenDiff({
      module_name: nextPreview.module_name,
      table_name: nextPreview.table_name,
      payload: nextPreview.payload,
      overwrite: generateOptions.overwrite,
      register_module: generateOptions.register_module,
      upsert_menu: generateOptions.upsert_menu,
    });
    generateResult.value = null;
    removeResult.value = null;
    notifyInfo('diff 已更新');
  } catch (error) {
    setActionError(error, '生成 diff 失败');
  } finally {
    diffing.value = false;
  }
}

async function saveCurrent() {
  if (!canOperate.value) {
    return;
  }
  saving.value = true;
  errorMessage.value = '';
  try {
    const nextPreview = await previewCurrent(false);
    if (!nextPreview) {
      return;
    }
    await saveCodegenHistory({
      module_name: nextPreview.module_name,
      table_name: nextPreview.table_name,
      payload: nextPreview.payload,
    });
    notifySuccess('生成配置已保存到历史');
    await loadHistory();
  } catch (error) {
    setActionError(error, '保存生成配置失败');
  } finally {
    saving.value = false;
  }
}

async function generateCurrent() {
  if (!canOperate.value) {
    return;
  }
  generating.value = true;
  errorMessage.value = '';
  try {
    const nextPreview = await previewCurrent(false);
    if (!nextPreview) {
      return;
    }
    generateResult.value = await generateCodegenFiles({
      module_name: nextPreview.module_name,
      table_name: nextPreview.table_name,
      payload: nextPreview.payload,
      overwrite: generateOptions.overwrite,
      register_module: generateOptions.register_module,
      upsert_menu: generateOptions.upsert_menu,
    });
    removeResult.value = null;
    selectedManagedModuleName.value = nextPreview.module_name;
    notifySuccess('代码生成完成');
    await refreshLifecycleData();
  } catch (error) {
    setActionError(error, '生成文件失败');
  } finally {
    generating.value = false;
  }
}

async function removeHistoryRecord(row: CodegenHistoryItem) {
  if (!window.confirm(`确认删除生成历史 #${row.id} 吗？`)) {
    return;
  }
  try {
    await deleteCodegenHistory(row.id);
    notifySuccess('生成历史已删除');
    await loadHistory();
  } catch (error) {
    setActionError(error, '删除生成历史失败');
  }
}

async function applyHistory(row: CodegenHistoryItem) {
  const payload = normalizePayload(row.payload);
  form.module_name = row.module_name;
  selectedManagedModuleName.value = row.module_name;
  await changeTable(row.table_name, payload, 'history');
  notifySuccess(`已载入历史配置 #${row.id}`);
}

async function generateFromHistory(row: CodegenHistoryItem) {
  await applyHistory(row);
  await generateCurrent();
}

async function viewHistoryDiff(row: CodegenHistoryItem) {
  await applyHistory(row);
  await diffCurrent();
}

async function regenerateFromHistory(row: CodegenHistoryItem) {
  regeneratingHistoryId.value = row.id;
  errorMessage.value = '';
  try {
    generateResult.value = await regenerateCodegenFiles({
      history_id: row.id,
      overwrite: generateOptions.overwrite,
      register_module: generateOptions.register_module,
      upsert_menu: generateOptions.upsert_menu,
    });
    removeResult.value = null;
    selectedManagedModuleName.value = row.module_name;
    notifySuccess(`已根据历史 #${row.id} 重新生成`);
    await refreshLifecycleData();
  } catch (error) {
    setActionError(error, '重新生成失败');
  } finally {
    regeneratingHistoryId.value = null;
  }
}

async function applyManagedModule(module: CodegenManagedModule, showNotice = true) {
  selectedManagedModuleName.value = module.module_name;
  form.module_name = module.module_name;
  await changeTable(module.table_name, module.payload, 'lock');
  if (showNotice) {
    notifySuccess(`已载入模块 ${module.module_name} 的 lock 配置`);
  }
}

async function exportManagedModule(module: CodegenManagedModule) {
  try {
    const result = await fetchCodegenExport(module.module_name);
    downloadJSONFile(`${module.module_name}.codegen.json`, result);
    notifySuccess(`已导出模块 ${module.module_name} 的配置`);
  } catch (error) {
    setActionError(error, '导出配置失败');
  }
}

function openExportImportDialog() {
  exportFileInput.value?.click();
}

async function applyExportDocument(document: CodegenExportFile) {
  if (document.format !== 'codegen-export' || !document.module_name || !document.table_name) {
    throw new Error('导入文件不是合法的 codegen export');
  }
  selectedManagedModuleName.value = document.module_name;
  form.module_name = document.module_name;
  await changeTable(document.table_name, document.payload, 'export');
}

async function handleExportImport(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    return;
  }
  try {
    const raw = JSON.parse(await file.text()) as CodegenExportFile;
    await applyExportDocument(raw);
    notifySuccess(`已载入 export 文件 ${file.name}`);
  } catch (error) {
    setActionError(error, '载入 export 文件失败');
  } finally {
    input.value = '';
  }
}

function selectManagedModule(moduleName: string) {
  selectedManagedModuleName.value = moduleName;
}

async function viewManagedDiff(module: CodegenManagedModule) {
  await applyManagedModule(module, false);
  await diffCurrent();
}

async function regenerateManagedModule(module: CodegenManagedModule) {
  regeneratingModuleName.value = module.module_name;
  errorMessage.value = '';
  try {
    await applyManagedModule(module, false);
    generateResult.value = await regenerateCodegenFiles({
      module_name: module.module_name,
      overwrite: generateOptions.overwrite,
      register_module: generateOptions.register_module,
      upsert_menu: generateOptions.upsert_menu,
    });
    removeResult.value = null;
    selectedManagedModuleName.value = module.module_name;
    notifySuccess(`模块 ${module.module_name} 已重新生成`);
    await refreshLifecycleData();
  } catch (error) {
    setActionError(error, '按模块重新生成失败');
  } finally {
    regeneratingModuleName.value = '';
  }
}

function openRemoveModal(module: CodegenManagedModule) {
  selectedManagedModuleName.value = module.module_name;
  removeModal.open = true;
  removeModal.module_name = module.module_name;
  removeModal.remove_files = true;
  removeModal.unregister_module = true;
  removeModal.remove_menu = true;
  removeModal.remove_history = false;
  removeModal.remove_lock = true;
}

function closeRemoveModal() {
  if (removing.value) {
    return;
  }
  removeModal.open = false;
}

async function confirmRemoveModule() {
  if (!removeModal.module_name) {
    notifyError('请选择要卸载的模块');
    return;
  }
  removing.value = true;
  errorMessage.value = '';
  try {
    removeResult.value = await removeCodegenModule({
      module_name: removeModal.module_name,
      remove_files: removeModal.remove_files,
      unregister_module: removeModal.unregister_module,
      remove_menu: removeModal.remove_menu,
      remove_history: removeModal.remove_history,
      remove_lock: removeModal.remove_lock,
    });
    diffResult.value = null;
    generateResult.value = null;
    notifySuccess(`模块 ${removeModal.module_name} 卸载流程已执行`);
    removeModal.open = false;
    await refreshLifecycleData();
  } catch (error) {
    setActionError(error, '卸载模块失败');
  } finally {
    removing.value = false;
  }
}

function autoBoolLabel(value: boolean) {
  return value ? '自动：是' : '自动：否';
}

function autoComponentLabel(row: FieldConfigRow) {
  return row.auto_component ? `自动：${row.auto_component}` : '自动推断';
}

function optionsPreviewText(row: FieldConfigRow) {
  if (!row.auto_options.length) {
    return '自动：无';
  }
  return `自动：${row.auto_options.map((item) => `${item.label}:${String(item.value)}`).join(' / ')}`;
}

function defaultValuePreviewText(row: FieldConfigRow) {
  if (row.auto_default_value === undefined || row.auto_default_value === null || row.auto_default_value === '') {
    return '自动：空';
  }
  return `自动：${stringifyEditableValue(row.auto_default_value)}`;
}

function diffHashes(row: { old_hash?: string; new_hash?: string }) {
  const oldHash = row.old_hash ? row.old_hash.slice(0, 8) : '-';
  const newHash = row.new_hash ? row.new_hash.slice(0, 8) : '-';
  return `${oldHash} -> ${newHash}`;
}

onMounted(async () => {
  await Promise.all([loadTables(), loadHistory(), loadModules()]);
});
</script>

<template>
  <section class="page-stack">
    <div class="codegen-layout">
      <aside class="card codegen-sidebar">
        <div class="section-heading compact">
          <div>
            <h3>业务表</h3>
            <p>只列当前可生成的业务表。</p>
          </div>
        </div>
        <div v-if="loadingTables" class="empty-state">加载业务表中...</div>
        <div v-else class="table-list">
          <button
            v-for="item in tables"
            :key="item.table_name"
            type="button"
            class="table-list__item"
            :class="{ active: form.table_name === item.table_name }"
            @click="changeTable(item.table_name)"
          >
            <strong>{{ item.display_name }}</strong>
            <span>{{ item.table_name }}</span>
            <span v-if="item.table_comment">{{ item.table_comment }}</span>
          </button>
        </div>
      </aside>

      <div class="page-stack">
        <article class="card page-card">
          <div class="section-heading">
            <div>
              <h3>代码生成</h3>
              <p>{{ tableHint }}</p>
            </div>
            <div class="table-actions align-end">
              <PermissionButton code="codegen.list">
                <button class="btn secondary" type="button" @click="openExportImportDialog()">
                  载入 Export
                </button>
              </PermissionButton>
              <PermissionButton code="codegen.save">
                <button class="btn secondary" type="button" :disabled="previewing" @click="previewCurrent()">
                  生成 Preview
                </button>
              </PermissionButton>
              <PermissionButton code="codegen.save">
                <button class="btn secondary" type="button" :disabled="diffing" @click="diffCurrent()">
                  查看 Diff
                </button>
              </PermissionButton>
              <PermissionButton code="codegen.save">
                <button class="btn secondary" type="button" :disabled="saving" @click="saveCurrent()">
                  保存历史
                </button>
              </PermissionButton>
              <PermissionButton code="codegen.save">
                <button class="btn" type="button" :disabled="generating" @click="generateCurrent()">
                  生成文件
                </button>
              </PermissionButton>
            </div>
          </div>

          <div v-if="errorMessage" class="error-banner">{{ errorMessage }}</div>

          <div class="form-grid two-columns">
            <FormField label="模块名" required>
              <input v-model.trim="form.module_name" class="input" placeholder="例如 demo_article" />
            </FormField>
            <FormField label="表名" required>
              <input v-model="form.table_name" class="input" readonly />
            </FormField>
            <FormField label="载入来源">
              <div class="tag-list">
                <span class="tag-chip">{{ loadedSourceKindLabel }}</span>
                <span class="tag-chip">source kind: {{ loadedSourceKind }}</span>
              </div>
            </FormField>
            <FormField label="标题覆盖">
              <input
                v-model.trim="form.payload.title"
                class="input"
                :placeholder="`自动：${effectiveTitle}`"
              />
            </FormField>
            <div class="field-groups">
              <label class="checkbox-item">
                <input v-model="generateOptions.overwrite" type="checkbox" />
                <span>允许覆盖生成器文件</span>
              </label>
              <label class="checkbox-item">
                <input v-model="generateOptions.register_module" type="checkbox" />
                <span>重建模块注册文件</span>
              </label>
              <label class="checkbox-item">
                <input v-model="generateOptions.upsert_menu" type="checkbox" />
                <span>同步菜单与权限</span>
              </label>
            </div>
          </div>
        </article>

        <article class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>模块生命周期</h3>
              <p>扫描 `codegen.lock.json`，统一管理 generate、regenerate、remove 的完整链路。</p>
            </div>
          </div>

          <div v-if="loadingModules" class="empty-state">扫描已生成模块中...</div>
          <div v-else-if="!modules.length" class="empty-state">当前还没有 lock 管理的已生成模块。</div>
          <div v-else class="module-grid">
            <section
              v-for="item in modules"
              :key="item.module_name"
              class="module-card inset-card"
              :class="{ active: selectedManagedModuleName === item.module_name }"
            >
              <div class="stack-xs">
                <strong>{{ item.module_name }}</strong>
                <span class="text-muted">{{ item.table_name }}</span>
                <span class="text-muted">{{ item.route_path }}</span>
              </div>

              <div class="module-meta-list">
                <span>生成时间：{{ formatTime(item.generated_at) }}</span>
                <span>模板版本：{{ item.template_version }}</span>
                <span>文件数：{{ item.files.length }}</span>
              </div>

              <div class="tag-list">
                <span v-for="code in item.permission_codes" :key="code" class="tag-chip">
                  {{ code }}
                </span>
              </div>

              <div class="table-actions align-start">
                <PermissionButton code="codegen.list">
                  <button class="btn secondary btn-sm" type="button" @click="selectManagedModule(item.module_name)">
                    查看摘要
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button class="btn secondary btn-sm" type="button" @click="applyManagedModule(item)">
                    载入配置
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.list">
                  <button class="btn secondary btn-sm" type="button" @click="exportManagedModule(item)">
                    导出配置
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button class="btn secondary btn-sm" type="button" @click="viewManagedDiff(item)">
                    查看 Diff
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button
                    class="btn secondary btn-sm"
                    type="button"
                    :disabled="regeneratingModuleName === item.module_name"
                    @click="regenerateManagedModule(item)"
                  >
                    重新生成
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.delete">
                  <button class="btn danger btn-sm" type="button" @click="openRemoveModal(item)">
                    卸载模块
                  </button>
                </PermissionButton>
              </div>
            </section>
          </div>
        </article>

        <article v-if="selectedManagedModule" class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>Lock 摘要</h3>
              <p>把 `codegen.lock.json` 的基础信息、schema 摘要和字段覆盖可视化展示出来。</p>
            </div>
          </div>

          <div class="preview-grid">
            <section class="preview-card inset-card">
              <strong>基础信息</strong>
              <div class="stack-xs">
                <span>模块：{{ selectedManagedModule.module_name }}</span>
                <span>数据表：{{ selectedManagedModule.table_name }}</span>
                <span>路由：{{ selectedManagedModule.route_path }}</span>
                <span>生成时间：{{ formatTime(selectedManagedModule.generated_at) }}</span>
                <span>模板版本：{{ selectedManagedModule.template_version }}</span>
              </div>
            </section>

            <section class="preview-card inset-card">
              <strong>Route / API</strong>
              <div class="stack-xs">
                <span>菜单标题：{{ selectedManagedModule.preview_summary.page.menu_title || selectedManagedModule.module_name }}</span>
                <span>页面：{{ selectedManagedModule.preview_summary.page.page_name }}</span>
                <span>视图文件：{{ selectedManagedModule.preview_summary.page.view_file }}</span>
                <span>List：{{ selectedManagedModule.preview_summary.api.list }}</span>
                <span>Save：{{ selectedManagedModule.preview_summary.api.save }}</span>
                <span>Delete：{{ selectedManagedModule.preview_summary.api.delete }}</span>
                <span>Feature Flags：{{ (selectedManagedModule.preview_summary.page.feature_flags || []).join(', ') || '-' }}</span>
              </div>
            </section>

            <section class="preview-card inset-card">
              <strong>权限码</strong>
              <div class="tag-list">
                <span v-for="code in selectedManagedModule.permission_codes" :key="code" class="tag-chip">
                  {{ code }}
                </span>
              </div>
            </section>

            <section class="preview-card inset-card">
              <strong>Generated Files</strong>
              <pre class="mini-code">{{ prettyJSON(selectedManagedModule.files) }}</pre>
            </section>

            <section class="preview-card preview-card--full inset-card">
              <strong>Field Overrides 摘要</strong>
              <div v-if="selectedOverrideEntries.length" class="override-grid">
                <div v-for="item in selectedOverrideEntries" :key="item.field" class="override-card">
                  <strong>{{ item.field }}</strong>
                  <pre class="mini-code">{{ prettyJSON(item.config) }}</pre>
                </div>
              </div>
              <div v-else class="empty-state">当前 lock 没有字段级 overrides。</div>
            </section>
          </div>
        </article>

        <article class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>字段元数据</h3>
              <p>读取当前 PostgreSQL 字段、注释与默认值。</p>
            </div>
          </div>
          <AppTable :columns="columnTable" :rows="columns" :loading="loadingColumns" empty-text="先选择一张业务表">
            <template #cell-is_nullable="{ value }">
              <span>{{ value ? '是' : '否' }}</span>
            </template>
            <template #cell-is_primary_key="{ value }">
              <span>{{ value ? '是' : '否' }}</span>
            </template>
            <template #cell-column_comment="{ value }">
              <span>{{ value || '-' }}</span>
            </template>
          </AppTable>
        </article>

        <article class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>字段配置表</h3>
              <p>选字段只是第一步，这里管理每个字段的 label、component、placeholder、options 和布尔开关覆盖。</p>
            </div>
          </div>

          <div v-if="!fieldConfigs.length" class="empty-state">先选择一张业务表并生成 preview。</div>
          <div v-else class="app-table codegen-field-config">
            <table class="app-table__inner">
              <thead>
                <tr>
                  <th style="width: 120px">字段</th>
                  <th style="width: 110px">类型</th>
                  <th style="width: 70px">列表</th>
                  <th style="width: 70px">表单</th>
                  <th style="width: 70px">搜索</th>
                  <th style="width: 160px">Label</th>
                  <th style="width: 150px">组件</th>
                  <th style="width: 96px">必填</th>
                  <th style="width: 96px">只读</th>
                  <th style="width: 96px">隐藏</th>
                  <th style="width: 96px">可搜</th>
                  <th style="width: 96px">可排</th>
                  <th style="width: 120px">宽度</th>
                  <th style="width: 220px">Placeholder</th>
                  <th style="width: 220px">Options JSON</th>
                  <th style="width: 180px">Default Value</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in fieldConfigs" :key="row.column_name">
                  <td>
                    <div class="stack-xs">
                      <strong>{{ row.column_name }}</strong>
                      <small class="text-muted">{{ row.column_comment || row.auto_label }}</small>
                    </div>
                  </td>
                  <td>{{ row.data_type }}</td>
                  <td>
                    <input :checked="row.in_list" type="checkbox" @change="changeSelection(row, 'list', ($event.target as HTMLInputElement).checked)" />
                  </td>
                  <td>
                    <input :checked="row.in_form" type="checkbox" @change="changeSelection(row, 'form', ($event.target as HTMLInputElement).checked)" />
                  </td>
                  <td>
                    <input :checked="row.in_search" type="checkbox" @change="changeSelection(row, 'search', ($event.target as HTMLInputElement).checked)" />
                  </td>
                  <td>
                    <input v-model="row.label" class="input" :placeholder="`自动：${row.auto_label}`" />
                  </td>
                  <td>
                    <select v-model="row.component" class="input">
                      <option v-for="item in componentOptions" :key="item.value || 'auto'" :value="item.value">
                        {{ item.value ? item.label : autoComponentLabel(row) }}
                      </option>
                    </select>
                  </td>
                  <td>
                    <select v-model="row.required" class="input">
                      <option value="">{{ autoBoolLabel(row.auto_required) }}</option>
                      <option value="true">是</option>
                      <option value="false">否</option>
                    </select>
                  </td>
                  <td>
                    <select v-model="row.readonly" class="input">
                      <option value="">{{ autoBoolLabel(row.auto_readonly) }}</option>
                      <option value="true">是</option>
                      <option value="false">否</option>
                    </select>
                  </td>
                  <td>
                    <select v-model="row.hidden" class="input">
                      <option value="">{{ autoBoolLabel(row.auto_hidden) }}</option>
                      <option value="true">是</option>
                      <option value="false">否</option>
                    </select>
                  </td>
                  <td>
                    <select v-model="row.searchable" class="input">
                      <option value="">{{ autoBoolLabel(row.auto_searchable) }}</option>
                      <option value="true">是</option>
                      <option value="false">否</option>
                    </select>
                  </td>
                  <td>
                    <select v-model="row.sortable" class="input">
                      <option value="">{{ autoBoolLabel(row.auto_sortable) }}</option>
                      <option value="true">是</option>
                      <option value="false">否</option>
                    </select>
                  </td>
                  <td>
                    <input v-model.trim="row.width" class="input" :placeholder="row.auto_width || '自动'" />
                  </td>
                  <td>
                    <textarea
                      v-model="row.placeholder"
                      class="input textarea codegen-textarea--sm"
                      :placeholder="`自动：${row.auto_placeholder || '-'}`"
                    />
                  </td>
                  <td>
                    <textarea
                      v-model="row.options_text"
                      class="input textarea codegen-textarea--sm"
                      :placeholder="optionsPreviewText(row)"
                    />
                  </td>
                  <td>
                    <textarea
                      v-model="row.default_value_text"
                      class="input textarea codegen-textarea--sm"
                      :placeholder="defaultValuePreviewText(row)"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>

        <article v-if="preview" class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>Preview 方案稿</h3>
              <p>这里展示未来真实生成会使用的 page / api / schema 结构。</p>
            </div>
          </div>

          <div class="preview-grid">
            <section class="preview-card inset-card">
              <strong>页面元信息</strong>
              <div class="stack-xs">
                <span>标题：{{ preview.payload.title || effectiveTitle }}</span>
                <span>菜单标题：{{ preview.page.menu_title || effectiveTitle }}</span>
                <span>路由：{{ preview.page.route_path }}</span>
                <span>页面名：{{ preview.page.page_name }}</span>
                <span>文件：{{ preview.page.view_file }}</span>
                <span>Feature Flags：{{ (preview.page.feature_flags || []).join(', ') || '-' }}</span>
              </div>
            </section>
            <section class="preview-card inset-card">
              <strong>API 元信息</strong>
              <div class="stack-xs">
                <span>模块：{{ preview.api.module_code }}</span>
                <span>List：{{ preview.api.list }}</span>
                <span>Detail：{{ preview.api.detail }}</span>
                <span>Save：{{ preview.api.save }}</span>
                <span>Delete：{{ preview.api.delete }}</span>
              </div>
            </section>
            <section class="preview-card inset-card">
              <strong>字段推断</strong>
              <pre class="mini-code">{{ prettyJSON(preview.inferred_fields) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>表单 Schema</strong>
              <pre class="mini-code">{{ prettyJSON(preview.form_schema) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>列表 Schema</strong>
              <pre class="mini-code">{{ prettyJSON(preview.list_schema) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>搜索 Schema</strong>
              <pre class="mini-code">{{ prettyJSON(preview.search_schema) }}</pre>
            </section>
            <section class="preview-card preview-card--full inset-card">
              <strong>生成说明</strong>
              <ul class="note-list">
                <li v-for="item in preview.notes" :key="item">{{ item }}</li>
              </ul>
            </section>
          </div>
        </article>

        <article v-if="diffResult" class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>Diff 预览</h3>
              <p>生成前先看本次会创建、覆盖还是跳过哪些文件。</p>
            </div>
          </div>

          <div class="preview-grid">
            <section class="preview-card inset-card">
              <strong>将创建</strong>
              <pre class="mini-code">{{ prettyJSON(diffResult.would_create_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>将覆盖</strong>
              <pre class="mini-code">{{ prettyJSON(diffResult.would_overwrite_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>将跳过</strong>
              <pre class="mini-code">{{ prettyJSON(diffResult.would_skip_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>权限码</strong>
              <pre class="mini-code">{{ prettyJSON(diffResult.permission_codes) }}</pre>
            </section>
            <section class="preview-card preview-card--full inset-card">
              <strong>文件 diff 摘要</strong>
              <AppTable :columns="diffTable" :rows="diffResult.per_file_diff_summary">
                <template #cell-status="{ value }">
                  <span class="status-pill" :class="{
                    'is-active': value === 'create',
                    'is-warning': value === 'overwrite',
                    'is-muted': value === 'skip',
                  }">
                    {{ value }}
                  </span>
                </template>
                <template #cell-hashes="{ row }">
                  <code>{{ diffHashes(row) }}</code>
                </template>
                <template #cell-changed_sections="{ value }">
                  <div class="stack-xs">
                    <span v-for="item in value" :key="item">{{ item }}</span>
                  </div>
                </template>
              </AppTable>
            </section>
          </div>
        </article>

        <article v-if="generateResult" class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>生成结果</h3>
              <p>这里是最近一次 generate / regenerate 的真实写盘结果。</p>
            </div>
          </div>

          <div class="preview-grid">
            <section class="preview-card inset-card">
              <strong>新生成文件</strong>
              <pre class="mini-code">{{ prettyJSON(generateResult.generated_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>覆盖文件</strong>
              <pre class="mini-code">{{ prettyJSON(generateResult.overwritten_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>跳过文件</strong>
              <pre class="mini-code">{{ prettyJSON(generateResult.skipped_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>权限与菜单</strong>
              <pre class="mini-code">{{ prettyJSON({
                route_path: generateResult.route_path,
                permission_codes: generateResult.permission_codes,
                menu_records: generateResult.menu_records,
              }) }}</pre>
            </section>
            <section v-if="generateResult.warnings?.length" class="preview-card preview-card--full inset-card">
              <strong>Warnings</strong>
              <pre class="mini-code">{{ prettyJSON(generateResult.warnings) }}</pre>
            </section>
          </div>
        </article>

        <article v-if="removeResult" class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>卸载结果</h3>
              <p>这里展示最近一次 remove 的真实删除、跳过和注册文件重建结果。</p>
            </div>
          </div>

          <div class="preview-grid">
            <section class="preview-card inset-card">
              <strong>已删除文件</strong>
              <pre class="mini-code">{{ prettyJSON(removeResult.removed_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>跳过文件</strong>
              <pre class="mini-code">{{ prettyJSON(removeResult.skipped_files) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>已删除菜单 / 权限</strong>
              <pre class="mini-code">{{ prettyJSON({
                removed_menu_records: removeResult.removed_menu_records,
                removed_role_menu_links: removeResult.removed_role_menu_links,
              }) }}</pre>
            </section>
            <section class="preview-card inset-card">
              <strong>注册文件 / 历史</strong>
              <pre class="mini-code">{{ prettyJSON({
                regenerated_registry_files: removeResult.regenerated_registry_files,
                removed_history_ids: removeResult.removed_history_ids,
              }) }}</pre>
            </section>
            <section v-if="removeResult.warnings?.length" class="preview-card preview-card--full inset-card">
              <strong>Warnings</strong>
              <pre class="mini-code">{{ prettyJSON(removeResult.warnings) }}</pre>
            </section>
          </div>
        </article>

        <article class="card page-card">
          <div class="section-heading compact">
            <div>
              <h3>生成历史</h3>
              <p>历史记录支持载入配置、直接生成、查看 diff 和重新生成。</p>
            </div>
          </div>

          <AppTable :columns="historyTable" :rows="historyRows" :loading="loadingHistory" empty-text="暂无生成历史">
            <template #cell-created_at="{ value }">
              {{ formatTime(value) }}
            </template>
            <template #cell-actions="{ row }">
              <div class="table-actions">
                <PermissionButton code="codegen.save">
                  <button class="btn secondary btn-sm" type="button" @click="applyHistory(row)">载入配置</button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button class="btn secondary btn-sm" type="button" @click="viewHistoryDiff(row)">查看 Diff</button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button class="btn secondary btn-sm" type="button" @click="generateFromHistory(row)">直接生成</button>
                </PermissionButton>
                <PermissionButton code="codegen.save">
                  <button
                    class="btn secondary btn-sm"
                    type="button"
                    :disabled="regeneratingHistoryId === row.id"
                    @click="regenerateFromHistory(row)"
                  >
                    重新生成
                  </button>
                </PermissionButton>
                <PermissionButton code="codegen.delete">
                  <button class="btn danger btn-sm" type="button" @click="removeHistoryRecord(row)">删除</button>
                </PermissionButton>
              </div>
            </template>
          </AppTable>
        </article>
      </div>
    </div>
  </section>

  <AppModal
    :open="removeModal.open"
    title="卸载生成模块"
    width="920px"
    @close="closeRemoveModal"
  >
    <div class="page-stack">
      <div class="hint-banner">
        仅会删除生成器管理的文件；手写文件不会被覆盖或误删。建议先看一次 diff，再执行 remove。
      </div>

      <div class="form-grid two-columns">
        <label class="check-card">
          <input v-model="removeModal.remove_files" type="checkbox" />
          <div>
            <strong>删除模块文件</strong>
            <small>删除 module/model/types/meta、admin api、admin 页面。</small>
          </div>
        </label>
        <label class="check-card">
          <input v-model="removeModal.unregister_module" type="checkbox" />
          <div>
            <strong>重建注册文件</strong>
            <small>重写 `modules_gen.go` 和 `generated/routes.ts`，把模块从注册表移除。</small>
          </div>
        </label>
        <label class="check-card">
          <input v-model="removeModal.remove_menu" type="checkbox" />
          <div>
            <strong>删除菜单与权限</strong>
            <small>删除 `admin_menu` / `admin_role_menu` 里的当前模块菜单和按钮节点。</small>
          </div>
        </label>
        <label class="check-card">
          <input v-model="removeModal.remove_history" type="checkbox" />
          <div>
            <strong>删除生成历史</strong>
            <small>清理 `codegen_history` 里当前模块的记录，默认不勾选。</small>
          </div>
        </label>
        <label class="check-card">
          <input v-model="removeModal.remove_lock" type="checkbox" />
          <div>
            <strong>删除 lock 文件</strong>
            <small>移除 `internal/modules/{module}/codegen.lock.json`。</small>
          </div>
        </label>
      </div>

      <div v-if="removeTargetModule" class="preview-grid">
        <section class="preview-card inset-card">
          <strong>将删除的模块文件</strong>
          <pre class="mini-code">{{ prettyJSON(removeModal.remove_files ? removeTargetFiles.moduleFiles : []) }}</pre>
        </section>
        <section class="preview-card inset-card">
          <strong>将重建的注册文件</strong>
          <pre class="mini-code">{{ prettyJSON(removeModal.unregister_module ? removeTargetFiles.registry : []) }}</pre>
        </section>
        <section class="preview-card inset-card">
          <strong>将删除的菜单与权限</strong>
          <pre class="mini-code">{{ prettyJSON(removeModal.remove_menu ? removeTargetMenuRecords : []) }}</pre>
        </section>
        <section class="preview-card inset-card">
          <strong>将删除的 lock / history</strong>
          <pre class="mini-code">{{ prettyJSON({
            remove_lock: removeModal.remove_lock,
            remove_history: removeModal.remove_history,
            lock_file: removeModal.remove_lock ? `internal/modules/${removeTargetModule.module_name}/codegen.lock.json` : '',
            history_scope: removeModal.remove_history ? removeTargetModule.module_name : '',
          }) }}</pre>
        </section>
      </div>
      <div v-else class="empty-state">未找到对应 lock 摘要，remove 将只按 module_name 尝试清理。</div>
    </div>

    <template #footer>
      <div class="modal-actions">
        <button class="btn secondary" type="button" :disabled="removing" @click="closeRemoveModal">取消</button>
        <button class="btn danger" type="button" :disabled="removing" @click="confirmRemoveModule">
          {{ removing ? '卸载中...' : '确认卸载' }}
        </button>
      </div>
    </template>
  </AppModal>

  <input
    ref="exportFileInput"
    type="file"
    accept="application/json,.json"
    style="display: none"
    @change="handleExportImport"
  />
</template>
