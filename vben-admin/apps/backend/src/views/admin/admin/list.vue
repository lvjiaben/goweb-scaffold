<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, VbenButton, VbenButtonGroup, useVbenDrawer } from '@vben/common-ui';
import { Check, Eraser, Plus, X } from '@vben/icons';
import { $t } from '@vben/locales';

import { InputSearch, message, Popconfirm, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteAdmin,
  getAdminDetail,
  getAdminList,
} from '#/api/admin/admin';

import type { AdminAdminApi } from '#/api/admin/admin';

import {
  getAdminStatusOptions,
  useColumns,
  useSearchFormSchema,
} from './data';
import FormDrawerComponent from './modules/form-drawer.vue';

const selectedCount = ref(0);
const searchValue = ref('');

const statusOptions = getAdminStatusOptions();

const getSelectedRecords = (): AdminAdminApi.Admin[] => {
  return (gridApi.grid?.getCheckboxRecords() as AdminAdminApi.Admin[]) || [];
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
          return await getAdminList({
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

const getStatusOption = (value: number) => {
  return statusOptions.find((item) => item.value === value);
};

const onRefresh = () => {
  gridApi.grid?.clearCheckboxRow();
  gridApi.query();
  selectedCount.value = 0;
};

const onEdit = async (row: AdminAdminApi.Admin) => {
  const detail = await getAdminDetail(row.id);
  formDrawerApi.setData(detail).open();
};

const onCreate = () => {
  formDrawerApi
    .setData({
      password: '',
      realname: '',
      role_ids: [],
      status: 1,
      username: '',
    })
    .open();
};

const onDelete = (row?: AdminAdminApi.Admin) => {
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
  deleteAdmin({ ids })
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
      <template #status="{ row }">
        <Tag :color="getStatusOption(row.status)?.color">
          {{ getStatusOption(row.status)?.label ?? row.status }}
        </Tag>
      </template>
      <template #is_super="{ row }">
        <Tag :color="row.is_super ? 'gold' : 'default'">
          {{ row.is_super ? '是' : '否' }}
        </Tag>
      </template>
      <template #role_names="{ row }">
        <span>{{ Array.isArray(row.role_names) ? row.role_names.join('、') : '-' }}</span>
      </template>
      <template #operation="{ row }">
        <VbenButtonGroup>
          <VbenButton size="small" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm
            :title="`确认删除管理员 ${row.username} 吗？`"
            @confirm="onDelete(row)"
          >
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
            {{ $t('ui.actionTitle.create', [$t('admin.admin.name')]) }}
          </VbenButton>
          <Popconfirm title="确认删除选中的管理员吗？" @confirm="onDelete()">
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
