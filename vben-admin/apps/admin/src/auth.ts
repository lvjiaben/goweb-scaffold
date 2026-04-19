import { reactive } from 'vue';

import { adminLogin, adminLogout, fetchAdminMe, fetchAdminMenus } from '@/api/auth';
import { adminTokenKey } from '@/api/request';
import type { AdminMe, MenuItem } from '@/types';

type AdminState = {
  token: string;
  user: AdminMe | null;
  menus: MenuItem[];
  bootstrapped: boolean;
};

export const adminState = reactive<AdminState>({
  token: localStorage.getItem(adminTokenKey) || '',
  user: null,
  menus: [],
  bootstrapped: false,
});

export function setAdminToken(token: string) {
  adminState.token = token;
  localStorage.setItem(adminTokenKey, token);
}

export function clearAdminSession() {
  adminState.token = '';
  adminState.user = null;
  adminState.menus = [];
  adminState.bootstrapped = false;
  localStorage.removeItem(adminTokenKey);
}

export async function bootstrapAdminSession() {
  if (!adminState.token) {
    clearAdminSession();
    return false;
  }

  try {
    const [user, menuPayload] = await Promise.all([fetchAdminMe(), fetchAdminMenus()]);
    adminState.user = user;
    adminState.menus = menuPayload.list || [];
    adminState.bootstrapped = true;
    return true;
  } catch {
    clearAdminSession();
    return false;
  }
}

export async function loginAndLoad(username: string, password: string) {
  const result = await adminLogin({ username, password });
  setAdminToken(result.token);
  return bootstrapAdminSession();
}

export async function logoutAndClear() {
  try {
    await adminLogout();
  } catch {
    // ignore expired session logout
  }
  clearAdminSession();
}

export function hasAccess(code: string) {
  if (!code) {
    return true;
  }
  return Boolean(adminState.user?.access_codes.includes(code));
}
