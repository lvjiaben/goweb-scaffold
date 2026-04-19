<script setup lang="ts">
import type { MenuItem } from '@/types';

defineProps<{
  items: MenuItem[];
}>();

function resolveMenuPath(item: MenuItem): string {
  if (item.children?.length) {
    return resolveMenuPath(item.children[0]);
  }
  return item.path || '/dashboard';
}
</script>

<template>
  <ul class="menu-tree">
    <li v-for="item in items" :key="item.id">
      <RouterLink :to="resolveMenuPath(item)" class="menu-link">
        <span>{{ item.title }}</span>
      </RouterLink>
      <MenuTree v-if="item.children?.length" :items="item.children" />
    </li>
  </ul>
</template>

<style scoped>
.menu-tree {
  margin: 0;
  padding: 0 0 0 14px;
  list-style: none;
}

.menu-link {
  display: block;
  padding: 8px 0;
  color: rgba(246, 241, 232, 0.82);
}

.menu-link.router-link-active {
  color: #f7d27a;
  font-weight: 700;
}
</style>
