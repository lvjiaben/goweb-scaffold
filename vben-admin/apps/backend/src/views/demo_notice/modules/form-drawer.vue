<script lang="ts" setup>
import type { VbenFormSchema } from '#/adapter/form';
import type { DemoNoticeApi } from '#/api/demo_notice';

import { computed, nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { breakpointsTailwind, useBreakpoints } from '@vueuse/core';

import { useVbenForm, z } from '#/adapter/form';
import { saveDemoNotice } from '#/api/demo_notice';
import { getDemoNoticeStatusOptions } from '../data';

const emit = defineEmits<{
  success: [];
}>();

const formData = ref<DemoNoticeApi.DemoNotice>();
const statusOptions = getDemoNoticeStatusOptions();

const schema = computed((): VbenFormSchema[] => {
  return [
    {
      component: "Input",
      componentProps: {
        placeholder: "请输入公告标题",
      },
      fieldName: "title",
      formItemClass: 'col-span-2 md:col-span-1',
      label: "公告标题",
      rules: z.string().min(1, "公告标题不能为空"),
    },
    {
      component: "Textarea",
      componentProps: {
        placeholder: "请输入公告正文",
        rows: 4,
      },
      fieldName: "content",
      formItemClass: 'col-span-2 md:col-span-1',
      label: "公告内容",
      rules: z.string().min(1, "公告内容不能为空"),
    },
    {
      component: "RadioGroup",
      componentProps: {
        buttonStyle: 'solid',
        options: statusOptions,
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: "status",
      formItemClass: 'col-span-2 md:col-span-1',
      label: "状态",
    },
    {
      component: "InputNumber",
      componentProps: {
        class: 'w-full',
        placeholder: "请输入数字",
      },
      fieldName: "sort",
      formItemClass: 'col-span-2 md:col-span-1',
      label: "排序值",
    },
  ];
});

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
      const data = drawerApi.getData<DemoNoticeApi.DemoNotice>();
      formApi.resetForm();
      formData.value = data?.id ? data : undefined;
      await nextTick();
      if (data) {
        formApi.setValues({
          ...data,
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
    const values = await formApi.getValues<Partial<DemoNoticeApi.DemoNotice>>();
    await saveDemoNotice(formData.value?.id || 0, values);
    drawerApi.close();
    emit('success');
  } finally {
    drawerApi.unlock();
  }
}

const getDrawerTitle = computed(() =>
  formData.value?.id ? '编辑演示公告' : '新增演示公告',
);
</script>

<template>
  <Drawer class="w-full max-w-[720px]" :title="getDrawerTitle">
    <Form class="mx-4" :layout="isHorizontal ? 'horizontal' : 'vertical'" />
  </Drawer>
</template>
