<script lang="ts" setup>
import type { Recordable } from '@vben/types';
import type { VbenFormSchema } from '#/adapter/form';

import { computed, nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { getMenuOptions, saveMenu } from '#/api/admin/menu';
import { $t } from '#/locales';
import { componentKeys } from '#/router/routes';

import type { AdminMenuApi } from '#/api/admin/menu';

import {
  getMenuTypeOptions,
  getStatusOptions,
  getVisibleOptions,
} from '../data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<AdminMenuApi.AdminMenu>();

const schema: VbenFormSchema[] = [
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: getMenuTypeOptions(),
      optionType: 'button',
    },
    defaultValue: 'menu',
    fieldName: 'type',
    formItemClass: 'col-span-2 md:col-span-2',
    label: $t('admin.menu.type'),
  },
  {
    component: 'ApiTreeSelect',
    componentProps: {
      allowClear: true,
      api: getMenuOptions,
      childrenField: 'children',
      class: 'w-full',
      filterTreeNode(input: string, node: Recordable<any>) {
        if (!input || input.length === 0) {
          return true;
        }
        const title = String(node.title ?? node.label ?? '');
        const name = String(node.name ?? '');
        return title.includes(input) || name.includes(input);
      },
      getPopupContainer,
      labelField: 'title',
      placeholder: $t('admin.menu.parentPlaceholder'),
      showSearch: true,
      treeDefaultExpandAll: true,
      valueField: 'id',
    },
    fieldName: 'pid',
    label: $t('admin.menu.parent'),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: $t('admin.menu.menuNamePlaceholder'),
    },
    fieldName: 'title',
    label: $t('admin.menu.menuTitle'),
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('admin.menu.menuTitle'), 2]))
      .max(50, $t('ui.formRules.maxLength', [$t('admin.menu.menuTitle'), 50])),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: $t('admin.menu.menuNameEnPlaceholder'),
    },
    fieldName: 'name',
    label: $t('admin.menu.menuName'),
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('admin.menu.menuName'), 2]))
      .max(50, $t('ui.formRules.maxLength', [$t('admin.menu.menuName'), 50])),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: $t('admin.menu.pathPlaceholder'),
    },
    dependencies: {
      show: (values) => values.type === 'menu',
      triggerFields: ['type'],
    },
    fieldName: 'path',
    label: $t('admin.menu.path'),
    rules: z
      .string()
      .min(1, $t('ui.formRules.requiredSelect', [$t('admin.menu.path')]))
      .max(120, $t('ui.formRules.maxLength', [$t('admin.menu.path'), 120])),
  },
  {
    component: 'AutoComplete',
    componentProps: {
      class: 'w-full',
      filterOption(input: string, option: { value: string }) {
        return option.value.toLowerCase().includes(input.toLowerCase());
      },
      options: componentKeys.map((value) => ({ value })),
      placeholder: $t('admin.menu.componentPlaceholder'),
    },
    dependencies: {
      show: (values) => values.type === 'menu',
      triggerFields: ['type'],
    },
    fieldName: 'component',
    label: $t('admin.menu.component'),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: $t('admin.menu.authCode'),
    },
    dependencies: {
      show: (values) => values.type === 'button',
      triggerFields: ['type'],
    },
    fieldName: 'permission',
    label: $t('admin.menu.authCode'),
    rules: z.string().optional().or(z.literal('')),
  },
  {
    component: 'IconPicker',
    componentProps: {
      placeholder: $t('admin.menu.iconPlaceholder'),
      prefix: 'carbon',
    },
    fieldName: 'icon',
    label: $t('admin.menu.icon'),
  },
  {
    component: 'InputNumber',
    componentProps: {
      min: 0,
      precision: 0,
      style: { width: '100%' },
    },
    fieldName: 'sort',
    label: $t('common.sort'),
  },
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: getVisibleOptions(),
      optionType: 'button',
    },
    defaultValue: 1,
    fieldName: 'visible',
    label: $t('common.show'),
  },
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: getStatusOptions(),
      optionType: 'button',
    },
    defaultValue: 1,
    fieldName: 'status',
    label: $t('admin.menu.status'),
  },
];

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    formItemClass: 'col-span-2 md:col-span-1',
  },
  schema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Drawer, drawerApi] = useVbenDrawer({
  onConfirm: onSubmit,
  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<AdminMenuApi.AdminMenu>();
      formApi.resetForm();
      formData.value = data?.id ? data : undefined;
      await nextTick();
      if (data) {
        formApi.setValues({
          ...data,
          pid: data.pid || undefined,
        });
      }
    }
  },
});

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) {
    return;
  }
  drawerApi.lock();
  try {
    const values =
      await formApi.getValues<
        Omit<
          AdminMenuApi.AdminMenu,
          'children' | 'created_at' | 'id' | 'updated_at'
        >
      >();
    await saveMenu(formData.value?.id || 0, {
      ...values,
      pid: Number(values.pid ?? 0),
      sort: Number(values.sort ?? 0),
      status: Number(values.status ?? 1),
      visible: Number(values.visible ?? 1),
    });
    drawerApi.close();
    emit('success');
  } finally {
    drawerApi.unlock();
  }
}

const getDrawerTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('admin.menu.name')])
    : $t('ui.actionTitle.create', [$t('admin.menu.name')]),
);
</script>

<template>
  <Drawer class="w-full max-w-[720px]" :title="getDrawerTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Drawer>
</template>
