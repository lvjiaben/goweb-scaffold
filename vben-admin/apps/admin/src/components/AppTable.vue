<script setup lang="ts">
import type { TableColumn } from '@/types';

withDefaults(
  defineProps<{
    columns: TableColumn[];
    rows: Record<string, any>[];
    rowKey?: string;
    loading?: boolean;
    emptyText?: string;
  }>(),
  {
    rowKey: 'id',
    loading: false,
    emptyText: '暂无数据',
  },
);
</script>

<template>
  <div class="app-table">
    <div v-if="loading" class="app-table__loading">加载中...</div>
    <table v-else-if="rows.length" class="app-table__inner">
      <thead>
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            :style="{ width: column.width || 'auto', textAlign: column.align || 'left' }"
          >
            {{ column.title }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="String(row[rowKey])">
          <td
            v-for="column in columns"
            :key="column.key"
            :style="{ textAlign: column.align || 'left' }"
          >
            <slot
              :name="`cell-${column.key}`"
              :row="row"
              :value="row[column.key]"
              :column="column"
            >
              {{ row[column.key] ?? '-' }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="app-table__empty">{{ emptyText }}</div>
  </div>
</template>
