import { request } from './request';
import type { AdminMe, MenuItem } from '@/types';

export function adminLogin(payload: { username: string; password: string }) {
  return request.post('/auth/login', payload);
}

export function adminLogout() {
  return request.post('/auth/logout');
}

export function fetchAdminMe(): Promise<AdminMe> {
  return request.get('/auth/me');
}

export function fetchAdminMenus(): Promise<{ list: MenuItem[] }> {
  return request.get('/auth/menus');
}
