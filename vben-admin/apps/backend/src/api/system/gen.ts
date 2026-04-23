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

function normalizeField(column: Record<string, any>): GenApi.FieldConfig {
  return {
    column_comment: column.column_comment ?? '',
    column_name: column.column_name ?? '',
    column_type: column.data_type ?? column.column_type ?? '',
    field_name: column.column_name ?? '',
    field_type: column.data_type ?? '',
    form_component: column.guessed_form_component ?? 'Input',
    form_component_props: {},
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
    is_required: !column.is_nullable && !column.is_primary_key,
    is_search_field: false,
    is_set_field: false,
    is_sort_field: Boolean(column.guessed_sortable),
    is_text_field: ['text', 'varchar', 'character varying'].includes(column.data_type),
    is_time_field: ['timestamp', 'timestamptz', 'timestamp with time zone'].includes(
      column.data_type,
    ),
    json_tag: column.column_name ?? '',
    search_form_type: column.guessed_searchable ? 'Input' : '',
    show_in_form: true,
    show_in_table: column.guessed_list_display !== false,
    table_display_type: column.guessed_list_display ?? 'text',
    table_search_type: column.guessed_searchable ? 'input' : '',
    table_searchable: Boolean(column.guessed_searchable),
    table_sortable: Boolean(column.guessed_sortable),
    validate_rules: '',
  } as GenApi.FieldConfig;
}

function normalizeConfig(
  tableName: string,
  columns: Array<Record<string, any>>,
  tableComment = '',
): GenApi.GenConfig {
  const moduleName = tableName;
  return {
    default_sort_field: 'id',
    default_sort_order: 'desc',
    fields: columns.map(normalizeField),
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
    search_fields: columns
      .filter((item) => item.guessed_searchable)
      .map((item) => item.column_name),
    struct_name: moduleName
      .split('_')
      .map((item) => item.slice(0, 1).toUpperCase() + item.slice(1))
      .join(''),
    table_comment: tableComment,
    table_name: tableName,
  };
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
  return normalizeConfig(
    params.table_name,
    response?.list ?? [],
    tableInfo?.table_comment ?? '',
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
      field_overrides: {},
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
      field_overrides: {},
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
