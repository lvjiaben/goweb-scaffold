<script setup lang="ts">
import { onMounted, ref } from 'vue';

import { deleteModule, listModule, uploadAttachment } from '@/api';

type AttachmentItem = {
  id: number;
  original_name: string;
  url: string;
  file_size: number;
  created_at: string;
};

const attachments = ref<AttachmentItem[]>([]);
const selectedFile = ref<File | null>(null);
const message = ref('');

async function load() {
  const result = await listModule('attachment');
  attachments.value = Array.isArray(result.list) ? result.list : [];
}

async function upload() {
  if (!selectedFile.value) {
    message.value = '请选择文件';
    return;
  }
  try {
    await uploadAttachment(selectedFile.value);
    message.value = '上传成功';
    selectedFile.value = null;
    await load();
  } catch (error) {
    message.value = error instanceof Error ? error.message : '上传失败';
  }
}

async function remove(id: number) {
  await deleteModule('attachment', { id });
  await load();
}

onMounted(load);
</script>

<template>
  <section class="card attachment-page">
    <h2 class="section-title">附件管理</h2>
    <div class="upload-box">
      <input
        type="file"
        @change="selectedFile = (($event.target as HTMLInputElement).files || [])[0] || null"
      />
      <button class="btn" @click="upload">上传到 /admin-api/attachment/upload</button>
    </div>
    <p v-if="message">{{ message }}</p>

    <div class="attachment-list">
      <article v-for="item in attachments" :key="item.id" class="card attachment-item">
        <div>
          <strong>{{ item.original_name }}</strong>
          <p>{{ item.file_size }} bytes</p>
        </div>
        <div class="attachment-actions">
          <a :href="item.url" target="_blank" class="btn secondary">打开</a>
          <button class="btn danger" @click="remove(item.id)">删除</button>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.attachment-page {
  padding: 22px;
}

.upload-box {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 18px;
}

.attachment-list {
  display: grid;
  gap: 14px;
}

.attachment-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
}

.attachment-item p {
  margin: 6px 0 0;
  color: rgba(20, 33, 61, 0.66);
}

.attachment-actions {
  display: flex;
  gap: 10px;
}

@media (max-width: 900px) {
  .upload-box,
  .attachment-item {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
