import { reactive } from 'vue';

import type { NoticeType } from '@/types';

type NoticeItem = {
  id: number;
  type: NoticeType;
  message: string;
};

const dismissTimers = new Map<number, number>();
let seed = 1;

export const noticeState = reactive<{
  items: NoticeItem[];
}>({
  items: [],
});

export function pushNotice(type: NoticeType, message: string, duration = 2200) {
  const item: NoticeItem = {
    id: seed++,
    type,
    message,
  };
  noticeState.items.push(item);

  const timer = window.setTimeout(() => {
    removeNotice(item.id);
  }, duration);
  dismissTimers.set(item.id, timer);
  return item.id;
}

export function removeNotice(id: number) {
  const timer = dismissTimers.get(id);
  if (timer) {
    window.clearTimeout(timer);
    dismissTimers.delete(id);
  }
  noticeState.items = noticeState.items.filter((item) => item.id !== id);
}

export function notifySuccess(message: string) {
  return pushNotice('success', message);
}

export function notifyError(message: string) {
  return pushNotice('error', message, 2600);
}

export function notifyWarning(message: string) {
  return pushNotice('warning', message, 2600);
}

export function notifyInfo(message: string) {
  return pushNotice('info', message);
}
