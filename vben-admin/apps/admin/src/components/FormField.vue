<script setup lang="ts">
defineProps<{
  label: string;
  required?: boolean;
  hint?: string;
  error?: string;
}>();
</script>

<template>
  <div class="form-field">
    <span class="form-field__label">
      {{ label }}
      <em v-if="required">*</em>
    </span>
    <slot />
    <small v-if="$slots.hint" class="form-field__hint">
      <slot name="hint" />
    </small>
    <small v-else-if="hint" class="form-field__hint">{{ hint }}</small>
    <small v-if="$slots.error || error" class="form-field__error">
      <slot name="error">{{ error }}</slot>
    </small>
  </div>
</template>
