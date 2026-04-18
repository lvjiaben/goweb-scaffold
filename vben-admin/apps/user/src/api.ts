import axios from 'axios';

export const userTokenKey = 'goweb_user_token';

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem(userTokenKey);
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

export function userLogin(payload: { username: string; password: string }) {
  return request.post('/auth/login', payload);
}

export function userRegister(payload: { username: string; password: string; nickname: string }) {
  return request.post('/auth/register', payload);
}

export function userLogout() {
  return request.post('/auth/logout');
}

export function fetchProfile() {
  return request.get('/user/profile');
}

export function saveProfile(payload: { nickname: string; email: string; mobile: string }) {
  return request.post('/user/profile/save', payload);
}

export function changePassword(payload: { old_password: string; new_password: string }) {
  return request.post('/user/password/change', payload);
}
