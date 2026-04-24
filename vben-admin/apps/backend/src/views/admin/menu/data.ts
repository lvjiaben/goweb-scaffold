import type { VbenFormProps } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { AdminMenuApi } from '#/api/admin/menu';

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

export function getMenuTypeOptions() {
  return [
    { color: 'default', label: $t('admin.menu.typeMenu'), value: 'menu' },
    { color: 'error', label: $t('admin.menu.typeButton'), value: 'button' },
    { color: 'warning', label: 'Iframe', value: 'iframe' },
    { color: 'success', label: 'Link', value: 'link' },
  ];
}

export function getVisibleOptions() {
  return [
    { color: 'success', label: $t('common.show'), value: 1 },
    { color: 'default', label: $t('common.hide'), value: 0 },
  ];
}

export function getStatusOptions() {
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
      fieldName: 'title',
      label: $t('admin.menu.menuTitle'),
    },
    {
      component: 'Input',
      fieldName: 'path',
      label: $t('admin.menu.path'),
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        options: [
          { label: $t('common.all'), value: '' },
          { label: $t('admin.menu.typeMenu'), value: 'menu' },
          { label: $t('admin.menu.typeButton'), value: 'button' },
          { label: 'Iframe', value: 'iframe' },
          { label: 'Link', value: 'link' },
        ],
      },
      fieldName: 'type',
      label: $t('admin.menu.type'),
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
      label: $t('admin.menu.status'),
    },
  ];
}

export function useColumns(): VxeTableGridOptions<AdminMenuApi.AdminMenu>['columns'] {
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
      field: 'title',
      title: $t('admin.menu.menuTitle'),
      treeNode: true,
      minWidth: 220,
      slots: { default: 'title' },
    },
    {
      align: 'left',
      field: 'name',
      title: $t('admin.menu.menuName'),
      minWidth: 160,
    },
    {
      align: 'left',
      field: 'enname',
      title: $t('admin.menu.menuNameEn'),
      minWidth: 160,
    },
    {
      align: 'center',
      field: 'type',
      slots: { default: 'type' },
      title: $t('admin.menu.type'),
      width: 100,
    },
    {
      align: 'left',
      field: 'path',
      title: $t('admin.menu.path'),
      minWidth: 180,
    },
    {
      align: 'left',
      field: 'component',
      slots: { default: 'component' },
      title: $t('admin.menu.component'),
      minWidth: 220,
    },
    {
      align: 'center',
      field: 'visible',
      slots: { default: 'visible' },
      title: $t('common.show'),
      width: 100,
    },
    {
      align: 'center',
      field: 'status',
      slots: { default: 'status' },
      title: $t('admin.menu.status'),
      width: 100,
    },
    {
      align: 'center',
      field: 'sort',
      title: $t('common.sort'),
      width: 90,
      sortable: true,
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
      title: $t('admin.menu.operation'),
      width: 220,
    },
  ];
}
