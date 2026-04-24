<script lang="ts" setup>
import type { VxeTableGridOptions } from '#/adapter/vxe-table';

import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

import { Page, VbenButton, VbenButtonGroup, useVbenDrawer } from '@vben/common-ui';
import { Plus, X } from '@vben/icons';
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
const route = useRoute();
const searchFormSchema = useSearchFormSchema();
const searchFormFields = searchFormSchema.map((item) => String(item.fieldName));

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
      reserve: true,
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
  formDrawerApi.setData({}).open();
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
      <template #is_super="{ row }">
        <Tag :color="row.is_super ? 'gold' : 'default'">
          {{ row.is_super ? '是' : '否' }}
        </Tag>
      </template>
      <template #role_names="{ row }">
        <span>{{ Array.isArray(row.role_names) ? row.role_names.join('、') : '-' }}</span>
      </template>
      <template #operation="{ row }">
        <VbenButtonGroup border>
          <VbenButton variant="icon" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm
            :title="`确认删除管理员 ${row.username} 吗？`"
            @confirm="onDelete(row)"
          >
            <VbenButton variant="icon" status="danger">
              删除
            </VbenButton>
          </Popconfirm>
        </VbenButtonGroup>
      </template>
      <template #toolbar-actions>
        <VbenButton class="mr-3" size="sm" variant="outline" @click="onCreate">
          <Plus class="size-3" />
          {{ $t('ui.actionTitle.create', [$t('admin.admin.name')]) }}
        </VbenButton>
        <VbenButtonGroup v-show="selectedCount > 0" border>
          <Popconfirm title="确认删除选中的管理员吗？" @confirm="onDelete()">
            <VbenButton variant="outline">
              <X class="size-3" />
              删除
            </VbenButton>
          </Popconfirm>
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
