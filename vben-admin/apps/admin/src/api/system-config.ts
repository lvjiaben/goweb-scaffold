import { request } from './request';
import type { Paginated, SystemConfigItem } from '@/types';

export type SystemConfigForm = {
  id?: number;
  config_key: string;
  config_name: string;
  config_value: unknown;
  remark: string;
};

export function fetchSystemConfigs(params?: { keyword?: string; page?: number; page_size?: number }) {
  return request.get<Paginated<SystemConfigItem>>('/system_config/list', { params });
}

export function fetchSystemConfigDetail(id: number) {
  return request.get<SystemConfigForm>('/system_config/detail', { params: { id } });
}

export function saveSystemConfig(payload: SystemConfigForm) {
  return request.post('/system_config/save', payload);
}
