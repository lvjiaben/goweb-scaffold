import axios from 'axios';

import { adminForbiddenEvent, adminSessionExpiredEvent } from '@/events';

export const adminTokenKey = 'goweb_admin_token';
let handlingUnauthorized = false;

export const request = axios.create({
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
    const status = Number(error.response?.status || 0);
    const message = error.response?.data?.msg || error.message || '请求失败';
    if (status === 401) {
      const hadToken = Boolean(localStorage.getItem(adminTokenKey));
      localStorage.removeItem(adminTokenKey);
      if (hadToken && !handlingUnauthorized) {
        handlingUnauthorized = true;
        window.dispatchEvent(new CustomEvent(adminSessionExpiredEvent, { detail: { message } }));
        window.setTimeout(() => {
          handlingUnauthorized = false;
        }, 300);
      }
    }
    if (status === 403) {
      window.dispatchEvent(new CustomEvent(adminForbiddenEvent, { detail: { message } }));
    }
    return Promise.reject(new Error(message));
  },
);
