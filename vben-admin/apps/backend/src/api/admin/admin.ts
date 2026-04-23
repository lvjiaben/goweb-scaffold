import { requestClient } from '#/api/request';

export namespace AdminAdminApi {
  export interface Admin {
    [key: string]: any;
    id: number;
    username: string;
    realname: string;
    status: number;
    is_super: number;
    created_at: number;
    updated_at: number;
    role_ids?: number[];
    role_names?: string[];
    password?: string;
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
    list: Admin[];
    total: number;
    page: number;
    limit: number;
  }
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

function normalizeAdmin(row: Record<string, any>): AdminAdminApi.Admin {
  const roleNames = Array.isArray(row.role_names)
    ? row.role_names
    : typeof row.role_names === 'string'
      ? row.role_names
          .split(',')
          .map((item: string) => item.trim())
          .filter(Boolean)
      : [];
  return {
    created_at: toUnixTimestamp(row.created_at),
    id: Number(row.id ?? 0),
    is_super: Number(row.is_super ?? (Number(row.id ?? 0) === 1 ? 1 : 0)),
    password: '',
    realname: row.realname ?? row.nickname ?? '',
    role_ids: Array.isArray(row.role_ids)
      ? row.role_ids.map((item: unknown) => Number(item))
      : [],
    role_names: roleNames,
    status: Number(row.status ?? 1),
    updated_at: toUnixTimestamp(row.updated_at),
    username: row.username ?? '',
  };
}

async function getAdminList(params: AdminAdminApi.ListParams) {
  const response = await requestClient.get<{
    limit?: number;
    list?: Array<Record<string, any>>;
    page?: number;
    total?: number;
  }>('/admin/user/list', { params });
  return {
    limit: Number(response?.limit ?? params.page_size ?? 10),
    list: (response?.list ?? []).map(normalizeAdmin),
    page: Number(response?.page ?? params.page ?? 1),
    total: Number(response?.total ?? 0),
  } satisfies AdminAdminApi.ListResponse;
}

async function getAdminDetail(id: number) {
  const response = await requestClient.get<Record<string, any>>(
    '/admin/user/detail',
    { params: { id } },
  );
  return normalizeAdmin(response ?? {});
}

async function saveAdmin(
  id: number,
  data: Omit<
    AdminAdminApi.Admin,
    'created_at' | 'id' | 'is_super' | 'role_names' | 'updated_at'
  >,
) {
  const payload: Record<string, any> = {
    nickname: data.realname,
    password: data.password,
    role_ids: data.role_ids ?? [],
    status: Number(data.status ?? 1),
    username: data.username,
  };
  if (id > 0) {
    payload.id = id;
  }
  if (!payload.password) {
    delete payload.password;
  }
  return requestClient.post('/admin/user/save', payload);
}

async function deleteAdmin(data: { ids: number[] }) {
  return requestClient.post('/admin/user/delete', data);
}

export { deleteAdmin, getAdminDetail, getAdminList, saveAdmin };
