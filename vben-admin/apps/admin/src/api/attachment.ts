import { request } from './request';
import type { AttachmentItem, Paginated } from '@/types';

export function fetchAttachments(params?: { keyword?: string; page?: number; page_size?: number }) {
  return request.get<Paginated<AttachmentItem>>('/attachment/list', { params });
}

export function uploadAttachment(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  return request.post('/attachment/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

export function deleteAttachment(id: number) {
  return request.post('/attachment/delete', { id });
}
