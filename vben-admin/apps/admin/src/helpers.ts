import type { FlatMenuItem, MenuItem, MenuOption } from './types';

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
  const message = error instanceof Error ? error.message : fallback;
  const normalized = message.toLowerCase();
  if (normalized.includes('permission denied') || normalized.includes('forbidden')) {
    return '当前账号无权执行该操作';
  }
  if (normalized.includes('unauthorized') || normalized.includes('token')) {
    return '登录状态已失效，请重新登录';
  }
  return message || fallback;
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

export function findMenuTrail(items: MenuItem[], path: string, trail: MenuItem[] = []): MenuItem[] {
  for (const item of items) {
    const nextTrail = [...trail, item];
    if (item.path === path) {
      return nextTrail;
    }
    if (item.children?.length) {
      const matched = findMenuTrail(item.children, path, nextTrail);
      if (matched.length) {
        return matched;
      }
    }
  }
  return [];
}

export function findFirstMenuPath(items: MenuItem[]): string {
  for (const item of items) {
    if (item.path) {
      return item.path;
    }
    if (item.children?.length) {
      const childPath = findFirstMenuPath(item.children);
      if (childPath) {
        return childPath;
      }
    }
  }
  return '/forbidden';
}

export function collectMenuPaths(items: MenuItem[]): string[] {
  return flattenMenuTree(items)
    .map((item) => item.path)
    .filter((item): item is string => Boolean(item));
}

export function menuTreeToOptions(items: MenuItem[]): MenuOption[] {
  return items.map((item) => ({
    label: item.title,
    value: item.id,
    menu_type: item.menu_type,
    children: menuTreeToOptions(item.children || []),
  }));
}
