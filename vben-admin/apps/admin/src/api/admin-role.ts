import { request } from './request';
import type { MenuOption, Paginated, RoleItem, RoleOption } from '@/types';

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

export function fetchRoleOptionList() {
  return request.get<{ list: RoleOption[] }>('/admin_role/options');
}

export function fetchMenuTreeOptions() {
  return request.get<{ list: MenuOption[] }>('/admin_menu/options');
}
