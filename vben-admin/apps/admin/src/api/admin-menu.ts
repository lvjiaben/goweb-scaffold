import { request } from './request';
import type { MenuItem, MenuOption } from '@/types';

export type AdminMenuForm = {
  id?: number;
  parent_id: number;
  name: string;
  title: string;
  path: string;
  component: string;
  menu_type: 'button' | 'menu';
  permission_code: string;
  icon: string;
  sort: number;
  visible: boolean;
  status: number;
};

export function fetchAdminMenus() {
  return request.get<{ list: MenuItem[] }>('/admin_menu/list');
}

export function fetchAdminMenuTree() {
  return request.get<{ list: MenuItem[] }>('/admin_menu/tree');
}

export function fetchAdminMenuOptions() {
  return request.get<{ list: MenuOption[] }>('/admin_menu/options');
}

export function fetchAdminMenuDetail(id: number) {
  return request.get<AdminMenuForm>('/admin_menu/detail', { params: { id } });
}

export function saveAdminMenu(payload: AdminMenuForm) {
  return request.post('/admin_menu/save', payload);
}

export function deleteAdminMenu(id: number) {
  return request.post('/admin_menu/delete', { id });
}
