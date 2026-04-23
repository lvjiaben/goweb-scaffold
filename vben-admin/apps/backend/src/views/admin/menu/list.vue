<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, VbenButton, VbenButtonGroup, useVbenDrawer } from '@vben/common-ui';
import { Check, Eraser, Plus, X } from '@vben/icons';
import { $t } from '@vben/locales';

import { InputSearch, message, Popconfirm, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteMenu,
  getMenuDetail,
  getMenuList,
} from '#/api/admin/menu';

import type { AdminMenuApi } from '#/api/admin/menu';

import {
  getMenuTypeOptions,
  getStatusOptions,
  getVisibleOptions,
  useColumns,
  useSearchFormSchema,
} from './data';
import FormDrawerComponent from './modules/form-drawer.vue';

const selectedCount = ref(0);
const searchValue = ref('');

const typeOptions = getMenuTypeOptions();
const statusOptions = getStatusOptions();
const visibleOptions = getVisibleOptions();

const getSelectedRecords = (): AdminMenuApi.AdminMenu[] => {
  return (gridApi.grid?.getCheckboxRecords() as AdminMenuApi.AdminMenu[]) || [];
};

const updateSelectedCount = () => {
  requestAnimationFrame(() => {
    selectedCount.value = getSelectedRecords().length;
  });
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
          return await getMenuList({
            filter: JSON.stringify(formValues ?? {}),
            page: page.currentPage,
            page_size: page.pageSize,
            search: searchValue.value || undefined,
            sort_by: sort.field ? String(sort.field) : 'sort',
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

const getTypeOption = (value: string) => {
  return typeOptions.find((item) => item.value === value);
};

const getStatusOption = (value: number) => {
  return statusOptions.find((item) => item.value === value);
};

const getVisibleOption = (value: number) => {
  return visibleOptions.find((item) => item.value === value);
};

const onRefresh = () => {
  gridApi.grid?.clearCheckboxRow();
  gridApi.query();
  selectedCount.value = 0;
};

const onEdit = async (row: AdminMenuApi.AdminMenu) => {
  const detail = await getMenuDetail(row.id);
  formDrawerApi.setData(detail).open();
};

const onCreate = () => {
  formDrawerApi
    .setData({
      component: '',
      icon: '',
      name: '',
      path: '',
      permission: '',
      pid: 0,
      sort: 0,
      status: 1,
      title: '',
      type: 'menu',
      visible: 1,
    })
    .open();
};

const onCreateChild = (row: AdminMenuApi.AdminMenu) => {
  formDrawerApi
    .setData({
      component: '',
      icon: '',
      name: '',
      path: '',
      permission: '',
      pid: row.id,
      sort: 0,
      status: 1,
      title: '',
      type: 'menu',
      visible: 1,
    })
    .open();
};

const onDelete = (row?: AdminMenuApi.AdminMenu) => {
  let ids: number[] = [];

  if (row) {
    ids = [row.id];
  } else {
    const selectRecords = getSelectedRecords();
    if (selectRecords.length === 0) {
      message.warning($t('common.tableSelectTip'));
      return;
    }
    ids = selectRecords.map((item) => item.id);
  }

  const hideLoading = message.loading({
    content: $t('ui.actionMessage.deleting'),
    duration: 0,
    key: 'action_process_msg',
  });
  deleteMenu({ ids })
    .then(() => {
      message.success({
        content: $t('ui.actionMessage.deleteSuccess'),
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
      <template #title="{ row }">
        <div class="flex items-center gap-2">
          <span>{{ row.title }}</span>
          <span class="text-text-secondary text-xs">{{ row.name }}</span>
        </div>
      </template>
      <template #type="{ row }">
        <Tag :color="getTypeOption(row.type)?.color">
          {{ getTypeOption(row.type)?.label ?? row.type }}
        </Tag>
      </template>
      <template #component="{ row }">
        <span>{{ row.type === 'button' ? row.permission : row.component || '-' }}</span>
      </template>
      <template #visible="{ row }">
        <Tag :color="getVisibleOption(row.visible)?.color">
          {{ getVisibleOption(row.visible)?.label ?? row.visible }}
        </Tag>
      </template>
      <template #status="{ row }">
        <Tag :color="getStatusOption(row.status)?.color">
          {{ getStatusOption(row.status)?.label ?? row.status }}
        </Tag>
      </template>
      <template #operation="{ row }">
        <VbenButtonGroup>
          <VbenButton size="small" @click="onCreateChild(row)">
            新增下级
          </VbenButton>
          <VbenButton size="small" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm :title="`确认删除菜单 ${row.title} 吗？`" @confirm="onDelete(row)">
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
            {{ $t('ui.actionTitle.create', [$t('admin.menu.name')]) }}
          </VbenButton>
          <Popconfirm title="确认删除选中的菜单吗？" @confirm="onDelete()">
            <VbenButton :disabled="selectedCount === 0" status="danger">
              <Eraser class="size-4" />
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
            :placeholder="$t('common.fuzzySearch')"
            @clear="onClearSearch"
            @search="onSearch"
          />
        </div>
      </template>
    </Grid>
  </Page>
</template>
