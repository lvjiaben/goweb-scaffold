import { request } from './request';
import { menuTreeToOptions } from '@/helpers';
import type { MenuItem, MenuOption, Paginated, RoleItem } from '@/types';

export type AdminRoleForm = {
  id?: number;
  name: string;
  code: string;
  status: number;
  menu_ids: number[];
};

export function fetchAdminRoles(params?: { keyword?: string; page?: number; page_size?: number }) {
  return request.get<Paginated<RoleItem>>('/admin_role/list', { params });
}

export function fetchAdminRoleDetail(id: number) {
  return request.get<AdminRoleForm>('/admin_role/detail', { params: { id } });
}

export function saveAdminRole(payload: AdminRoleForm) {
  return request.post('/admin_role/save', payload);
}

export function deleteAdminRole(id: number) {
  return request.post('/admin_role/delete', { id });
}

export async function fetchMenuTreeOptions() {
  const result = await request.get<{ list: MenuItem[] }>('/admin_menu/tree');
  return {
    list: menuTreeToOptions(result.list || []),
  };
}
