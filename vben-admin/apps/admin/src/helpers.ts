import type { FlatMenuItem, MenuItem } from './types';

export function formatTime(value?: string | Date | null) {
  if (!value) {
    return '-';
  }
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return date.toLocaleString('zh-CN', { hour12: false });
}

export function prettyJSON(value: unknown) {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

export async function copyText(text: string) {
  await navigator.clipboard.writeText(text);
}

export function getErrorMessage(error: unknown, fallback = '操作失败') {
  return error instanceof Error ? error.message : fallback;
}

export function isImageFile(file: { mime_type?: string; file_ext?: string }) {
  const mime = (file.mime_type || '').toLowerCase();
  const ext = (file.file_ext || '').toLowerCase();
  return mime.startsWith('image/') || ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg'].includes(ext);
}

export function flattenMenuTree(items: MenuItem[], depth = 0): FlatMenuItem[] {
  return items.flatMap((item) => [
    { ...item, depth },
    ...flattenMenuTree(item.children || [], depth + 1),
  ]);
}
