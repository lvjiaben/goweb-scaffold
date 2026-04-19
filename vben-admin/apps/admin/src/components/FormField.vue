<script setup lang="ts">
defineProps<{
  label: string;
  required?: boolean;
  hint?: string;
}>();
</script>

<template>
  <div class="form-field">
    <span class="form-field__label">
      {{ label }}
      <em v-if="required">*</em>
    </span>
    <slot />
    <small v-if="hint" class="form-field__hint">{{ hint }}</small>
  </div>
</template>
