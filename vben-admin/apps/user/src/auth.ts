import { reactive } from 'vue';

import { fetchProfile, userLogin, userLogout, userRegister, userTokenKey } from './api';

type UserProfile = {
  id: number;
  username: string;
  nickname: string;
  email: string;
  mobile: string;
};

export const userState = reactive<{
  token: string;
  profile: UserProfile | null;
  bootstrapped: boolean;
}>({
  token: localStorage.getItem(userTokenKey) || '',
  profile: null,
  bootstrapped: false,
});

export function setUserToken(token: string) {
  userState.token = token;
  localStorage.setItem(userTokenKey, token);
}

export function clearUserSession() {
  userState.token = '';
  userState.profile = null;
  userState.bootstrapped = false;
  localStorage.removeItem(userTokenKey);
}

export async function bootstrapUserSession() {
  if (!userState.token) {
    clearUserSession();
    return false;
  }
  try {
    userState.profile = await fetchProfile();
    userState.bootstrapped = true;
    return true;
  } catch {
    clearUserSession();
    return false;
  }
}

export async function loginUser(username: string, password: string) {
  const result = await userLogin({ username, password });
  setUserToken(result.token);
  return bootstrapUserSession();
}

export async function registerUser(username: string, password: string, nickname: string) {
  await userRegister({ username, password, nickname });
  return loginUser(username, password);
}

export async function logoutUser() {
  try {
    await userLogout();
  } catch {
    // ignore logout failures
  }
  clearUserSession();
}
