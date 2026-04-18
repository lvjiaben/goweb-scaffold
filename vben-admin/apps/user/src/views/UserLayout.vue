<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';

import { logoutUser, userState } from '@/auth';

const route = useRoute();
const router = useRouter();

async function logout() {
  await logoutUser();
  await router.push('/login');
}
</script>

<template>
  <div class="layout">
    <header class="card topbar">
      <div>
        <strong>Goweb User</strong>
        <p>{{ String(route.meta.title || '首页') }}</p>
      </div>
      <nav class="nav">
        <RouterLink to="/home">首页</RouterLink>
        <RouterLink to="/profile">资料</RouterLink>
        <button class="btn secondary" @click="logout">退出</button>
      </nav>
    </header>

    <main class="content">
      <RouterView />
    </main>

    <footer class="card footer">
      <span>当前用户：{{ userState.profile?.nickname || userState.profile?.username }}</span>
      <span>接口基址：/api</span>
    </footer>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  padding: 24px;
  display: grid;
  grid-template-rows: auto 1fr auto;
  gap: 18px;
}

.topbar,
.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 18px 22px;
}

.topbar p {
  margin: 4px 0 0;
  color: rgba(31, 41, 55, 0.66);
}

.nav {
  display: flex;
  gap: 14px;
  align-items: center;
}

.nav a.router-link-active {
  font-weight: 700;
}

.content {
  display: grid;
}

@media (max-width: 860px) {
  .topbar,
  .footer {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .nav {
    flex-wrap: wrap;
  }
}
</style>
