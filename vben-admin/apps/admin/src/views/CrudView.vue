<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import { deleteModule, detailModule, listModule, saveModule } from '@/api';

const route = useRoute();
const items = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const errorMessage = ref('');
const editingId = ref<number | null>(null);
const payloadText = ref('{\n  \n}');

const title = computed(() => String(route.meta.title || '模块'));
const moduleName = computed(() => String(route.meta.module || ''));

async function load() {
  if (!moduleName.value) {
    return;
  }
  loading.value = true;
  errorMessage.value = '';
  try {
    const response = await listModule(moduleName.value);
    items.value = Array.isArray(response.list) ? response.list : [];
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载失败';
  } finally {
    loading.value = false;
  }
}

function createNew() {
  editingId.value = null;
  payloadText.value = '{\n  \n}';
}

async function editRow(row: Record<string, unknown>) {
  const id = Number(row.id || 0);
  if (!id) {
    payloadText.value = JSON.stringify(row, null, 2);
    editingId.value = null;
    return;
  }
  try {
    const detail = await detailModule(moduleName.value, id);
    editingId.value = id;
    payloadText.value = JSON.stringify(detail, null, 2);
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载详情失败';
  }
}

async function saveCurrent() {
  errorMessage.value = '';
  try {
    const payload = JSON.parse(payloadText.value) as Record<string, unknown>;
    await saveModule(moduleName.value, payload);
    await load();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '保存失败';
  }
}

async function removeRow(row: Record<string, unknown>) {
  const id = Number(row.id || 0);
  if (!id) {
    return;
  }
  if (!window.confirm(`确认删除 #${id} 吗？`)) {
    return;
  }
  try {
    await deleteModule(moduleName.value, { id });
    await load();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '删除失败';
  }
}

onMounted(load);
watch(moduleName, createNew);
watch(moduleName, load);
</script>

<template>
  <section class="crud-page">
    <article class="card crud-list">
      <div class="toolbar">
        <div>
          <h2 class="section-title">{{ title }}</h2>
          <p>这里保留最小 JSON 维护入口，直接对接 `{{ moduleName }}` 模块的 `list/save/delete`。</p>
        </div>
        <button class="btn secondary" @click="createNew">新建 / 清空</button>
      </div>

      <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>
      <p v-if="loading">加载中...</p>

      <div v-else class="crud-grid">
        <article v-for="row in items" :key="String(row.id)" class="card row-card">
          <div class="row-meta">
            <strong>#{{ row.id }}</strong>
            <div class="row-actions">
              <button class="btn secondary" @click="editRow(row)">编辑</button>
              <button class="btn danger" @click="removeRow(row)">删除</button>
            </div>
          </div>
          <pre class="json-box">{{ JSON.stringify(row, null, 2) }}</pre>
        </article>
      </div>
    </article>

    <article class="card editor">
      <h2 class="section-title">保存面板</h2>
      <p>编辑 ID：{{ editingId ?? '新记录' }}</p>
      <div class="field">
        <label>JSON Payload</label>
        <textarea v-model="payloadText" rows="24" />
      </div>
      <button class="btn" @click="saveCurrent">提交到 /save</button>
    </article>
  </section>
</template>

<style scoped>
.crud-page {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 20px;
}

.crud-list,
.editor {
  padding: 22px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
}

.toolbar p {
  margin: 0;
  max-width: 60ch;
  color: rgba(20, 33, 61, 0.66);
}

.crud-grid {
  display: grid;
  gap: 16px;
}

.row-card {
  padding: 16px;
}

.row-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.row-actions {
  display: flex;
  gap: 10px;
}

.error-text {
  color: #8e2d2d;
}

@media (max-width: 1100px) {
  .crud-page {
    grid-template-columns: 1fr;
  }
}
</style>
