export type TableColumn = {
  key: string;
  title: string;
  width?: string;
  align?: 'left' | 'center' | 'right';
};

export type NoticeType = 'error' | 'info' | 'success' | 'warning';

export type Paginated<T> = {
  list: T[];
  total: number;
  page: number;
  page_size: number;
};

export interface MenuItem {
  id: number;
  parent_id: number;
  name: string;
  title: string;
  path: string;
  component: string;
  menu_type: string;
  icon: string;
  sort: number;
  permission_code: string;
  visible?: boolean;
  status?: number;
  created_at?: string;
  updated_at?: string;
  children?: MenuItem[];
}

export interface MenuOption {
  label: string;
  value: number;
  menu_type: string;
  children?: MenuOption[];
}

export type FlatMenuItem = MenuItem & { depth: number };

export interface AdminMe {
  id: number;
  username: string;
  nickname: string;
  is_super: boolean;
  role_ids: number[];
  access_codes: string[];
}

export interface RoleOption {
  label: string;
  value: number;
  code: string;
}

export interface AdminUserItem {
  id: number;
  username: string;
  nickname: string;
  status: number;
  is_super: boolean;
  role_ids: number[];
  role_names: string[];
  created_at: string;
}

export interface RoleItem {
  id: number;
  name: string;
  code: string;
  status: number;
  created_at: string;
}

export interface SystemConfigItem {
  id: number;
  config_key: string;
  config_name: string;
  config_value: unknown;
  remark: string;
  created_at: string;
}

export interface AttachmentItem {
  id: number;
  original_name: string;
  saved_name: string;
  url: string;
  file_path: string;
  file_ext: string;
  mime_type: string;
  file_size: number;
  uploader_kind: string;
  uploader_id: number;
  created_at: string;
}

export interface CodegenTableInfo {
  table_name: string;
  display_name: string;
  table_comment?: string;
}

export interface CodegenColumn {
  column_name: string;
  data_type: string;
  is_nullable: boolean;
  column_default: string;
  ordinal_position: number;
  is_primary_key: boolean;
  column_comment?: string;
  table_comment?: string;
}

export interface CodegenHistoryItem {
  id: number;
  module_name: string;
  table_name: string;
  status: string;
  payload: unknown;
  remark: string;
  created_at: string;
}

export interface CodegenInferredField {
  column_name: string;
  data_type: string;
  is_nullable: boolean;
  is_primary_key: boolean;
  column_comment?: string;
  guessed_label: string;
  guessed_form_component: string;
  guessed_list_display: string;
  guessed_searchable: boolean;
  guessed_sortable: boolean;
}

export interface CodegenFieldOption {
  label: string;
  value: unknown;
}

export interface CodegenFieldOverride {
  label?: string;
  component?: string;
  placeholder?: string;
  required?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  sortable?: boolean;
  searchable?: boolean;
  width?: string;
  options?: CodegenFieldOption[];
  default_value?: unknown;
}

export interface CodegenSchemaItem {
  field: string;
  label: string;
  component: string;
  display?: string;
  operator?: string;
  required?: boolean;
  readonly?: boolean;
  hidden?: boolean;
  searchable?: boolean;
  sortable?: boolean;
  placeholder?: string;
  width?: string;
  options?: CodegenFieldOption[];
  default_value?: unknown;
}

export interface CodegenPayloadBody {
  list_fields: string[];
  form_fields: string[];
  search_fields: string[];
  title?: string;
  field_overrides?: Record<string, CodegenFieldOverride>;
}

export interface CodegenPreview {
  module_name: string;
  table_name: string;
  table_comment?: string;
  page: {
    route_path: string;
    page_name: string;
    view_file: string;
    menu_title?: string;
    feature_flags?: string[];
  };
  api: {
    module_code: string;
    list: string;
    detail: string;
    save: string;
    delete: string;
  };
  inferred_fields: CodegenInferredField[];
  form_schema: CodegenSchemaItem[];
  list_schema: CodegenSchemaItem[];
  search_schema: CodegenSchemaItem[];
  payload: CodegenPayloadBody;
  notes: string[];
}

export interface CodegenGenerateResult {
  generated_files: string[];
  overwritten_files: string[];
  skipped_files: string[];
  module_name: string;
  route_path: string;
  permission_codes: string[];
  menu_records: Array<{
    id: number;
    name: string;
    title: string;
    path?: string;
    menu_type: string;
    permission_code?: string;
  }>;
  warnings: string[];
}

export interface CodegenMenuRecord {
  id: number;
  name: string;
  title: string;
  path?: string;
  menu_type: string;
  permission_code?: string;
}

export interface CodegenDiffFileSummary {
  path: string;
  status: string;
  changed_sections: string[];
  old_hash?: string;
  new_hash?: string;
}

export interface CodegenDiffResult {
  would_create_files: string[];
  would_overwrite_files: string[];
  would_skip_files: string[];
  per_file_diff_summary: CodegenDiffFileSummary[];
  module_name: string;
  route_path: string;
  permission_codes: string[];
  warnings: string[];
}

export interface CodegenPreviewSummary {
  table_comment?: string;
  page: {
    route_path: string;
    page_name: string;
    view_file: string;
    menu_title?: string;
    feature_flags?: string[];
  };
  api: {
    module_code: string;
    list: string;
    detail: string;
    save: string;
    delete: string;
  };
  inferred_fields: CodegenInferredField[];
  form_schema: CodegenSchemaItem[];
  list_schema: CodegenSchemaItem[];
  search_schema: CodegenSchemaItem[];
}

export interface CodegenManagedModule {
  module_name: string;
  table_name: string;
  generated_at: string;
  template_version: string;
  route_path: string;
  permission_codes: string[];
  files: string[];
  payload: CodegenPayloadBody;
  preview_summary: CodegenPreviewSummary;
}

export type CodegenSourceKind = 'direct' | 'payload' | 'export' | 'lock' | 'history';

export interface CodegenExportFile {
  generated_by: string;
  format: string;
  version: string;
  module_name: string;
  table_name: string;
  template_version?: string;
  payload: CodegenPayloadBody;
  preview_summary: CodegenPreviewSummary;
  permission_codes: string[];
  route_path: string;
}

export interface CodegenRemovePayload {
  module_name: string;
  remove_files: boolean;
  unregister_module: boolean;
  remove_menu: boolean;
  remove_history: boolean;
  remove_lock: boolean;
}

export interface CodegenRemoveResult {
  module_name: string;
  removed_files: string[];
  skipped_files: string[];
  removed_menu_records: CodegenMenuRecord[];
  removed_role_menu_links: number;
  removed_history_ids: number[];
  regenerated_registry_files: string[];
  warnings: string[];
}
