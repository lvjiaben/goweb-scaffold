import type { VbenFormProps } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { AdminAdminApi } from '#/api/admin/admin';

import { $t } from '#/locales';

function formatDateTime(cellValue?: null | number | string) {
  if (!cellValue) {
    return '-';
  }
  if (typeof cellValue === 'number') {
    return new Date(cellValue * 1000).toLocaleString();
  }
  return new Date(cellValue).toLocaleString();
}

export function getAdminStatusOptions() {
  return [
    { color: 'success', label: $t('common.enable'), value: 1 },
    { color: 'error', label: $t('common.disable'), value: 0 },
  ];
}

export function useSearchFormSchema(): VbenFormProps['schema'] {
  return [
    {
      component: 'Input',
      fieldName: 'id',
      label: 'ID',
    },
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('admin.admin.username'),
    },
    {
      component: 'Input',
      fieldName: 'realname',
      label: $t('admin.admin.realname'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          {
            label: $t('common.all'),
            value: '',
          },
          {
            label: $t('common.enable'),
            value: '1',
          },
          {
            label: $t('common.disable'),
            value: '0',
          },
        ],
      },
      fieldName: 'status',
      label: $t('admin.admin.status'),
    },
    {
      component: 'RangePicker',
      componentProps: {
        format: 'YYYY-MM-DD HH:mm:ss',
        showTime: true,
        timestampFormat: true,
      },
      fieldName: 'created_at',
      label: $t('common.createTime'),
    },
  ];
}

export function useColumns(): VxeTableGridOptions<AdminAdminApi.Admin>['columns'] {
  return [
    {
      type: 'checkbox',
      width: 50,
      align: 'center',
      fixed: 'left',
    },
    {
      align: 'center',
      field: 'id',
      title: 'ID',
      width: 80,
      sortable: true,
    },
    {
      align: 'left',
      field: 'username',
      title: $t('admin.admin.username'),
      minWidth: 140,
    },
    {
      align: 'left',
      field: 'realname',
      title: $t('admin.admin.realname'),
      minWidth: 140,
    },
    {
      align: 'center',
      field: 'is_super',
      slots: { default: 'is_super' },
      title: '超级管理员',
      width: 120,
    },
    {
      align: 'left',
      field: 'role_names',
      slots: { default: 'role_names' },
      title: $t('admin.admin.roles'),
      minWidth: 180,
    },
    {
      align: 'center',
      field: 'status',
      slots: { default: 'status' },
      title: $t('admin.admin.status'),
      width: 100,
    },
    {
      align: 'center',
      field: 'created_at',
      formatter: ({ cellValue }) => formatDateTime(cellValue),
      title: $t('common.createTime'),
      width: 180,
      sortable: true,
    },
    {
      field: 'operation',
      fixed: 'right',
      headerAlign: 'center',
      slots: { default: 'operation' },
      title: $t('admin.admin.operation'),
      width: 180,
    },
  ];
}
