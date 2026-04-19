import { request } from './request';
import type {
  CodegenColumn,
  CodegenDiffResult,
  CodegenGenerateResult,
  CodegenHistoryItem,
  CodegenPayloadBody,
  CodegenPreview,
  CodegenTableInfo,
} from '@/types';

export type CodegenPayload = {
  module_name: string;
  table_name: string;
  payload: CodegenPayloadBody;
};

export type CodegenGeneratePayload = CodegenPayload & {
  overwrite: boolean;
  register_module: boolean;
  upsert_menu: boolean;
};

export type CodegenRegeneratePayload = {
  module_name?: string;
  history_id?: number;
  overwrite: boolean;
  register_module: boolean;
  upsert_menu: boolean;
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

export function fetchCodegenDiff(payload: CodegenGeneratePayload) {
  return request.post<CodegenDiffResult>('/codegen/diff', payload);
}

export function generateCodegenFiles(payload: CodegenGeneratePayload) {
  return request.post<CodegenGenerateResult>('/codegen/generate', payload);
}

export function regenerateCodegenFiles(payload: CodegenRegeneratePayload) {
  return request.post<CodegenGenerateResult>('/codegen/regenerate', payload);
}

export function saveCodegenHistory(payload: CodegenPayload) {
  return request.post('/codegen/save', payload);
}

export function deleteCodegenHistory(id: number) {
  return request.post('/codegen/delete', { id });
}
