<script lang="ts" setup>
import { ref, computed } from 'vue';

import { IconPicker, Page, VbenButton } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { $t } from '@vben/locales';

import {
  Card,
  Checkbox,
  Divider,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  SelectOption,
  Space,
  Steps,
  Table,
  Tabs,
  Tag,
  Textarea,
} from 'ant-design-vue';

import {
  deleteGenerated,
  downloadCode,
  generateCode,
  getHistory,
  getTableConfig,
  getTableList,
  previewCode,
} from '#/api/system/gen';
import type { GenApi } from '#/api/system/gen';

import PreviewDrawer from './modules/preview-drawer.vue';

// 当前标签页
const activeTab = ref('generator');

// 当前步骤
const currentStep = ref(0);

// 表列表
const tables = ref<GenApi.TableInfo[]>([]);
const tableLoading = ref(false);
const tableSearch = ref('');

// 选中的表
const selectedTable = ref<string>('');

// 配置
const config = ref<GenApi.GenConfig>({
  table_name: '',
  table_comment: '',
  module_name: '',
  struct_name: '',
  package_name: '',
  frontend_src_path: '',
  methods: {
    list: true,
    create: true,
    update: true,
    delete: true,
    operate: false,
  },
  fields: [],
  search_fields: [],
  operate_fields: [],
  default_sort_field: 'id',
  default_sort_order: 'desc',
  menu_config: {
    parent_menu_name: 'AutoPlay',
    menu_name: '',
    menu_icon: '',
    menu_sort: 50,
  },
});

// 预览代码
const previewVisible = ref(false);
const previewData = ref<GenApi.GeneratedCode | null>(null);

// 历史记录
const histories = ref<GenApi.GenHistory[]>([]);
const historyLoading = ref(false);

// 组件配置Modal
const componentPropsModalVisible = ref(false);
const currentField = ref<GenApi.FieldConfig | null>(null);
const componentPropsForm = ref({
  options: '',
  api: '',
  labelField: 'label',
  valueField: 'value',
  multiple: false,
});

const supportedTableDisplayTypes = new Set([
  'text',
  'tag',
  'datetime',
  'image',
  'link',
  'links',
  'bool',
  'number',
]);

const configurableComponents = new Set([
  'Select',
  'RadioGroup',
  'Switch',
  'TableSelect',
  'TableSelectMultiple',
]);

const normalizeTableDisplayType = (value?: string, columnType?: string) => {
  const raw = String(value || '').trim();
  const aliasMap: Record<string, string> = {
    editable: 'text',
    id: 'number',
    'json-preview': 'text',
    'boolean-tag': 'bool',
    'option-tag': 'tag',
  };
  const normalized = aliasMap[raw] || raw;
  if (supportedTableDisplayTypes.has(normalized)) {
    return normalized;
  }
  if (['bigint', 'integer', 'int', 'numeric', 'smallint'].includes(String(columnType || '').toLowerCase())) {
    return 'number';
  }
  return 'text';
};

const normalizeFormComponent = (value?: string) => {
  const raw = String(value || '').trim();
  const aliasMap: Record<string, string> = {
    hidden: 'Input',
    radio: 'RadioGroup',
    select: 'Select',
    switch: 'Switch',
    textarea: 'Textarea',
    'datetime-picker': 'DatePicker',
    'json-editor': 'JsonTextarea',
    'number-input': 'InputNumber',
    'readonly-datetime': 'DatePicker',
    'readonly-text': 'Input',
    'table-select': 'TableSelect',
    'table-select-multiple': 'TableSelectMultiple',
    'text-input': 'Input',
    upload: 'Upload',
  };
  const normalized = aliasMap[raw] || raw;
  const supported = new Set(formComponentOptions.map((item) => item.value));
  return supported.has(normalized) ? normalized : 'Input';
};

const normalizeSearchFormType = (value?: string) => {
  const raw = String(value || '').trim();
  if (!raw || raw === 'hidden') {
    return '';
  }
  const aliasMap: Record<string, string> = {
    radio: 'RadioGroup',
    select: 'Select',
    switch: 'Switch',
    'datetime-picker': 'DatePicker',
    'datetime-range': 'RangePicker',
    'number-input': 'InputNumber',
    'readonly-datetime': 'DatePicker',
    'table-select': 'TableSelect',
    'table-select-multiple': 'TableSelectMultiple',
    'text-input': 'Input',
  };
  const normalized = aliasMap[raw] || raw;
  const supported = new Set(searchFormTypeOptions.map((item) => item.value));
  return supported.has(normalized) ? normalized : 'Input';
};

const defaultOptionsForField = (field: GenApi.FieldConfig) => {
  const name = field.column_name.toLowerCase();
  if (name.startsWith('is_') || name.startsWith('has_') || name === 'enabled' || field.form_component === 'Switch') {
    return [
      { label: '否', value: false },
      { label: '是', value: true },
    ];
  }
  if (name === 'state') {
    return [
      { label: '关闭', value: 0 },
      { label: '开启', value: 1 },
    ];
  }
  if (name === 'status' || name.endsWith('_status')) {
    return [
      { label: '禁用', value: 0 },
      { label: '启用', value: 1 },
    ];
  }
  return [];
};

const normalizeFieldVisualConfig = (field: GenApi.FieldConfig) => {
  field.form_component = normalizeFormComponent(field.form_component);
  field.search_form_type = normalizeSearchFormType(field.search_form_type);
  field.table_display_type = normalizeTableDisplayType(field.table_display_type, field.column_type);

  const props = field.form_component_props || {};
  const options = props.options || (field as any).options || defaultOptionsForField(field);
  if (Array.isArray(options) && options.length > 0) {
    props.options = options;
    (field as any).options = options;
  }

  if (field.form_component === 'TableSelectMultiple' || field.search_form_type === 'TableSelectMultiple') {
    props.multiple = true;
    props.config = {
      ...(props.config || {}),
      labelField: props.config?.labelField || props.labelField || 'name',
      multiple: true,
      valueField: props.config?.valueField || props.valueField || 'id',
    };
  }

  if ((field.form_component === 'TableSelect' || field.search_form_type === 'TableSelect') && !props.config) {
    props.config = {
      api: props.api || '',
      labelField: props.labelField || 'name',
      valueField: props.valueField || 'id',
    };
  }

  field.form_component_props = props;
  return field;
};

const hasComponentConfig = (field: GenApi.FieldConfig) => {
  normalizeFieldVisualConfig(field);
  const props = field.form_component_props || {};
  return (
    configurableComponents.has(field.form_component) ||
    configurableComponents.has(field.search_form_type) ||
    Object.keys(props).length > 0 ||
    Array.isArray((field as any).options)
  );
};

const shouldShowOptionsEditor = (field?: GenApi.FieldConfig | null) => {
  if (!field) return false;
  return (
    ['Select', 'RadioGroup', 'Switch'].includes(field.form_component) ||
    ['Select', 'RadioGroup', 'Switch'].includes(field.search_form_type) ||
    Array.isArray(field.form_component_props?.options) ||
    Array.isArray((field as any).options)
  );
};

const shouldShowTableSelectEditor = (field?: GenApi.FieldConfig | null) => {
  if (!field) return false;
  return (
    ['TableSelect', 'TableSelectMultiple'].includes(field.form_component) ||
    ['TableSelect', 'TableSelectMultiple'].includes(field.search_form_type) ||
    Boolean(field.form_component_props?.config?.api || field.form_component_props?.api)
  );
};

// 加载表列表
const loadTables = async () => {
  tableLoading.value = true;
  try {
    const res = await getTableList({ search: tableSearch.value });
    tables.value = res;
  } catch (error) {
    message.error('加载表列表失败');
  } finally {
    tableLoading.value = false;
  }
};

// 加载历史记录
const loadHistory = async () => {
  historyLoading.value = true;
  try {
    const res = await getHistory();
    histories.value = res;
  } catch (error) {
    message.error('加载历史记录失败');
  } finally {
    historyLoading.value = false;
  }
};

// 选择表
const selectTable = async (tableName: string) => {
  selectedTable.value = tableName;
  try {
    const res = await getTableConfig({ table_name: tableName });
    // 自动为 _id/_ids 结尾的字段设置表格显示类型
    // 同时设置图片、排序、富文本、text类型、desc/description/content结尾字段默认不可搜索
    res.fields.forEach((field) => {
      normalizeFieldVisualConfig(field);
      const columnName = field.column_name.toLowerCase();
      // 设置表格显示类型
      if (!field.table_display_type || field.table_display_type === 'text') {
        if (field.column_name.endsWith('_ids')) {
          field.table_display_type = 'links';
        } else if (field.column_name.endsWith('_id') && !field.is_primary_key) {
          field.table_display_type = 'link';
        }
      }
      // 图片、排序、富文本、text类型、desc/description/content结尾字段默认不可搜索
      if (field.is_image_field || field.is_images_field || field.is_sort_field ||
          field.is_text_field ||
          columnName.endsWith('desc') || columnName.endsWith('description') ||
          columnName.endsWith('content')) {
        field.table_searchable = false;
      }
    });
    config.value = res;
    currentStep.value = 1;
  } catch (error) {
    message.error('加载表配置失败');
  }
};

// 表格显示类型选项
const tableDisplayTypeOptions = [
  { label: '普通文本', value: 'text' },
  { label: 'Tag标签', value: 'tag' },
  { label: '时间格式化', value: 'datetime' },
  { label: '图片显示', value: 'image' },
  { label: '参数跳转', value: 'link' },
  { label: '多参数跳转', value: 'links' },
  { label: '布尔显示', value: 'bool' },
  { label: '数字', value: 'number' },
];

// 搜索表单类型选项（与表单组件一致）
const searchFormTypeOptions = [
  { label: 'Input', value: 'Input' },
  { label: 'InputNumber', value: 'InputNumber' },
  { label: 'DatePicker', value: 'DatePicker' },
  { label: 'RangePicker', value: 'RangePicker' },
  { label: 'Switch', value: 'Switch' },
  { label: 'Select', value: 'Select' },
  { label: 'RadioGroup', value: 'RadioGroup' },
  { label: 'TableSelect', value: 'TableSelect' },
  { label: 'TableSelectMultiple', value: 'TableSelectMultiple' },
];

// 表单组件选项
const formComponentOptions = [
  { label: 'Input', value: 'Input' },
  { label: 'Textarea', value: 'Textarea' },
  { label: 'InputNumber', value: 'InputNumber' },
  { label: 'Select', value: 'Select' },
  { label: 'RadioGroup', value: 'RadioGroup' },
  { label: 'Switch', value: 'Switch' },
  { label: 'DatePicker', value: 'DatePicker' },
  { label: 'RangePicker', value: 'RangePicker' },
  { label: 'TimePicker', value: 'TimePicker' },
  { label: 'TableSelect', value: 'TableSelect' },
  { label: 'TableSelectMultiple', value: 'TableSelectMultiple' },
  { label: 'Upload', value: 'Upload' },
  { label: 'IconPicker', value: 'IconPicker' },
  { label: 'JsonTextarea', value: 'JsonTextarea' },
];

// 字段表格列
const fieldColumns = computed(() => [
  { title: $t('system.gen.fieldConfig.columnName'), dataIndex: 'column_name', width: 120, fixed: 'left' as const },
  { title: $t('system.gen.fieldConfig.columnType'), dataIndex: 'column_type', width: 100 },
  { title: $t('system.gen.fieldConfig.columnComment'), dataIndex: 'column_comment', width: 120 },
  { title: $t('system.gen.fieldConfig.showInTable'), dataIndex: 'show_in_table', width: 80 },
  { title: $t('system.gen.fieldConfig.tableDisplayType'), dataIndex: 'table_display_type', width: 120 },
  { title: $t('system.gen.fieldConfig.tableSortable'), dataIndex: 'table_sortable', width: 70 },
  { title: $t('system.gen.fieldConfig.tableSearchable'), dataIndex: 'table_searchable', width: 70 },
  { title: $t('system.gen.fieldConfig.searchFormType'), dataIndex: 'search_form_type', width: 120 },
  { title: $t('system.gen.fieldConfig.showInForm'), dataIndex: 'show_in_form', width: 80 },
  { title: $t('system.gen.fieldConfig.formComponent'), dataIndex: 'form_component', width: 130 },
  { title: $t('system.gen.fieldConfig.componentProps'), dataIndex: 'component_props', width: 100 },
  { title: $t('system.gen.fieldConfig.isRequired'), dataIndex: 'is_required', width: 60 },
]);

// 历史记录表格列
const historyColumns = computed(() => [
  { title: $t('system.gen.history.id'), dataIndex: 'id', width: 60 },
  { title: $t('system.gen.history.tableName'), dataIndex: 'table_name', width: 150 },
  { title: $t('system.gen.history.tableComment'), dataIndex: 'table_comment', width: 150 },
  { title: $t('system.gen.history.moduleName'), dataIndex: 'module_name', width: 120 },
  { title: $t('system.gen.history.moduleName'), dataIndex: 'struct_name', width: 120 },
  { title: $t('system.gen.history.methods'), dataIndex: 'methods', width: 200 },
  { title: $t('system.gen.history.createdAt'), dataIndex: 'created_at', width: 180 },
  { title: $t('system.gen.history.action'), dataIndex: 'action', width: 150, fixed: 'right' as const },
]);

// 搜索字段选项
const searchFieldOptions = computed(() => {
  return config.value.fields
    .filter((f) => f.field_type === 'string' && !f.is_primary_key)
    .map((f) => ({
      label: `${f.column_name} (${f.column_comment})`,
      value: f.column_name,
    }));
});

// Operate 字段选项（所有字段都可选择）
const operateFieldOptions = computed(() => {
  return config.value.fields.map((f) => ({
    label: `${f.column_name} (${f.column_comment})`,
    value: f.column_name,
  }));
});

// 打开组件配置Modal
const openComponentPropsModal = (field: GenApi.FieldConfig) => {
  currentField.value = normalizeFieldVisualConfig(field);
  // 解析已有配置
  const props = currentField.value.form_component_props || {};
  // 兼容新的 config 格式和旧格式
  const configData = props.config || props;
  const options = props.options || (currentField.value as any).options || defaultOptionsForField(currentField.value);
  componentPropsForm.value = {
    options: Array.isArray(options) && options.length > 0 ? JSON.stringify(options, null, 2) : '',
    api: configData.api || '',
    labelField: configData.labelField || 'name',
    valueField: configData.valueField || 'id',
    multiple: Boolean(configData.multiple || props.multiple || currentField.value.form_component === 'TableSelectMultiple' || currentField.value.search_form_type === 'TableSelectMultiple'),
  };
  componentPropsModalVisible.value = true;
};

// 保存组件配置
const saveComponentProps = () => {
  if (!currentField.value) return;

  const props: any = {};

  // Select 组件配置
  if (componentPropsForm.value.options) {
    try {
      props.options = JSON.parse(componentPropsForm.value.options);
    } catch (e) {
      message.error('Options 格式错误，请输入有效的 JSON 数组');
      return;
    }
  }

  // TableSelect 组件配置 - 使用 config 属性
  if (componentPropsForm.value.api) {
    props.config = {
      api: componentPropsForm.value.api,
      labelField: componentPropsForm.value.labelField || 'name',
      multiple: componentPropsForm.value.multiple,
      valueField: componentPropsForm.value.valueField || 'id',
    };
  }
  if (componentPropsForm.value.multiple) {
    props.multiple = true;
  }

  currentField.value.form_component_props = props;
  if (props.options) {
    (currentField.value as any).options = props.options;
  }
  componentPropsModalVisible.value = false;
  message.success('配置已保存');
};

// 预览
const handlePreview = async () => {
  try {
    const res = await previewCode({ config: config.value });
    previewData.value = res;
    previewVisible.value = true;
  } catch (error) {
    message.error('预览失败');
  }
};

// 生成
const handleGenerate = async () => {
  try {
    await generateCode({ config: config.value });
    message.success('生成成功');
    // 重置
    currentStep.value = 0;
    selectedTable.value = '';
    // 刷新历史记录
    loadHistory();
  } catch (error) {
    message.error('生成失败');
  }
};

// 下载代码
const handleDownload = async () => {
  try {
    const blob = await downloadCode({ config: config.value });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${config.value.module_name}_generated_code.zip`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
    message.success('代码下载成功！');
  } catch (error) {
    message.error('代码下载失败');
  }
};

// 删除生成的代码
const handleDelete = (record: GenApi.GenHistory) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除 ${record.table_name} 的生成代码吗？此操作将删除所有生成的文件！`,
    okText: '确定',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteGenerated({ id: record.id });
        message.success('删除成功！');
        loadHistory();
      } catch (error) {
        message.error('删除失败');
      }
    },
  });
};

// 复用历史配置
const handleReuse = (record: GenApi.GenHistory) => {
  try {
    // 解析配置
    const savedConfig = JSON.parse(record.config);
    // 设置表名
    selectedTable.value = record.table_name;
    // 回显配置
    config.value = savedConfig;
    // 切换到代码生成Tab和第二步
    activeTab.value = 'generator';
    currentStep.value = 1;
    message.success(`已加载 ${record.table_name} 的配置`);
  } catch (error) {
    message.error('配置解析失败');
  }
};

// 格式化时间
const formatTime = (timestamp: number) => {
  return new Date(timestamp * 1000).toLocaleString('zh-CN');
};

// 初始化
loadTables();
loadHistory();
</script>

<template>
  <Page :description="$t('system.gen.description')" :title="$t('system.gen.title')">
    <Card>
      <Tabs v-model:activeKey="activeTab" destroyInactiveTabPane>
        <Tabs.TabPane key="generator" :tab="$t('system.gen.tabs.generator')">
          <Steps :current="currentStep" class="mb-6">
            <Steps.Step :title="$t('system.gen.steps.selectTable')" />
            <Steps.Step :title="$t('system.gen.steps.configFields')" />
            <Steps.Step :title="$t('system.gen.steps.configMethods')" />
            <Steps.Step :title="$t('system.gen.steps.generateCode')" />
          </Steps>

      <!-- 步骤 1: 选择表 -->
      <div v-if="currentStep === 0">
        <div class="mb-4">
          <Input
            v-model:value="tableSearch"
            :placeholder="$t('system.gen.table.searchPlaceholder')"
            style="width: 300px"
            @pressEnter="loadTables"
          />
          <VbenButton class="ml-2" @click="loadTables"> {{ $t('system.gen.table.search') }} </VbenButton>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          <Card
            v-for="table in tables"
            :key="table.table_name"
            :hoverable="true"
            class="cursor-pointer"
            @click="selectTable(table.table_name)"
          >
            <div class="text-lg font-semibold">{{ table.table_name }}</div>
            <div class="text-gray-500">{{ table.table_comment }}</div>
          </Card>
        </div>
      </div>

      <!-- 步骤 2: 配置字段 -->
      <div v-if="currentStep === 1">
        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.basicConfig.title') }}</h3>
          <div class="grid grid-cols-1 gap-4">
            <div>
              <label class="mb-1 block">{{ $t('system.gen.basicConfig.frontendSrcPath') }}</label>
              <Input
                v-model:value="config.frontend_src_path"
                placeholder="vben-admin/apps/web-antd/src"
              />
              <div class="mt-1 text-sm text-gray-500">
                {{ $t('system.gen.basicConfig.frontendSrcPathTip') }}
              </div>
            </div>
          </div>
        </div>

        <Divider />

        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.fieldConfig.title') }}</h3>
          <Table
            :columns="fieldColumns"
            :data-source="config.fields"
            :pagination="false"
            :scroll="{ x: 1200, y: 400 }"
            row-key="column_name"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'show_in_table'">
                <Checkbox v-model:checked="record.show_in_table" />
              </template>
              <template v-else-if="column.dataIndex === 'table_display_type'">
                <Select
                  v-model:value="record.table_display_type"
                  :options="tableDisplayTypeOptions"
                  size="small"
                  style="width: 100%"
                />
              </template>
              <template v-else-if="column.dataIndex === 'table_sortable'">
                <Checkbox v-model:checked="record.table_sortable" />
              </template>
              <template v-else-if="column.dataIndex === 'table_searchable'">
                <Checkbox v-model:checked="record.table_searchable" />
              </template>
              <template v-else-if="column.dataIndex === 'search_form_type'">
                <Select
                  v-model:value="record.search_form_type"
                  :options="searchFormTypeOptions"
                  size="small"
                  style="width: 100%"
                />
              </template>
              <template v-else-if="column.dataIndex === 'show_in_form'">
                <Checkbox v-model:checked="record.show_in_form" />
              </template>
              <template v-else-if="column.dataIndex === 'form_component'">
                <Select
                  v-model:value="record.form_component"
                  :options="formComponentOptions"
                  size="small"
                  style="width: 100%"
                />
              </template>
              <template v-else-if="column.dataIndex === 'component_props'">
                <VbenButton
                  :disabled="!hasComponentConfig(record as GenApi.FieldConfig)"
                  size="small"
                  @click="openComponentPropsModal(record as GenApi.FieldConfig)"
                >
                  {{ hasComponentConfig(record as GenApi.FieldConfig) ? $t('system.gen.fieldConfig.config') : '无配置' }}
                </VbenButton>
              </template>
              <template v-else-if="column.dataIndex === 'is_required'">
                <Checkbox v-model:checked="record.is_required" />
              </template>
            </template>
          </Table>
        </div>

        <Divider />

        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.searchConfig.title') }}</h3>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="mb-1 block">{{ $t('system.gen.searchConfig.searchFieldsLabel') }}</label>
              <Select
                v-model:value="config.search_fields"
                :options="searchFieldOptions"
                mode="multiple"
                :placeholder="$t('system.gen.searchConfig.searchFieldsPlaceholder')"
                style="width: 100%"
              />
            </div>
            <div>
              <label class="mb-1 block">{{ $t('system.gen.searchConfig.operateFieldsLabel') }}</label>
              <Select
                v-model:value="config.operate_fields"
                :options="operateFieldOptions"
                mode="multiple"
                :placeholder="$t('system.gen.searchConfig.operateFieldsPlaceholder')"
                style="width: 100%"
              />
            </div>
          </div>
        </div>

        <div class="flex justify-between">
          <VbenButton @click="currentStep = 0"> {{ $t('system.gen.nav.prev') }} </VbenButton>
          <VbenButton type="primary" @click="currentStep = 2">
            {{ $t('system.gen.nav.next') }}
          </VbenButton>
        </div>
      </div>

      <!-- 步骤 3: 配置方法 -->
      <div v-if="currentStep === 2">
        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.methods.title') }}</h3>
          <Space direction="vertical" size="large">
            <div>
              <Checkbox :checked="true" disabled> {{ $t('system.gen.methods.list') }} </Checkbox>
            </div>
            <div>
              <Checkbox v-model:checked="config.methods.create">
                {{ $t('system.gen.methods.create') }}
              </Checkbox>
            </div>
            <div>
              <Checkbox v-model:checked="config.methods.update">
                {{ $t('system.gen.methods.update') }}
              </Checkbox>
            </div>
            <div>
              <Checkbox v-model:checked="config.methods.delete">
                {{ $t('system.gen.methods.delete') }}
              </Checkbox>
            </div>
            <div>
              <Checkbox
                v-model:checked="config.methods.operate"
                :disabled="config.operate_fields.length === 0"
              >
                {{ $t('system.gen.methods.operate') }}
                <Tag v-if="config.operate_fields.length === 0" color="orange">
                  {{ $t('system.gen.methods.operateDisabledTip') }}
                </Tag>
              </Checkbox>
            </div>
          </Space>
        </div>

        <Divider />

        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.sortConfig.title') }}</h3>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="mb-1 block">{{ $t('system.gen.sortConfig.defaultSortField') }}</label>
              <Input v-model:value="config.default_sort_field" :placeholder="$t('system.gen.sortConfig.defaultSortFieldPlaceholder')" />
            </div>
            <div>
              <label class="mb-1 block">{{ $t('system.gen.sortConfig.defaultSortOrder') }}</label>
              <Select v-model:value="config.default_sort_order" style="width: 100%">
                <SelectOption value="desc">{{ $t('system.gen.sortConfig.desc') }}</SelectOption>
                <SelectOption value="asc">{{ $t('system.gen.sortConfig.asc') }}</SelectOption>
              </Select>
            </div>
          </div>
        </div>

        <div class="mb-6">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.menuConfig.title') }}</h3>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="mb-1 block">{{ $t('system.gen.menuConfig.parentMenuName') }}</label>
              <Input v-model:value="config.menu_config.parent_menu_name" />
            </div>
            <div>
              <label class="mb-1 block">{{ $t('system.gen.menuConfig.menuName') }}</label>
              <Input v-model:value="config.menu_config.menu_name" />
            </div>
            <div>
              <label class="mb-1 block">{{ $t('system.gen.menuConfig.menuIcon') }}</label>
              <IconPicker v-model="config.menu_config.menu_icon" />
            </div>
            <div>
              <label class="mb-1 block">{{ $t('system.gen.menuConfig.menuSort') }}</label>
              <InputNumber
                v-model:value="config.menu_config.menu_sort"
                :min="0"
                style="width: 100%"
              />
            </div>
          </div>
        </div>

        <div class="flex justify-between">
          <VbenButton @click="currentStep = 1"> {{ $t('system.gen.nav.prev') }} </VbenButton>
          <VbenButton type="primary" @click="currentStep = 3">
            {{ $t('system.gen.nav.next') }}
          </VbenButton>
        </div>
      </div>

      <!-- 步骤 4: 生成代码 -->
      <div v-if="currentStep === 3">
        <div class="mb-6 text-center">
          <h3 class="mb-4 text-lg font-semibold">{{ $t('system.gen.generate.readyTitle') }}</h3>
          <p class="mb-4 text-gray-500">
            {{ $t('system.gen.generate.tableName') }}：{{ config.table_name }} ({{ config.table_comment }})
          </p>
          <p class="mb-4 text-gray-500">
            {{ $t('system.gen.generate.moduleName') }}：{{ config.module_name }}
          </p>
          <p class="mb-4 text-gray-500">
            {{ $t('system.gen.generate.generatedMethods') }}：
            <Tag v-if="config.methods.list" color="blue">List</Tag>
            <Tag v-if="config.methods.create" color="green">Create</Tag>
            <Tag v-if="config.methods.update" color="orange">Update</Tag>
            <Tag v-if="config.methods.delete" color="red">Delete</Tag>
            <Tag v-if="config.methods.operate" color="purple">Operate</Tag>
          </p>
        </div>

        <div class="flex justify-center gap-4">
          <VbenButton @click="currentStep = 2"> {{ $t('system.gen.nav.prev') }} </VbenButton>
          <VbenButton @click="handlePreview">
            <template #icon>
              <IconifyIcon icon="mdi:eye" class="mr-1" />
            </template>
            {{ $t('system.gen.generate.preview') }}
          </VbenButton>
          <VbenButton @click="handleDownload">
            <template #icon>
              <IconifyIcon icon="mdi:download" class="mr-1" />
            </template>
            {{ $t('system.gen.generate.download') }}
          </VbenButton>
          <VbenButton type="primary" @click="handleGenerate">
            <template #icon>
              <IconifyIcon icon="mdi:play" class="mr-1" />
            </template>
            {{ $t('system.gen.generate.generate') }}
          </VbenButton>
        </div>
      </div>
        </Tabs.TabPane>

        <!-- 历史记录标签页 -->
        <Tabs.TabPane key="history" :tab="$t('system.gen.tabs.history')">
          <Table
            :columns="historyColumns"
            :data-source="histories"
            :loading="historyLoading"
            row-key="id"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.dataIndex === 'created_at'">
                {{ formatTime(record.created_at) }}
              </template>
              <template v-else-if="column.dataIndex === 'methods'">
                <template v-if="record.config">
                  <Tag
                    v-for="(value, key) in (JSON.parse(record.config).methods as Record<string, boolean>)"
                    v-show="value"
                    :key="key"
                    :color="
                      key === 'list'
                        ? 'blue'
                        : key === 'create'
                          ? 'green'
                          : key === 'update'
                            ? 'orange'
                            : key === 'delete'
                              ? 'red'
                              : 'purple'
                    "
                  >
                    {{ key }}
                  </Tag>
                </template>
              </template>
              <template v-else-if="column.dataIndex === 'action'">
                <Space>
                  <VbenButton @click="handleReuse(record as GenApi.GenHistory)">
                    {{ $t('system.gen.history.reuse') }}
                  </VbenButton>
                  <VbenButton danger @click="handleDelete(record as GenApi.GenHistory)">
                    {{ $t('system.gen.history.delete') }}
                  </VbenButton>
                </Space>
              </template>
            </template>
          </Table>
        </Tabs.TabPane>
      </Tabs>
    </Card>

    <!-- 预览抽屉 -->
    <PreviewDrawer
      v-model:visible="previewVisible"
      :preview-data="previewData"
    />

    <!-- 组件配置Modal -->
    <Modal
      v-model:open="componentPropsModalVisible"
      :title="$t('system.gen.componentConfig.title')"
      width="600px"
      @ok="saveComponentProps"
    >
      <div class="space-y-4">
        <div v-if="shouldShowOptionsEditor(currentField)">
          <label class="mb-1 block">{{ $t('system.gen.componentConfig.options') }}</label>
          <Textarea
            v-model:value="componentPropsForm.options"
            :rows="6"
            placeholder='[{"label": "选项1", "value": "1"}, {"label": "选项2", "value": "2"}]'
          />
          <div class="mt-1 text-xs text-gray-500">
            {{ $t('system.gen.componentConfig.optionsTip') }}：[{"label": "启用", "value": "1"}, {"label": "禁用", "value": "0"}]
          </div>
        </div>

        <div v-if="shouldShowTableSelectEditor(currentField)">
          <div class="mb-3">
            <label class="mb-1 block">{{ $t('system.gen.componentConfig.apiPath') }}</label>
            <Input
              v-model:value="componentPropsForm.api"
              placeholder="/api/user/list"
            />
            <div class="mt-1 text-xs text-gray-500">
              {{ $t('system.gen.componentConfig.apiPathTip') }}
            </div>
          </div>

          <div class="mb-3">
            <label class="mb-1 block">{{ $t('system.gen.componentConfig.labelField') }}</label>
            <Input
              v-model:value="componentPropsForm.labelField"
              placeholder="name"
            />
            <div class="mt-1 text-xs text-gray-500">
              {{ $t('system.gen.componentConfig.labelFieldTip') }}
            </div>
          </div>

          <div>
            <label class="mb-1 block">{{ $t('system.gen.componentConfig.valueField') }}</label>
            <Input
              v-model:value="componentPropsForm.valueField"
              placeholder="id"
            />
            <div class="mt-1 text-xs text-gray-500">
              {{ $t('system.gen.componentConfig.valueFieldTip') }}
            </div>
          </div>

          <div class="mt-3">
            <Checkbox v-model:checked="componentPropsForm.multiple">
              多选
            </Checkbox>
          </div>
        </div>

        <div v-if="!shouldShowOptionsEditor(currentField) && !shouldShowTableSelectEditor(currentField)" class="text-sm text-gray-500">
          当前字段没有可配置组件参数。
        </div>
      </div>
    </Modal>
  </Page>
</template>
