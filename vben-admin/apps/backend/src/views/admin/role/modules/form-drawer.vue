<script lang="ts" setup>
import type { DataNode } from 'ant-design-vue/es/tree';
import type { VbenFormSchema } from '#/adapter/form';

import { computed, nextTick, ref } from 'vue';

import { Tree, useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { Spin } from 'ant-design-vue';
import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { getMenuTree, saveRole } from '#/api/admin/role';

import type { AdminRoleApi } from '#/api/admin/role';

import { getRoleStatusOptions } from '../data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<AdminRoleApi.AdminRole>();
const permissions = ref<DataNode[]>([]);
const loadingPermissions = ref(false);

const schema: VbenFormSchema[] = [
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: getRoleStatusOptions(),
      optionType: 'button',
    },
    defaultValue: 1,
    fieldName: 'status',
    label: $t('admin.role.status'),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: $t('admin.role.roleNamePlaceholder'),
    },
    fieldName: 'name',
    label: $t('admin.role.roleName'),
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', [$t('admin.role.roleName'), 2]))
      .max(64, $t('ui.formRules.maxLength', [$t('admin.role.roleName'), 64])),
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: '请输入角色编码',
    },
    fieldName: 'code',
    label: '角色编码',
    rules: z
      .string()
      .min(2, $t('ui.formRules.minLength', ['角色编码', 2]))
      .max(64, $t('ui.formRules.maxLength', ['角色编码', 64])),
  },
  {
    component: 'Input',
    fieldName: 'menu_ids',
    formItemClass: 'col-span-2 md:col-span-2 items-start',
    label: $t('admin.role.permissions'),
    modelPropName: 'modelValue',
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
      const data = drawerApi.getData<AdminRoleApi.AdminRole>();
      formApi.resetForm();
      formData.value = data?.id ? data : undefined;
      if (permissions.value.length === 0) {
        await loadPermissions();
      }
      await nextTick();
      if (data) {
        formApi.setValues({
          ...data,
          menu_ids: data.menu_ids ?? [],
        });
      }
    }
  },
});

async function loadPermissions() {
  loadingPermissions.value = true;
  try {
    permissions.value = (await getMenuTree()) as unknown as DataNode[];
  } finally {
    loadingPermissions.value = false;
  }
}

async function onSubmit() {
  const { valid } = await formApi.validate();
  if (!valid) {
    return;
  }
  drawerApi.lock();
  try {
    const values =
      await formApi.getValues<
        Omit<AdminRoleApi.AdminRole, 'created_at' | 'id' | 'updated_at'>
      >();
    await saveRole(formData.value?.id || 0, {
      ...values,
      menu_ids: values.menu_ids ?? [],
      status: Number(values.status ?? 1),
    });
    drawerApi.close();
    emit('success');
  } finally {
    drawerApi.unlock();
  }
}

const getDrawerTitle = computed(() =>
  formData.value?.id
    ? $t('ui.actionTitle.edit', [$t('admin.role.name')])
    : $t('ui.actionTitle.create', [$t('admin.role.name')]),
);
</script>

<template>
  <Drawer class="w-full max-w-[640px]" :title="getDrawerTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'">
      <template #menu_ids="slotProps">
        <Spin :spinning="loadingPermissions" wrapper-class-name="w-full">
          <Tree
            :tree-data="permissions"
            bordered
            multiple
            :default-expanded-level="2"
            label-field="name"
            value-field="id"
            children-field="children"
            v-bind="slotProps"
          />
        </Spin>
      </template>
    </Form>
  </Drawer>
</template>
