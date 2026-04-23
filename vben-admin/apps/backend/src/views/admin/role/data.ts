import type { VbenFormProps } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { AdminRoleApi } from '#/api/admin/role';

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

export function getRoleStatusOptions() {
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
      fieldName: 'name',
      label: $t('admin.role.roleName'),
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '角色编码',
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('common.all'), value: '' },
          { label: $t('common.enable'), value: '1' },
          { label: $t('common.disable'), value: '0' },
        ],
      },
      fieldName: 'status',
      label: $t('admin.role.status'),
    },
  ];
}

export function useColumns(): VxeTableGridOptions<AdminRoleApi.AdminRole>['columns'] {
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
      field: 'name',
      title: $t('admin.role.roleName'),
      minWidth: 160,
    },
    {
      align: 'left',
      field: 'code',
      title: '角色编码',
      minWidth: 180,
    },
    {
      align: 'center',
      field: 'status',
      slots: { default: 'status' },
      title: $t('admin.role.status'),
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
      title: $t('admin.role.operation'),
      width: 180,
    },
  ];
}
