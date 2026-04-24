<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';

import { computed, nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { saveAdmin } from '#/api/admin/admin';
import { getRoleOptions } from '#/api/admin/role';
import { $t } from '#/locales';

import { message } from 'ant-design-vue';

import type { AdminAdminApi } from '#/api/admin/admin';

import { getAdminStatusOptions } from '../data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<AdminAdminApi.Admin>();

function buildSchema(isEdit: boolean): VbenFormSchema[] {
  return [
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: getAdminStatusOptions(),
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: $t('admin.admin.status'),
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('admin.admin.usernamePlaceholder'),
      },
      fieldName: 'username',
      label: $t('admin.admin.username'),
      rules: z
        .string()
        .min(3, $t('ui.formRules.minLength', [$t('admin.admin.username'), 3]))
        .max(32, $t('ui.formRules.maxLength', [$t('admin.admin.username'), 32])),
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('admin.admin.realnamePlaceholder'),
      },
      fieldName: 'realname',
      label: $t('admin.admin.realname'),
      rules: z
        .string()
        .min(2, $t('ui.formRules.minLength', [$t('admin.admin.realname'), 2]))
        .max(32, $t('ui.formRules.maxLength', [$t('admin.admin.realname'), 32])),
    },
    {
      component: 'InputPassword',
      componentProps: {
        placeholder: isEdit
          ? '编辑时留空则不修改密码'
          : `${$t('admin.admin.passwordPlaceholder')}（必填）`,
      },
      fieldName: 'password',
      label: $t('admin.admin.password'),
      rules: z
        .string()
        .min(6, $t('ui.formRules.minLength', [$t('admin.admin.password'), 6]))
        .max(32, $t('ui.formRules.maxLength', [$t('admin.admin.password'), 32]))
        .optional()
        .or(z.literal('')),
    },
    {
      component: 'ApiSelect',
      componentProps: {
        allowClear: true,
        api: getRoleOptions,
        class: 'w-full',
        labelField: 'name',
        mode: 'multiple',
        placeholder: $t('admin.admin.rolesPlaceholder'),
        valueField: 'id',
      },
      fieldName: 'role_ids',
      formItemClass: 'col-span-2 md:col-span-2',
      label: $t('admin.admin.roles'),
    },
  ];
}

const breakpoints = useBreakpoints(breakpointsTailwind);
const isHorizontal = computed(() => breakpoints.greaterOrEqual('md').value);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    colon: true,
    formItemClass: 'col-span-2 md:col-span-1',
  },
  schema: buildSchema(false),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const [Drawer, drawerApi] = useVbenDrawer({
  onConfirm: onSubmit,
  async onOpenChange(isOpen) {
    if (isOpen) {
      const data = drawerApi.getData<AdminAdminApi.Admin>();
      formApi.resetForm();
      formData.value = data?.id ? data : undefined;
      formApi.updateSchema([
        {
          componentProps: {
            placeholder: formData.value?.id
              ? '编辑时留空则不修改密码'
              : `${$t('admin.admin.passwordPlaceholder')}（必填）`,
          },
          fieldName: 'password',
        },
      ]);
      await nextTick();
      if (data) {
        formApi.setValues({
          ...data,
          password: '',
          role_ids: data.role_ids ?? [],
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
          AdminAdminApi.Admin,
          'created_at' | 'id' | 'is_super' | 'role_names' | 'updated_at'
        >
      >();
    if (!formData.value?.id && !values.password) {
      message.warning(`${$t('admin.admin.password')}必填`);
      return;
    }
    if (formData.value?.id && !values.password) {
      delete values.password;
    }
    await saveAdmin(formData.value?.id || 0, {
      ...values,
      role_ids: values.role_ids ?? [],
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
    ? $t('ui.actionTitle.edit', [$t('admin.admin.name')])
    : $t('ui.actionTitle.create', [$t('admin.admin.name')]),
);
</script>

<template>
  <Drawer class="w-full max-w-[600px]" :title="getDrawerTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Drawer>
</template>
