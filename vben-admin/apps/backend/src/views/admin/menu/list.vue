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
const route = useRoute();
const searchFormSchema = useSearchFormSchema();
const searchFormFields = searchFormSchema.map((item) => String(item.fieldName));

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
    schema: searchFormSchema,
    showCollapseButton: true,
    submitOnChange: false,
    submitOnEnter: false,
  },
  gridOptions: {
    columns: useColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {
      enabled: false,
    },
    proxyConfig: {
      ajax: {
        query: async (params, formValues) => {
          const sort = params?.sort as { field?: string; order?: string } | undefined;
          return await getMenuList({
            filter: JSON.stringify(formValues ?? {}),
            page: 1,
            page_size: 0,
            search: searchValue.value || undefined,
            sort_by: sort?.field ? String(sort.field) : 'sort',
            sort_order: sort?.order === 'asc' ? 'asc' : 'desc',
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
    treeConfig: {
      childrenField: 'children',
      expandAll: false,
      parentField: 'parent_id',
      rowField: 'id',
      transform: false,
    },
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
      enname: '',
      external: '',
      icon: '',
      iframe: '',
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
      enname: '',
      external: '',
      icon: '',
      iframe: '',
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
  <Page auto-content-height>
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
        <VbenButtonGroup border>
          <VbenButton variant="icon" @click="onCreateChild(row)">
            新增下级
          </VbenButton>
          <VbenButton variant="icon" @click="onEdit(row)">
            编辑
          </VbenButton>
          <Popconfirm :title="`确认删除菜单 ${row.title} 吗？`" @confirm="onDelete(row)">
            <VbenButton variant="icon" status="danger">
              删除
            </VbenButton>
          </Popconfirm>
        </VbenButtonGroup>
      </template>
      <template #toolbar-actions>
        <VbenButton class="mr-3" size="sm" variant="outline" @click="onCreate">
          <Plus class="size-3" />
          {{ $t('ui.actionTitle.create', [$t('admin.menu.name')]) }}
        </VbenButton>
        <VbenButtonGroup v-show="selectedCount > 0" border>
          <Popconfirm title="确认删除选中的菜单吗？" @confirm="onDelete()">
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
