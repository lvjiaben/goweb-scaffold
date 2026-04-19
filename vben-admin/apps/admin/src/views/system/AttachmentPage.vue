<script setup lang="ts">
import { onMounted, ref } from 'vue';

import AppTable from '@/components/AppTable.vue';
import PermissionButton from '@/components/PermissionButton.vue';
import { deleteAttachment, fetchAttachments, uploadAttachment } from '@/api/attachment';
import { copyText, formatTime, getErrorMessage, isImageFile } from '@/helpers';
import type { AttachmentItem, TableColumn } from '@/types';

const columns: TableColumn[] = [
  { key: 'preview', title: '预览', width: '120px' },
  { key: 'original_name', title: '文件名' },
  { key: 'url', title: 'URL', width: '260px' },
  { key: 'file_size', title: '大小', width: '120px', align: 'right' },
  { key: 'created_at', title: '上传时间', width: '180px' },
  { key: 'actions', title: '操作', width: '220px', align: 'right' },
];

const rows = ref<AttachmentItem[]>([]);
const loading = ref(false);
const uploading = ref(false);
const selectedFile = ref<File | null>(null);
const keyword = ref('');
const errorMessage = ref('');
const total = ref(0);

function formatSize(size: number) {
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

async function load() {
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await fetchAttachments({ keyword: keyword.value || undefined, page: 1, page_size: 50 });
    rows.value = result.list || [];
    total.value = result.total || 0;
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '加载附件失败');
  } finally {
    loading.value = false;
  }
}

async function submitUpload() {
  if (!selectedFile.value) {
    errorMessage.value = '请先选择要上传的文件';
    return;
  }
  uploading.value = true;
  errorMessage.value = '';
  try {
    await uploadAttachment(selectedFile.value);
    selectedFile.value = null;
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '上传附件失败');
  } finally {
    uploading.value = false;
  }
}

async function removeRow(row: AttachmentItem) {
  if (!window.confirm(`确认删除文件「${row.original_name}」吗？`)) {
    return;
  }
  try {
    await deleteAttachment(row.id);
    await load();
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '删除附件失败');
  }
}

async function copyURL(url: string) {
  try {
    await copyText(url);
  } catch (error) {
    errorMessage.value = getErrorMessage(error, '复制 URL 失败');
  }
}

onMounted(load);
</script>

<template>
  <section class="page-stack">
    <article class="card page-card">
      <div class="section-heading">
        <div>
          <h3>附件管理</h3>
          <p>支持上传、图片预览、复制 URL 和删除附件，非图片文件展示文件名与访问地址。</p>
        </div>
      </div>

      <div class="upload-panel card inset-card">
        <div class="upload-panel__meta">
          <strong>上传新文件</strong>
          <span class="text-muted">上传目录：storage/uploads</span>
        </div>
        <div class="upload-panel__actions">
          <input
            type="file"
            class="input file-input"
            @change="selectedFile = (($event.target as HTMLInputElement).files || [])[0] || null"
          />
          <PermissionButton code="attachment.upload">
            <button class="btn" type="button" :disabled="uploading" @click="submitUpload">
              {{ uploading ? '上传中...' : '上传文件' }}
            </button>
          </PermissionButton>
        </div>
      </div>

      <div class="toolbar-row">
        <div class="search-group">
          <input v-model="keyword" class="input" placeholder="搜索文件名" @keyup.enter="load" />
          <button class="btn secondary" type="button" @click="load">搜索</button>
        </div>
        <span class="text-muted">共 {{ total }} 个附件</span>
      </div>

      <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

      <AppTable :columns="columns" :rows="rows" :loading="loading" empty-text="暂无附件">
        <template #cell-preview="{ row }">
          <div class="attachment-preview">
            <img v-if="isImageFile(row)" :src="row.url" :alt="row.original_name" />
            <div v-else class="attachment-fallback">
              <strong>{{ row.file_ext || 'file' }}</strong>
            </div>
          </div>
        </template>
        <template #cell-original_name="{ row }">
          <div class="stack-xs">
            <strong>{{ row.original_name }}</strong>
            <small class="text-muted">{{ row.mime_type || '-' }}</small>
          </div>
        </template>
        <template #cell-url="{ value }">
          <a :href="value" class="link-text" target="_blank" rel="noreferrer">{{ value }}</a>
        </template>
        <template #cell-file_size="{ value }">
          {{ formatSize(Number(value || 0)) }}
        </template>
        <template #cell-created_at="{ value }">
          {{ formatTime(value) }}
        </template>
        <template #cell-actions="{ row }">
          <div class="table-actions">
            <button class="btn secondary btn-sm" type="button" @click="copyURL(row.url)">复制 URL</button>
            <a class="btn secondary btn-sm" :href="row.url" target="_blank" rel="noreferrer">打开</a>
            <PermissionButton code="attachment.delete">
              <button class="btn danger btn-sm" type="button" @click="removeRow(row)">删除</button>
            </PermissionButton>
          </div>
        </template>
      </AppTable>
    </article>
  </section>
</template>
