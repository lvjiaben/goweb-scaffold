<script setup lang="ts">
withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    width?: string;
  }>(),
  {
    width: '720px',
  },
);

defineEmits<{
  close: [];
}>();
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="modal-overlay" @click.self="$emit('close')">
      <section class="modal-panel card" :style="{ width }">
        <header class="modal-header">
          <h3>{{ title }}</h3>
          <button class="icon-btn" type="button" @click="$emit('close')">×</button>
        </header>
        <div class="modal-body">
          <slot />
        </div>
        <footer class="modal-footer">
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </Teleport>
</template>
