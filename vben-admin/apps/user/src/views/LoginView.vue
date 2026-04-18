<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { loginUser, registerUser } from '@/auth';

const router = useRouter();
const mode = ref<'login' | 'register'>('login');
const username = ref('');
const password = ref('');
const nickname = ref('');
const message = ref('');
const submitting = ref(false);

async function submit() {
  message.value = '';
  submitting.value = true;
  try {
    if (mode.value === 'login') {
      await loginUser(username.value, password.value);
    } else {
      await registerUser(username.value, password.value, nickname.value || username.value);
    }
    await router.push('/home');
  } catch (error) {
    message.value = error instanceof Error ? error.message : '操作失败';
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <main class="login-page">
    <section class="card login-card">
      <p class="kicker">goweb user</p>
      <h1>{{ mode === 'login' ? '用户登录' : '用户注册' }}</h1>
      <div class="switcher">
        <button class="btn secondary" @click="mode = 'login'">登录</button>
        <button class="btn secondary" @click="mode = 'register'">注册</button>
      </div>

      <div class="field">
        <label>用户名</label>
        <input v-model="username" placeholder="demo-user" />
      </div>
      <div v-if="mode === 'register'" class="field">
        <label>昵称</label>
        <input v-model="nickname" placeholder="演示用户" />
      </div>
      <div class="field">
        <label>密码</label>
        <input v-model="password" type="password" placeholder="请输入密码" />
      </div>
      <p v-if="message" class="message">{{ message }}</p>
      <button class="btn" :disabled="submitting" @click="submit">
        {{ submitting ? '提交中...' : mode === 'login' ? '登录' : '注册并登录' }}
      </button>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
}

.login-card {
  width: min(460px, 100%);
  padding: 28px;
}

.kicker {
  margin: 0;
  color: #1d4ed8;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.switcher {
  display: flex;
  gap: 10px;
  margin: 18px 0;
}

.message {
  color: #b91c1c;
}
</style>
