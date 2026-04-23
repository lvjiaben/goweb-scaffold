import { requestClient } from '#/api/request';

export namespace AdminRoleApi {
  export interface AdminRole {
    [key: string]: any;
    id: number;
    name: string;
    code: string;
    status: number;
    created_at: number;
    updated_at: number;
    menu_ids?: number[];
  }

  export interface ListParams {
    page: number;
    page_size: number;
    search?: string;
    filter?: string;
    sort_by?: string;
    sort_order?: 'asc' | 'desc';
  }

  export interface ListResponse {
    list: AdminRole[];
    total: number;
    page: number;
    limit: number;
  }
}

export interface RoleOption {
  id: number;
  name: string;
}

function toUnixTimestamp(value?: null | number | string) {
  if (!value) {
    return 0;
  }
  if (typeof value === 'number') {
    return value > 1_000_000_000_000 ? Math.floor(value / 1000) : value;
  }
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? 0 : Math.floor(timestamp / 1000);
}

function normalizeRole(row: Record<string, any>): AdminRoleApi.AdminRole {
  return {
    code: row.code ?? row.description ?? '',
    created_at: toUnixTimestamp(row.created_at),
    id: Number(row.id ?? 0),
    menu_ids: Array.isArray(row.menu_ids)
      ? row.menu_ids.map((item: unknown) => Number(item))
      : [],
    name: row.name ?? '',
    status: Number(row.status ?? 1),
    updated_at: toUnixTimestamp(row.updated_at),
  };
}

function normalizeMenuTree(row: Record<string, any>): Record<string, any> {
  return {
    children: Array.isArray(row.children)
      ? row.children.map((item: Record<string, any>) => normalizeMenuTree(item))
      : [],
    icon: row.icon ?? '',
    id: Number(row.id ?? row.value ?? 0),
    name: row.title ?? row.label ?? row.name ?? '',
    path: row.path ?? '',
    permission: row.permission_code ?? row.permission ?? '',
    type: row.menu_type ?? row.type ?? 'menu',
  };
}

async function getRoleList(params: AdminRoleApi.ListParams) {
  const response = await requestClient.get<{
    limit?: number;
    list?: Array<Record<string, any>>;
    page?: number;
    total?: number;
  }>('/admin/role/list', { params });
  return {
    limit: Number(response?.limit ?? params.page_size ?? 10),
    list: (response?.list ?? []).map(normalizeRole),
    page: Number(response?.page ?? params.page ?? 1),
    total: Number(response?.total ?? 0),
  } satisfies AdminRoleApi.ListResponse;
}

async function getRoleDetail(id: number) {
  const response = await requestClient.get<Record<string, any>>(
    '/admin/role/detail',
    { params: { id } },
  );
  return normalizeRole(response ?? {});
}

async function getRoleOptions() {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/admin/role/options');
  return (response?.list ?? []).map((item) => ({
    id: Number(item.value ?? item.id ?? 0),
    name: item.label ?? item.name ?? '',
  })) satisfies RoleOption[];
}

async function getMenuTree() {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/admin/menu/tree');
  return (response?.list ?? []).map((item) => normalizeMenuTree(item));
}

async function saveRole(
  id: number,
  data: Omit<AdminRoleApi.AdminRole, 'created_at' | 'id' | 'updated_at'>,
) {
  const payload: Record<string, any> = {
    code: data.code,
    menu_ids: data.menu_ids ?? [],
    name: data.name,
    status: Number(data.status ?? 1),
  };
  if (id > 0) {
    payload.id = id;
  }
  return requestClient.post('/admin/role/save', payload);
}

async function deleteRole(data: { ids: number[] }) {
  return requestClient.post('/admin/role/delete', data);
}

export {
  deleteRole,
  getMenuTree,
  getRoleDetail,
  getRoleList,
  getRoleOptions,
  saveRole,
};
