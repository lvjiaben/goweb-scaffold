<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { adminState } from '@/auth';
import { findFirstMenuPath } from '@/helpers';

const route = useRoute();
const router = useRouter();

const sourcePath = computed(() => String(route.query.from || ''));

async function backToAllowedPage() {
  await router.replace(findFirstMenuPath(adminState.menus));
}
</script>

<template>
  <section class="page-stack">
    <article class="card page-card forbidden-page">
      <p class="eyebrow">403</p>
      <h2>当前路由不可访问</h2>
      <p>这个页面不在当前账号可访问的后台菜单中，或当前账号没有对应操作权限。</p>
      <p v-if="sourcePath" class="text-muted">尝试访问：{{ sourcePath }}</p>
      <div class="table-actions">
        <button class="btn" type="button" @click="backToAllowedPage">返回可访问页面</button>
      </div>
    </article>
  </section>
</template>
