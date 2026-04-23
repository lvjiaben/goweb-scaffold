import { requestClient } from '#/api/request';

export namespace AttachmentApi {
  /** 附件 */
  export interface Attachment {
    [key: string]: any;
    /** 附件ID */
    id: number;
    /** 文件名 */
    filename: string;
    /** 文件URL */
    url: string;
    /** 文件路径 */
    path: string;
    /** 文件大小（字节） */
    size: number;
    /** MIME类型 */
    mediatype: string;
    /** 所属目录 */
    parent: string;
    /** 文件扩展名 */
    extension: string;
    /** 创建时间 */
    created_at: number;
    /** 更新时间 */
    updated_at: number;
    /** 类型 */
    type?: string;
    /** 管理员ID */
    admin_id?: number;
    /** 用户ID */
    user_id?: number;
  }

  /** 目录信息 */
  export interface Directory {
    name: string;
    path: string;
    count: number;
  }

  /** 分页参数 */
  export interface ListParams {
    page?: number;
    page_size?: number;
    parent?: string;
    search?: string;
  }

  /** 分页响应 */
  export interface ListResponse {
    list: Attachment[];
    total: number;
    page: number;
    pageSize: number;
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

function normalizeAttachment(row: Record<string, any>): AttachmentApi.Attachment {
  return {
    admin_id: row.admin_id,
    created_at: toUnixTimestamp(row.created_at),
    extension: row.file_ext ?? row.extension ?? '',
    filename: row.original_name ?? row.filename ?? '',
    id: Number(row.id ?? 0),
    mediatype: row.mime_type ?? row.mediatype ?? '',
    parent: row.parent ?? String(row.file_path ?? '').split('/').slice(0, -1).join('/'),
    path: row.file_path ?? row.path ?? '',
    size: Number(row.file_size ?? row.size ?? 0),
    type: row.type,
    updated_at: toUnixTimestamp(row.updated_at),
    url: row.file_url ?? row.url ?? '',
    user_id: row.user_id,
  };
}

/**
 * 获取目录列表
 */
async function getDirectories() {
  const response = await requestClient.get<{
    list?: AttachmentApi.Directory[];
  }>(
    '/attachment/directories',
  );
  return response?.list ?? [];
}

/**
 * 获取附件列表
 */
async function getAttachmentList(params?: AttachmentApi.ListParams) {
  const response = await requestClient.get<{
    limit?: number;
    list?: Array<Record<string, any>>;
    page?: number;
    page_size?: number;
    total?: number;
  }>(
    '/attachment/list',
    { params },
  );
  return {
    list: (response?.list ?? []).map(normalizeAttachment),
    page: Number(response?.page ?? 1),
    pageSize: Number(response?.limit ?? response?.page_size ?? params?.page_size ?? 10),
    total: Number(response?.total ?? 0),
  };
}

/**
 * 上传附件
 */
async function uploadAttachment(file: File, parent?: string) {
  const formData = new FormData();
  formData.append('file', file);
  if (parent) {
    formData.append('parent', parent);
  }
  return requestClient.post<AttachmentApi.Attachment>(
    '/attachment/upload',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    },
  );
}

/**
 * 删除附件
 */
async function deleteAttachment(ids: number[]) {
  return requestClient.post('/attachment/delete', { ids });
}

export { deleteAttachment, getAttachmentList, getDirectories, uploadAttachment };
