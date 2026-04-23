import { requestClient } from '#/api/request';

export namespace ConfigApi {
  /** 配置项 */
  export interface Config {
    id: number;
    dir: string;
    key: string;
    name: string;
    tip: string;
    type: string;
    value: string;
    variable: string;
    created_at: number;
  }

  /** 配置分组 */
  export interface ConfigGroup {
    id: number;
    dir: string;
    children: Config[];
  }

  /** 列表响应 */
  export interface ListResponse {
    list: ConfigGroup[];
  }
}

function toUnixTimestamp(value?: string | number | null) {
  if (!value) {
    return 0;
  }
  if (typeof value === 'number') {
    return value > 1_000_000_000_000 ? Math.floor(value / 1000) : value;
  }
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? 0 : Math.floor(timestamp / 1000);
}

function stringifyValue(value: any) {
  if (typeof value === 'string') {
    return value;
  }
  if (value === null || value === undefined) {
    return '';
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

/**
 * 获取配置列表
 */
export async function getConfigList() {
  const response = await requestClient.get<{
    list?: Array<Record<string, any>>;
  }>('/system_config/list');
  const groups = new Map<string, ConfigApi.ConfigGroup>();
  for (const item of response?.list ?? []) {
    const dir = item.dir ?? 'system';
    if (!groups.has(dir)) {
      groups.set(dir, {
        children: [],
        dir,
        id: groups.size + 1,
      });
    }
    groups.get(dir)?.children.push({
      created_at: toUnixTimestamp(item.created_at),
      dir,
      id: Number(item.id ?? 0),
      key: item.config_key ?? '',
      name: item.config_name ?? '',
      tip: item.remark ?? '',
      type: item.type ?? 'input',
      value: stringifyValue(item.config_value),
      variable: item.variable ?? '',
    });
  }
  return {
    list: [...groups.values()],
  };
}

/**
 * 创建配置
 */
export async function createConfig(data: Partial<ConfigApi.Config>) {
  return requestClient.post('/system_config/save', {
    config_key: data.key,
    config_name: data.name,
    config_value: data.value ?? '',
    id: data.id,
    remark: data.tip ?? '',
  });
}

/**
 * 批量更新配置
 */
export async function updateConfigs(data: Record<string, any>) {
  const configs = Object.entries(data);
  await Promise.all(
    configs.map(([key, value]) =>
      requestClient.post('/system_config/save', {
        config_key: key,
        config_name: key,
        config_value: stringifyValue(value),
        remark: '',
      }),
    ),
  );
  return true;
}

/**
 * 删除配置
 */
export async function deleteConfig(id: number) {
  return requestClient.post('/system_config/delete', { ids: [id] });
}
