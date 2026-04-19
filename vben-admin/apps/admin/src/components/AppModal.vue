<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    width?: string;
    maskClosable?: boolean;
    escClosable?: boolean;
  }>(),
  {
    width: '720px',
    maskClosable: true,
    escClosable: true,
  },
);

const emit = defineEmits<{
  close: [];
}>();

function requestClose() {
  emit('close');
}

function onKeydown(event: KeyboardEvent) {
  if (!props.open || !props.escClosable) {
    return;
  }
  if (event.key === 'Escape') {
    requestClose();
  }
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      window.addEventListener('keydown', onKeydown);
      return;
    }
    window.removeEventListener('keydown', onKeydown);
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown);
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="modal-overlay"
      @click.self="maskClosable ? requestClose() : undefined"
    >
      <section class="modal-panel card" :style="{ width }">
        <header class="modal-header">
          <h3>{{ title }}</h3>
          <button class="icon-btn" type="button" @click="requestClose">×</button>
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
