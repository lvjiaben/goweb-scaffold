<script setup lang="ts">
import type { MenuOption } from '@/types';

defineOptions({ name: 'MenuTreeSelect' });

const props = withDefaults(
  defineProps<{
    options: MenuOption[];
    modelValue: number[] | number;
    multiple?: boolean;
    disabledValues?: number[];
    level?: number;
  }>(),
  {
    multiple: false,
    disabledValues: () => [],
    level: 0,
  },
);

const emit = defineEmits<{
  'update:modelValue': [value: number[] | number];
}>();

function isChecked(value: number) {
  if (props.multiple) {
    return Array.isArray(props.modelValue) && props.modelValue.includes(value);
  }
  return Number(props.modelValue) === value;
}

function toggle(value: number) {
  if (props.multiple) {
    const current = Array.isArray(props.modelValue) ? [...props.modelValue] : [];
    const next = current.includes(value) ? current.filter((item) => item !== value) : [...current, value];
    emit('update:modelValue', next);
    return;
  }
  emit('update:modelValue', Number(props.modelValue) === value ? 0 : value);
}
</script>

<template>
  <div class="tree-select">
    <div
      v-for="item in options"
      :key="item.value"
      class="tree-select__item"
      :style="{ paddingLeft: `${level * 18}px` }"
    >
      <label class="tree-select__label">
        <input
          :type="multiple ? 'checkbox' : 'radio'"
          :checked="isChecked(item.value)"
          :disabled="disabledValues.includes(item.value)"
          @change="toggle(item.value)"
        />
        <span>{{ item.label }}</span>
        <small>{{ item.menu_type }}</small>
      </label>
      <MenuTreeSelect
        v-if="item.children?.length"
        :options="item.children"
        :model-value="modelValue"
        :multiple="multiple"
        :disabled-values="disabledValues"
        :level="level + 1"
        @update:model-value="emit('update:modelValue', $event)"
      />
    </div>
  </div>
</template>
