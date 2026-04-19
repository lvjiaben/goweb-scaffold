import axios from 'axios';

export const adminTokenKey = 'goweb_admin_token';

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
    const message = error.response?.data?.msg || error.message || '请求失败';
    return Promise.reject(new Error(message));
  },
);
