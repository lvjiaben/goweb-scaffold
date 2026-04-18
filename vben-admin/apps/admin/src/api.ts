import axios from 'axios';

import type { MenuItem } from './types';

export const adminTokenKey = 'goweb_admin_token';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem(adminTokenKey);
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (response) => {
    const body = response.data;
    if (typeof body?.code !== 'number') {
      return body;
    }
    if (body.code !== 0) {
      return Promise.reject(new Error(body.msg || '请求失败'));
    }
    return body.data;
  },
  (error) => {
    const message = error.response?.data?.msg || error.message || '请求失败';
    return Promise.reject(new Error(message));
  },
);

export function adminLogin(payload: { username: string; password: string }) {
  return request.post('/auth/login', payload);
}

export function adminLogout() {
  return request.post('/auth/logout');
}

export function fetchAdminMe() {
  return request.get('/auth/me');
}

export function fetchAdminMenus(): Promise<{ list: MenuItem[] }> {
  return request.get('/auth/menus');
}

export function listModule(moduleName: string) {
  return request.get(`/${moduleName}/list`);
}

export function detailModule(moduleName: string, id: number) {
  return request.get(`/${moduleName}/detail`, { params: { id } });
}

export function saveModule(moduleName: string, payload: Record<string, unknown>) {
  return request.post(`/${moduleName}/save`, payload);
}

export function deleteModule(moduleName: string, payload: { id?: number; ids?: number[] }) {
  return request.post(`/${moduleName}/delete`, payload);
}

export function uploadAttachment(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request.post('/attachment/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}
