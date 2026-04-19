<script setup lang="ts">
import { hasAccess } from '@/auth';

defineProps<{
  code: string;
}>();
</script>

<template>
  <slot v-if="hasAccess(code)" />
</template>
