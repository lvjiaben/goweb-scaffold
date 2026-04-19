import { request } from './request';
import type { AdminUserItem, Paginated, RoleOption } from '@/types';

export type AdminUserForm = {
  id?: number;
  username: string;
  password?: string;
  nickname: string;
  status: number;
  is_super: boolean;
  role_ids: number[];
};

export function fetchAdminUsers(params: { keyword?: string; page?: number; page_size?: number }) {
  return request.get<Paginated<AdminUserItem>>('/admin_user/list', { params });
}

export function fetchAdminUserDetail(id: number) {
  return request.get<AdminUserForm>('/admin_user/detail', { params: { id } });
}

export function saveAdminUser(payload: AdminUserForm) {
  return request.post('/admin_user/save', payload);
}

export function deleteAdminUser(id: number) {
  return request.post('/admin_user/delete', { id });
}

export function fetchRoleOptions() {
  return request.get<{ list: RoleOption[] }>('/admin_role/options');
}
