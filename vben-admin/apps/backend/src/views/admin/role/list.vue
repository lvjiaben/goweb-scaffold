<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';

import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

import { Page, VbenButton, VbenButtonGroup, useVbenDrawer } from '@vben/common-ui';
import { Check, Eraser, Plus, X } from '@vben/icons';
import { $t } from '@vben/locales';

import { InputSearch, message, Popconfirm, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteRole,
  getRoleDetail,
  getRoleList,
} from '#/api/admin/role';

import type { AdminRoleApi } from '#/api/admin/role';

import {
  getRoleStatusOptions,
  useColumns,
  useSearchFormSchema,
} from './data';
import FormDrawerComponent from './modules/form-drawer.vue';

const selectedCount = ref(0);
const searchValue = ref('');
const route = useRoute();
const searchFormSchema = useSearchFormSchema();
const searchFormFields = searchFormSchema.map((item) => String(item.fieldName));
const statusOptions = getRoleStatusOptions();

const getSelectedRecords = (): AdminRoleApi.AdminRole[] => {
  return (gridApi.grid?.getCheckboxRecords() as AdminRoleApi.AdminRole[]) || [];
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
    schema: searchFormSchema,
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
          return await getRoleList({
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
      reserve: true,
      trigger: 'row',
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

const onEdit = async (row: AdminRoleApi.AdminRole) => {
  const detail = await getRoleDetail(row.id);
  formDrawerApi.setData(detail).open();
};

const onCreate = () => {
  formDrawerApi
    .setData({
      code: '',
      menu_ids: [],
      name: '',
      status: 1,
    })
    .open();
};

const onDelete = (row?: AdminRoleApi.AdminRole) => {
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
  deleteRole({ ids })
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

onMounted(() => {
  const formValues: Record<string, any> = {};
  for (const [key, value] of Object.entries(route.query)) {
    if (searchFormFields.includes(key) && value) {
      formValues[key] = Array.isArray(value) ? value[0] : value;
    }
  }
  if (Object.keys(formValues).length > 0) {
    gridApi.formApi.setValues(formValues);
    gridApi.query();
  }
});
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
      <template #operation="{ row }">
        <VbenButtonGroup border>
          <VbenButton variant="icon" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm :title="`确认删除角色 ${row.name} 吗？`" @confirm="onDelete(row)">
            <VbenButton variant="icon" status="danger">
              删除
            </VbenButton>
          </Popconfirm>
        </VbenButtonGroup>
      </template>
      <template #toolbar-actions>
        <VbenButton class="mr-3" size="sm" variant="outline" @click="onCreate">
          <Plus class="size-3" />
          {{ $t('ui.actionTitle.create', [$t('admin.role.name')]) }}
        </VbenButton>
        <VbenButtonGroup v-show="selectedCount > 0" border>
          <Popconfirm title="确认删除选中的角色吗？" @confirm="onDelete()">
            <VbenButton variant="icon" status="danger">
              <Eraser class="size-4" />
            </VbenButton>
          </Popconfirm>
          <VbenButton variant="icon" @click="gridApi.grid?.setAllCheckboxRow(true)">
            <Check class="size-4" />
          </VbenButton>
          <VbenButton variant="icon" @click="gridApi.grid?.clearCheckboxRow()">
            <X class="size-4" />
          </VbenButton>
        </VbenButtonGroup>
      </template>
      <template #toolbar-tools>
        <InputSearch
          allow-clear
          :placeholder="$t('common.fuzzySearch')"
          @clear="onClearSearch"
          @search="onSearch"
        />
      </template>
    </Grid>
  </Page>
</template>
