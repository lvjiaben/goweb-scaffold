<script setup lang="ts">
import { noticeState, removeNotice } from '@/notify';
</script>

<template>
  <Teleport to="body">
    <aside class="notice-stack">
      <article
        v-for="item in noticeState.items"
        :key="item.id"
        class="notice-item card"
        :class="`notice-item--${item.type}`"
      >
        <span>{{ item.message }}</span>
        <button class="notice-item__close" type="button" @click="removeNotice(item.id)">×</button>
      </article>
    </aside>
  </Teleport>
</template>

<style scoped>
.notice-stack {
  position: fixed;
  top: 18px;
  right: 18px;
  z-index: 1100;
  display: grid;
  gap: 10px;
  width: min(380px, calc(100vw - 24px));
}

.notice-item {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  padding: 14px 16px;
}

.notice-item--success {
  border-color: rgba(18, 130, 72, 0.22);
}

.notice-item--warning {
  border-color: rgba(200, 122, 37, 0.24);
}

.notice-item--error {
  border-color: rgba(173, 59, 59, 0.22);
}

.notice-item--info {
  border-color: rgba(35, 81, 124, 0.22);
}

.notice-item__close {
  border: none;
  color: rgba(24, 36, 61, 0.66);
  background: transparent;
}
</style>
