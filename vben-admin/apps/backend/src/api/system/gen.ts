import { requestClient } from '#/api/request';

export namespace GenApi {
  /** 表信息 */
  export interface TableInfo {
    table_name: string;
    table_comment: string;
  }

  /** 字段配置 */
  export interface FieldConfig {
    column_name: string;
    column_type: string;
    column_comment: string;
    is_nullable: boolean;
    is_primary_key: boolean;
    is_auto_increment: boolean;
    field_name: string;
    field_type: string;
    json_tag: string;
    gorm_tag: string;
    in_create: boolean;
    in_update: boolean;
    is_required: boolean;
    validate_rules: string;
    show_in_table: boolean;
    table_sortable: boolean;
    table_searchable: boolean;
    table_search_type: string;
    table_display_type: string;
    show_in_form: boolean;
    form_component: string;
    form_component_props: Record<string, any>;
    options?: Array<Record<string, any>>;
    search_form_type: string;
    is_operate_field: boolean;
    is_sort_field: boolean;
    is_time_field: boolean;
    is_relation_field: boolean;
    is_enum_field: boolean;
    is_set_field: boolean;
    is_text_field: boolean;
    is_bool_field: boolean;
    is_image_field: boolean;
    is_images_field: boolean;
  }

  /** 生成配置 */
  export interface GenConfig {
    table_name: string;
    table_comment: string;
    module_name: string;
    struct_name: string;
    package_name: string;
    frontend_src_path: string;
    methods: {
      list: boolean;
      create: boolean;
      update: boolean;
      delete: boolean;
      operate: boolean;
    };
    fields: FieldConfig[];
    search_fields: string[];
    operate_fields: string[];
    default_sort_field: string;
    default_sort_order: string;
    menu_config: {
      parent_menu_name: string;
      menu_name: string;
      menu_icon: string;
      menu_sort: number;
    };
  }

  /** 生成的代码 */
  export interface GeneratedCode {
    backend: {
      controller: string;
      service: string;
      model: string;
      validate: string;
      route: string;
    };
    frontend: {
      api: string;
      list_view: string;
      data_ts: string;
      form_vue: string;
    };
    menu: {
      sql: string;
    };
  }

  /** 生成历史 */
  export interface GenHistory {
    id: number;
    table_name: string;
    table_comment: string;
    module_name: string;
    struct_name: string;
    package_name: string;
    frontend_src_path: string;
    config: string;
    created_at: number;
    updated_at: number;
  }
}

function toPretty(value: any) {
  return JSON.stringify(value ?? {}, null, 2);
}

const supportedFormComponents = new Set([
  'Input',
  'Textarea',
  'InputNumber',
  'Select',
  'RadioGroup',
  'Switch',
  'DatePicker',
  'RangePicker',
  'TimePicker',
  'TableSelect',
  'TableSelectMultiple',
  'Upload',
  'IconPicker',
  'JsonTextarea',
]);

const supportedSearchComponents = new Set([
  'Input',
  'InputNumber',
  'Select',
  'RadioGroup',
  'Switch',
  'DatePicker',
  'RangePicker',
  'TableSelect',
  'TableSelectMultiple',
]);

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

function normalizeFormComponent(value?: string) {
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
  return supportedFormComponents.has(normalized) ? normalized : 'Input';
}

function normalizeSearchComponent(value?: string) {
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
  return supportedSearchComponents.has(normalized) ? normalized : 'Input';
}

function normalizeTableDisplayType(value?: string, dataType?: string) {
  const raw = String(value || '').trim();
  const aliasMap: Record<string, string> = {
    editable: 'text',
    id: 'number',
    'boolean-tag': 'bool',
    'json-preview': 'text',
    'option-tag': 'tag',
  };
  const normalized = aliasMap[raw] || raw;
  if (supportedTableDisplayTypes.has(normalized)) {
    return normalized;
  }
  if (['bigint', 'integer', 'int', 'numeric', 'smallint'].includes(String(dataType || '').toLowerCase())) {
    return 'number';
  }
  return 'text';
}

function defaultOptionsForColumn(column: Record<string, any>, component: string) {
  const name = String(column.column_name || '').toLowerCase();
  if (component === 'Switch' || name.startsWith('is_') || name.startsWith('has_') || name === 'enabled') {
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
}

function normalizeField(column: Record<string, any>, preview?: Record<string, any>): GenApi.FieldConfig {
  const listItem = preview?.list_schema?.find?.((item: Record<string, any>) => item.field === column.column_name);
  const formItem = preview?.form_schema?.find?.((item: Record<string, any>) => item.field === column.column_name);
  const searchItem = preview?.search_schema?.find?.((item: Record<string, any>) => item.field === column.column_name);
  const inferred = preview?.inferred_fields?.find?.((item: Record<string, any>) => item.column_name === column.column_name);
  const component = normalizeFormComponent(formItem?.component ?? inferred?.guessed_form_component ?? column.guessed_form_component ?? 'Input');
  const searchComponent = normalizeSearchComponent(searchItem?.component ?? (column.guessed_searchable ? inferred?.guessed_form_component ?? column.guessed_form_component ?? 'Input' : ''));
  const options = formItem?.options ?? searchItem?.options ?? listItem?.options ?? defaultOptionsForColumn(column, component);
  const formComponentProps: Record<string, any> = {
    options,
    placeholder: formItem?.placeholder ?? searchItem?.placeholder ?? '',
  };
  if (component === 'TableSelectMultiple' || searchComponent === 'TableSelectMultiple') {
    formComponentProps.multiple = true;
    formComponentProps.config = {
      labelField: 'name',
      multiple: true,
      valueField: 'id',
    };
  }
  if (component === 'TableSelect' || searchComponent === 'TableSelect') {
    formComponentProps.config = {
      labelField: 'name',
      valueField: 'id',
    };
  }
  return {
    column_comment: column.column_comment ?? '',
    column_name: column.column_name ?? '',
    column_type: column.data_type ?? column.column_type ?? '',
    field_name: column.column_name ?? '',
    field_type: column.data_type ?? '',
    form_component: component,
    form_component_props: formComponentProps,
    gorm_tag: '',
    in_create: true,
    in_update: true,
    is_auto_increment: Boolean(column.column_default?.includes?.('nextval')),
    is_bool_field: ['boolean', 'bool'].includes(column.data_type),
    is_enum_field: Array.isArray(column.options) && column.options.length > 0,
    is_image_field: false,
    is_images_field: false,
    is_nullable: Boolean(column.is_nullable),
    is_operate_field: false,
    is_primary_key: Boolean(column.is_primary_key),
    is_relation_field: false,
    is_required: Boolean(formItem?.required ?? (!column.is_nullable && !column.is_primary_key)),
    is_search_field: false,
    is_set_field: false,
    is_sort_field: Boolean(listItem?.sortable ?? inferred?.guessed_sortable ?? column.guessed_sortable),
    is_text_field: ['text', 'varchar', 'character varying'].includes(column.data_type),
    is_time_field: ['timestamp', 'timestamptz', 'timestamp with time zone'].includes(
      column.data_type,
    ),
    json_tag: column.column_name ?? '',
    options,
    search_form_type: searchComponent,
    show_in_form: Boolean(formItem),
    show_in_table: Boolean(listItem),
    table_display_type: normalizeTableDisplayType(listItem?.display ?? column.guessed_list_display ?? 'text', column.data_type ?? column.column_type),
    table_search_type: searchItem?.operator ?? (column.guessed_searchable ? 'input' : ''),
    table_searchable: Boolean(searchItem?.searchable ?? inferred?.guessed_searchable ?? column.guessed_searchable),
    table_sortable: Boolean(listItem?.sortable ?? inferred?.guessed_sortable ?? column.guessed_sortable),
    validate_rules: '',
  } as GenApi.FieldConfig;
}

function normalizeConfig(
  tableName: string,
  columns: Array<Record<string, any>>,
  tableComment = '',
  preview?: Record<string, any>,
): GenApi.GenConfig {
  const moduleName = tableName;
  return {
    default_sort_field: 'id',
    default_sort_order: 'desc',
    fields: columns.map((column) => normalizeField(column, preview)),
    frontend_src_path: `src/views/${moduleName}`,
    menu_config: {
      menu_icon: 'carbon:table',
      menu_name: tableComment || moduleName,
      menu_sort: 100,
      parent_menu_name: '系统管理',
    },
    methods: {
      create: true,
      delete: true,
      list: true,
      operate: true,
      update: true,
    },
    module_name: moduleName,
    operate_fields: ['id'],
    package_name: moduleName,
    search_fields: (preview?.search_schema ?? columns.filter((item) => item.guessed_searchable))
      .map((item: Record<string, any>) => item.field ?? item.column_name),
    struct_name: moduleName
      .split('_')
      .map((item) => item.slice(0, 1).toUpperCase() + item.slice(1))
      .join(''),
    table_comment: tableComment,
    table_name: tableName,
  };
}

function buildFieldOverrides(fields: GenApi.FieldConfig[]) {
  return Object.fromEntries(
    fields.map((item) => [
      item.column_name,
      {
        component: item.form_component,
        options: item.form_component_props?.options ?? [],
        placeholder: item.form_component_props?.placeholder ?? '',
        required: item.is_required,
        searchable: item.table_searchable,
        sortable: item.table_sortable,
      },
    ]),
  );
}

/**
 * 获取数据库表列表
 */
async function getTableList(params?: { search?: string }) {
  const response = await requestClient.get<{
    list?: GenApi.TableInfo[];
  }>('/system/codegen/tables');
  const keyword = params?.search?.trim().toLowerCase();
  return (response?.list ?? []).filter((item) => {
    if (!keyword) {
      return true;
    }
    return (
      item.table_name.toLowerCase().includes(keyword) ||
      String(item.table_comment ?? '').toLowerCase().includes(keyword)
    );
  });
}

/**
 * 获取表详细信息
 */
async function getTableInfo(params: { table_name: string }) {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/system/codegen/table-columns', { params });
  return response?.list ?? [];
}

/**
 * 获取表的默认配置
 */
async function getTableConfig(params: { table_name: string }) {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>(
    '/system/codegen/table-columns',
    {
      params,
    },
  );
  const tableInfo = (await getTableList()).find(
    (item) => item.table_name === params.table_name,
  );
  const preview = await requestClient.post<any>('/system/codegen/preview', {
    module_name: params.table_name,
    payload: {
      title: tableInfo?.table_comment ?? params.table_name,
    },
    table_name: params.table_name,
  });
  return normalizeConfig(
    params.table_name,
    response?.list ?? [],
    tableInfo?.table_comment ?? '',
    preview,
  );
}

/**
 * 预览生成的代码
 */
async function previewCode(data: { config: GenApi.GenConfig }) {
  const config = data.config;
  const response = await requestClient.post<any>('/system/codegen/preview', {
    module_name: config.module_name,
    payload: {
      field_overrides: buildFieldOverrides(config.fields),
      form_fields: config.fields.filter((item) => item.show_in_form).map((item) => item.column_name),
      list_fields: config.fields.filter((item) => item.show_in_table).map((item) => item.column_name),
      search_fields: config.search_fields ?? [],
      title: config.table_comment || config.menu_config.menu_name,
    },
    table_name: config.table_name,
  });
  return {
    backend: {
      controller: toPretty(response?.api ?? {}),
      model: toPretty(response?.inferred_fields ?? []),
      route: toPretty({
        delete: response?.api?.delete,
        detail: response?.api?.detail,
        list: response?.api?.list,
        save: response?.api?.save,
      }),
      service: toPretty(response?.list_schema ?? []),
      validate: toPretty(response?.form_schema ?? []),
    },
    frontend: {
      api: toPretty(response?.api ?? {}),
      data_ts: toPretty({
        list_schema: response?.list_schema ?? [],
        search_schema: response?.search_schema ?? [],
      }),
      form_vue: toPretty(response?.form_schema ?? []),
      list_view: toPretty(response?.page ?? {}),
    },
    menu: {
      sql: toPretty({
        notes: response?.notes ?? [],
        permission_codes: response?.permission_codes ?? [],
        route_path: response?.page?.route_path,
      }),
    },
  };
}

/**
 * 生成代码
 */
async function generateCode(data: { config: GenApi.GenConfig }) {
  const config = data.config;
  return requestClient.post('/system/codegen/generate', {
    module_name: config.module_name,
    overwrite: true,
    payload: {
      field_overrides: buildFieldOverrides(config.fields),
      form_fields: config.fields.filter((item) => item.show_in_form).map((item) => item.column_name),
      list_fields: config.fields.filter((item) => item.show_in_table).map((item) => item.column_name),
      search_fields: config.search_fields ?? [],
      title: config.table_comment || config.menu_config.menu_name,
    },
    register_module: true,
    table_name: config.table_name,
    upsert_menu: true,
  });
}

/**
 * 获取生成历史
 */
async function getHistory() {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/system/codegen/list');
  return (response?.list ?? []).map((item) => ({
    config: toPretty(item.payload ?? {}),
    created_at: Date.parse(item.created_at ?? '') || 0,
    frontend_src_path: item.route_path ?? '',
    id: Number(item.id ?? 0),
    module_name: item.module_name ?? '',
    package_name: item.module_name ?? '',
    struct_name: item.module_name ?? '',
    table_comment: item.title ?? '',
    table_name: item.table_name ?? '',
    updated_at: Date.parse(item.updated_at ?? item.created_at ?? '') || 0,
  }));
}

/**
 * 删除生成的代码
 */
async function deleteGenerated(data: { id: number }) {
  return requestClient.post('/system/codegen/delete', data);
}

/**
 * 下载生成的代码
 */
async function downloadCode(data: { config: GenApi.GenConfig }): Promise<Blob> {
  const config = data.config;
  const exportPayload = await requestClient.get<any>('/system/codegen/export', {
    params: { module_name: config.module_name },
  });
  return new Blob([JSON.stringify(exportPayload, null, 2)], {
    type: 'application/json',
  });
}

export {
  deleteGenerated,
  downloadCode,
  generateCode,
  getHistory,
  getTableConfig,
  getTableInfo,
  getTableList,
  previewCode,
};
