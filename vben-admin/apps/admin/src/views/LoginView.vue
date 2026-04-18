<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { loginAndLoad } from '@/auth';

const router = useRouter();
const username = ref('admin');
const password = ref('Admin@123456');
const errorMessage = ref('');
const submitting = ref(false);

async function submit() {
  errorMessage.value = '';
  submitting.value = true;
  try {
    const ok = await loginAndLoad(username.value, password.value);
    if (!ok) {
      throw new Error('登录失败');
    }
    await router.push('/dashboard');
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '登录失败';
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <main class="login-page">
    <section class="login-copy">
      <p class="eyebrow">goweb admin</p>
      <h1>后台登录已切到新内核</h1>
      <p>
        这里直接对接 `/admin-api`，菜单和权限都来自新的 `permission_code`
        表驱动模型，不再依赖旧后端架构。
      </p>
    </section>

    <section class="card login-card">
      <img src="/logo.png" alt="logo" class="logo" />
      <h2>登录后台</h2>
      <div class="field">
        <label>用户名</label>
        <input v-model="username" placeholder="admin" />
      </div>
      <div class="field">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="Admin@123456" />
      </div>
      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <button class="btn" :disabled="submitting" @click="submit">
        {{ submitting ? '登录中...' : '登录' }}
      </button>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 32px;
  min-height: 100vh;
  padding: 48px;
  align-items: center;
}

.login-copy {
  padding: 42px;
  color: #14213d;
}

.login-copy h1 {
  margin: 12px 0;
  font-size: clamp(40px, 6vw, 72px);
  line-height: 0.95;
}

.eyebrow {
  margin: 0;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: #b15c2e;
}

.login-card {
  max-width: 420px;
  justify-self: end;
  width: 100%;
  padding: 32px;
}

.logo {
  width: 64px;
  height: 64px;
  margin-bottom: 14px;
}

.error-text {
  color: #8e2d2d;
}

@media (max-width: 900px) {
  .login-page {
    grid-template-columns: 1fr;
    padding: 24px;
  }

  .login-card {
    justify-self: stretch;
  }
}
</style>
