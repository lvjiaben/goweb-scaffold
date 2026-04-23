import { requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
    captcha?:any;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

type BackendMenu = {
  children?: BackendMenu[];
  component?: string;
  icon?: string;
  id: number;
  name: string;
  path: string;
  title: string;
};

type BackendMe = {
  access_codes?: string[];
  id: number;
  is_super?: boolean;
  nickname?: string;
  role_ids?: number[];
  username: string;
};

const viewModules = import.meta.glob('../../views/**/*.vue');
const viewKeys = new Set(
  Object.keys(viewModules).map((key) =>
    key.replace('../../views/', '').replace(/\.vue$/, ''),
  ),
);

function hasView(viewPath: string) {
  return viewKeys.has(viewPath);
}

function toRouteName(value: string) {
  return value
    .split(/[\\/_-]+/)
    .filter(Boolean)
    .map((item) => item.charAt(0).toUpperCase() + item.slice(1))
    .join('');
}

function resolveViewPath(item: BackendMenu) {
  const directMap: Record<string, string> = {
    '/dashboard': 'home/index',
    '/system/admin-user': 'admin/admin/list',
    '/system/admin-role': 'admin/role/list',
    '/system/admin-menu': 'admin/menu/list',
    '/system/system-config': 'system/config/index',
    '/system/attachment': 'system/attachment/index',
    '/system/codegen': 'system/gen/list',
    '/user/list': 'user/list',
  };

  if (directMap[item.path]) {
    return directMap[item.path]!;
  }

  const candidates = [item.component].filter(Boolean) as string[];
  if (item.path.startsWith('/system/')) {
    const tail = item.path.replace('/system/', '');
    candidates.push(`${tail.replace(/-/g, '_')}/list`);
  }
  if (item.path.startsWith('/')) {
    candidates.push(`${item.path.slice(1)}/index`);
    candidates.push(item.path.slice(1));
  }

  return candidates.find((candidate) => hasView(candidate));
}

function normalizeMenu(item: BackendMenu): any | null {
  const children = (item.children ?? [])
    .map((child) => normalizeMenu(child))
    .filter(Boolean);

  if (item.path === '/system') {
    return {
      component: 'BasicLayout',
      meta: {
        icon: item.icon || 'mdi:cog-outline',
        order: 10,
        title: item.title,
      },
      name: toRouteName(item.name || 'system'),
      path: item.path,
      redirect: children[0]?.path,
      children,
    };
  }

  const viewPath = resolveViewPath(item);
  if (!viewPath) {
    return children[0] ?? null;
  }

  return {
    component: viewPath,
    meta: {
      icon: item.icon || undefined,
      title: item.title,
    },
    name: toRouteName(item.name || viewPath),
    path: item.path,
    children,
  };
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  const result = await requestClient.post<any>('/auth/login', data);
  return {
    accessToken: result?.token ?? '',
  };
}

/**
 * 获取所有菜单（Vben动态路由使用）
 */
export async function getAllMenusApi() {
  const result = await requestClient.get<{ list?: BackendMenu[] }>('/auth/menus');
  return (result?.list ?? [])
    .map((item) => normalizeMenu(item))
    .filter(Boolean);
}

/**
 * 获取用户信息
 */
export async function getUserInfoApi() {
  const result = await requestClient.get<BackendMe>('/auth/me');
  return {
    avatar: '',
    desc: result?.is_super ? 'super-admin' : 'admin',
    homePath: '/dashboard',
    realName: result?.nickname || result?.username || '',
    roles: (result?.role_ids ?? []).map((id) => `role:${id}`),
    userId: result?.id,
    username: result?.username || '',
  };
}

/**
 * 更新个人资料
 */
export async function updateProfileApi(data: { avatar: string; email: string }) {
  return requestClient.post('/auth/profile', data);
}

/**
 * 修改密码
 */
export async function changePasswordApi(data: {
  old_password: string;
  new_password: string;
}) {
  return requestClient.post('/auth/password', data);
}

/**
 * 获取操作日志
 */
export async function getOperationLogApi(params?: {
  page?: number;
  page_size?: number;
}) {
  return requestClient.get('/auth/log', { params });
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return <AuthApi.RefreshTokenResult>{
    data: '',
    status: 500,
  };
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return requestClient.post('/auth/logout', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi() {
  const result = await requestClient.get<BackendMe>('/auth/me');
  return result?.access_codes ?? [];
}
