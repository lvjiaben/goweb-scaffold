<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import MenuTree from '@/components/MenuTree.vue';
import { adminState, logoutAndClear } from '@/auth';

const route = useRoute();
const router = useRouter();

const currentTitle = computed(() => String(route.meta.title || '工作台'));

async function logout() {
  await logoutAndClear();
  await router.push('/login');
}
</script>

<template>
  <div class="page-shell">
    <aside class="sidebar">
      <div class="brand">
        <img src="/logo.png" alt="logo" class="brand-logo" />
        <div>
          <strong>Goweb Admin</strong>
          <p>vben-admin 底板已适配新 API</p>
        </div>
      </div>
      <nav class="nav card">
        <MenuTree :items="adminState.menus" />
      </nav>
    </aside>

    <section class="main-area">
      <header class="card topbar">
        <div>
          <p class="topbar-kicker">当前页面</p>
          <h1>{{ currentTitle }}</h1>
        </div>
        <div class="topbar-actions">
          <div class="user-meta">
            <strong>{{ adminState.user?.nickname || adminState.user?.username }}</strong>
            <span>{{ adminState.user?.is_super ? '超级管理员' : '管理员' }}</span>
          </div>
          <button class="btn secondary" @click="logout">退出</button>
        </div>
      </header>

      <main class="content">
        <RouterView />
      </main>
    </section>
  </div>
</template>

<style scoped>
.sidebar {
  width: 320px;
  padding: 24px 18px;
  background: linear-gradient(180deg, #14213d 0%, #1f304f 100%);
  color: #f6f1e8;
}

.brand {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 18px;
}

.brand p {
  margin: 4px 0 0;
  color: rgba(246, 241, 232, 0.66);
  font-size: 13px;
}

.brand-logo {
  width: 52px;
  height: 52px;
}

.nav {
  margin-top: 18px;
  padding: 12px;
  color: #f6f1e8;
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.12);
}

.main-area {
  flex: 1;
  padding: 24px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
}

.topbar h1 {
  margin: 6px 0 0;
}

.topbar-kicker {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #b15c2e;
}

.topbar-actions {
  display: flex;
  gap: 18px;
  align-items: center;
}

.user-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.user-meta span {
  color: rgba(20, 33, 61, 0.66);
  font-size: 13px;
}

.content {
  padding-top: 22px;
}

@media (max-width: 900px) {
  .sidebar {
    width: 100%;
  }

  .topbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
