import { requestClient } from '#/api/request';

export namespace AdminMenuApi {
  export const MenuTypes = ['menu', 'button'] as const;

  export interface AdminMenu {
    [key: string]: any;
    children?: AdminMenu[];
    component?: string;
    created_at?: number;
    icon?: string;
    id: number;
    name: string;
    path: string;
    permission?: string;
    pid: number;
    sort: number;
    status: number;
    title: string;
    type: (typeof MenuTypes)[number];
    updated_at?: number;
    visible: number;
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
    list: AdminMenu[];
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

function normalizeMenu(row: Record<string, any>): AdminMenuApi.AdminMenu {
  return {
    children: Array.isArray(row.children)
      ? row.children.map((item: Record<string, any>) => normalizeMenu(item))
      : [],
    component: row.component ?? '',
    created_at: toUnixTimestamp(row.created_at),
    icon: row.icon ?? '',
    id: Number(row.id ?? row.value ?? 0),
    name: row.name ?? row.title ?? row.label ?? '',
    path: row.path ?? '',
    permission: row.permission_code ?? row.permission ?? '',
    pid: Number(row.parent_id ?? row.pid ?? 0),
    sort: Number(row.sort ?? 0),
    status: Number(row.status ?? 1),
    title: row.title ?? row.label ?? row.name ?? '',
    type: row.menu_type ?? row.type ?? 'menu',
    updated_at: toUnixTimestamp(row.updated_at),
    visible: Number(
      typeof row.visible === 'boolean' ? (row.visible ? 1 : 0) : row.visible ?? 1,
    ),
  };
}

async function getMenuList(params: AdminMenuApi.ListParams) {
  const response = await requestClient.get<{
    limit?: number;
    list?: Array<Record<string, any>>;
    page?: number;
    total?: number;
  }>('/admin/menu/list', { params });
  return {
    limit: Number(response?.limit ?? params.page_size ?? 10),
    list: (response?.list ?? []).map(normalizeMenu),
    page: Number(response?.page ?? params.page ?? 1),
    total: Number(response?.total ?? 0),
  } satisfies AdminMenuApi.ListResponse;
}

async function getMenuDetail(id: number) {
  const response = await requestClient.get<Record<string, any>>(
    '/admin/menu/detail',
    { params: { id } },
  );
  return normalizeMenu(response ?? {});
}

async function getMenuOptions() {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/admin/menu/options');
  return (response?.list ?? []).map((item) => normalizeMenu(item));
}

async function saveMenu(
  id: number,
  data: Omit<
    AdminMenuApi.AdminMenu,
    'children' | 'created_at' | 'id' | 'updated_at'
  >,
) {
  const payload: Record<string, any> = {
    component: data.component ?? '',
    icon: data.icon ?? '',
    menu_type: data.type,
    name: data.name,
    parent_id: Number(data.pid ?? 0),
    path: data.path ?? '',
    permission_code: data.permission ?? '',
    sort: Number(data.sort ?? 0),
    status: Number(data.status ?? 1),
    title: data.title,
    visible: Number(data.visible ?? 1) === 1,
  };
  if (id > 0) {
    payload.id = id;
  }
  return requestClient.post('/admin/menu/save', payload);
}

async function deleteMenu(data: { ids: number[] }) {
  return requestClient.post('/admin/menu/delete', data);
}

export { deleteMenu, getMenuDetail, getMenuList, getMenuOptions, saveMenu };
