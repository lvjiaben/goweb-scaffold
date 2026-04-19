import { request } from './request';
import type { CodegenColumn, CodegenHistoryItem, CodegenPreview, CodegenTableInfo } from '@/types';

export type CodegenPayload = {
  module_name: string;
  table_name: string;
  payload: {
    list_fields: string[];
    form_fields: string[];
    search_fields: string[];
  };
};

export function fetchCodegenHistory() {
  return request.get<{ list: CodegenHistoryItem[] }>('/codegen/list');
}

export function fetchCodegenTables() {
  return request.get<{ list: CodegenTableInfo[] }>('/codegen/tables');
}

export function fetchCodegenTableColumns(tableName: string) {
  return request.get<{ list: CodegenColumn[] }>('/codegen/table-columns', {
    params: { table_name: tableName },
  });
}

export function fetchCodegenPreview(payload: CodegenPayload) {
  return request.post<CodegenPreview>('/codegen/preview', payload);
}

export function saveCodegenHistory(payload: CodegenPayload) {
  return request.post('/codegen/save', payload);
}

export function deleteCodegenHistory(id: number) {
  return request.post('/codegen/delete', { id });
}
