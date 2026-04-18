<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';

import { changePassword, fetchProfile, saveProfile } from '@/api';
import { userState } from '@/auth';

const form = reactive({
  nickname: '',
  email: '',
  mobile: '',
});
const passwordForm = reactive({
  old_password: '',
  new_password: '',
});
const message = ref('');

async function load() {
  const profile = await fetchProfile();
  userState.profile = profile;
  form.nickname = profile.nickname || '';
  form.email = profile.email || '';
  form.mobile = profile.mobile || '';
}

async function submitProfile() {
  try {
    await saveProfile(form);
    await load();
    message.value = '资料已保存';
  } catch (error) {
    message.value = error instanceof Error ? error.message : '保存失败';
  }
}

async function submitPassword() {
  try {
    await changePassword(passwordForm);
    passwordForm.old_password = '';
    passwordForm.new_password = '';
    message.value = '密码已修改';
  } catch (error) {
    message.value = error instanceof Error ? error.message : '修改失败';
  }
}

onMounted(load);
</script>

<template>
  <section class="profile-grid">
    <article class="card panel">
      <h2>个人资料</h2>
      <div class="field">
        <label>昵称</label>
        <input v-model="form.nickname" />
      </div>
      <div class="field">
        <label>邮箱</label>
        <input v-model="form.email" />
      </div>
      <div class="field">
        <label>手机号</label>
        <input v-model="form.mobile" />
      </div>
      <button class="btn" @click="submitProfile">保存资料</button>
    </article>

    <article class="card panel">
      <h2>修改密码</h2>
      <div class="field">
        <label>旧密码</label>
        <input v-model="passwordForm.old_password" type="password" />
      </div>
      <div class="field">
        <label>新密码</label>
        <input v-model="passwordForm.new_password" type="password" />
      </div>
      <button class="btn" @click="submitPassword">提交</button>
      <p v-if="message">{{ message }}</p>
    </article>
  </section>
</template>

<style scoped>
.profile-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.panel {
  padding: 22px;
}

@media (max-width: 900px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }
}
</style>
