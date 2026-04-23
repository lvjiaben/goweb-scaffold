<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { DemoArticleApi } from '#/api/demo_article';

import { ref } from 'vue';

import { Page, VbenButton, VbenButtonGroup, useVbenDrawer } from '@vben/common-ui';
import { Check, Plus, Trash, X } from '@vben/icons';
import { InputSearch, message, Popconfirm, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteDemoArticle,
  getDemoArticleDetail,
  getDemoArticleList,
} from '#/api/demo_article';

import {
  getDemoArticleStatusOptions,
  useColumns,
  useSearchFormSchema,
} from './data';
import FormDrawerComponent from './modules/form-drawer.vue';

const selectedCount = ref(0);
const searchValue = ref('');
const statusOptions = getDemoArticleStatusOptions();

const getSelectedRecords = (): DemoArticleApi.DemoArticle[] => {
  return (gridApi.grid?.getCheckboxRecords() as DemoArticleApi.DemoArticle[]) || [];
};

const updateSelectedCount = () => {
  requestAnimationFrame(() => {
    selectedCount.value = getSelectedRecords().length;
  });
};

const resolveOption = (options: Array<{ color?: string; label: string; value: any }>, value: any) => {
  return options.find((item) => String(item.value) === String(value));
};

const onClearSearch = () => {
  searchValue.value = '';
  onRefresh();
};

const onSearch = (value: string) => {
  searchValue.value = value;
  onRefresh();
};

const [FormDrawer, formDrawerApi] = useVbenDrawer({
  connectedComponent: FormDrawerComponent,
  destroyOnClose: true,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    collapsed: true,
    schema: useSearchFormSchema(),
    showCollapseButton: true,
    submitOnChange: false,
    submitOnEnter: false,
  },
  gridOptions: {
    columns: useColumns(),
    keepSource: true,
    pagerConfig: {
      enabled: true,
    },
    proxyConfig: {
      ajax: {
        query: async ({ page, sort }, formValues) => {
          return await getDemoArticleList({
            filter: JSON.stringify(formValues ?? {}),
            page: page.currentPage,
            page_size: page.pageSize,
            search: searchValue.value || undefined,
            sort_by: sort.field ? String(sort.field) : 'id',
            sort_order: sort.order === 'asc' ? 'asc' : 'desc',
          });
        },
      },
      sort: true,
    },
    rowConfig: {
      isHover: true,
      keyField: 'id',
    },
    sortConfig: {
      remote: true,
      trigger: 'cell',
    },
    stripe: false,
    toolbarConfig: {
      custom: true,
      refresh: true,
      zoom: true,
    },
    checkboxConfig: {
      highlight: true,
      reserve: false,
      trigger: 'cell',
    },
  } as VxeTableGridOptions,
  gridEvents: {
    checkboxAll: updateSelectedCount,
    checkboxChange: updateSelectedCount,
  },
});

const onRefresh = () => {
  gridApi.grid?.clearCheckboxRow();
  gridApi.query();
  selectedCount.value = 0;
};

const onEdit = async (row: DemoArticleApi.DemoArticle) => {
  const detail = await getDemoArticleDetail(row.id);
  formDrawerApi.setData(detail).open();
};

const onCreate = () => {
  formDrawerApi.setData({}).open();
};

const onDelete = (row?: DemoArticleApi.DemoArticle) => {
  let ids: number[] = [];

  if (row) {
    ids = [row.id];
  } else {
    const selectRecords = getSelectedRecords();
    if (selectRecords.length === 0) {
      message.warning('请选择要删除的数据');
      return;
    }
    ids = selectRecords.map((item) => item.id);
  }

  const hideLoading = message.loading({
    content: '删除中...',
    duration: 0,
    key: 'action_process_msg',
  });
  deleteDemoArticle({ ids })
    .then(() => {
      message.success({
        content: '删除成功',
        key: 'action_process_msg',
      });
      onRefresh();
    })
    .catch(() => {
      hideLoading();
    });
};
</script>

<template>
  <Page>
    <FormDrawer @success="onRefresh" />
    <Grid>
      <template #status="{ row }">
        <Tag :color="resolveOption(statusOptions, row.status)?.color">
          {{ resolveOption(statusOptions, row.status)?.label ?? row.status }}
        </Tag>
      </template>
      <template #operation="{ row }">
        <VbenButtonGroup>
          <VbenButton size="small" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm title="确认删除这条数据吗？" @confirm="onDelete(row)">
            <VbenButton size="small" status="danger">
              删除
            </VbenButton>
          </Popconfirm>
        </VbenButtonGroup>
      </template>
      <template #toolbar-tools>
        <div class="flex items-center gap-3">
          <VbenButton type="primary" @click="onCreate">
            <Plus class="size-5" />
            新增演示文章
          </VbenButton>
          <Popconfirm title="确认删除选中的数据吗？" @confirm="onDelete()">
            <VbenButton :disabled="selectedCount === 0" status="danger">
              <Trash class="size-4" />
              删除选中
            </VbenButton>
          </Popconfirm>
          <VbenButtonGroup>
            <VbenButton :disabled="selectedCount === 0" @click="gridApi.grid?.setAllCheckboxRow(true)">
              <Check class="size-4" />
              全选
            </VbenButton>
            <VbenButton :disabled="selectedCount === 0" @click="gridApi.grid?.clearCheckboxRow()">
              <X class="size-4" />
              清空
            </VbenButton>
          </VbenButtonGroup>
          <InputSearch
            allow-clear
            placeholder="请输入关键词"
            @clear="onClearSearch"
            @search="onSearch"
          />
        </div>
      </template>
    </Grid>
  </Page>
</template>
