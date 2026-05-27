<template>
  <aside class="sidebar">
    <h1 class="brand">海量敏感词过滤系统</h1>
    <nav class="module-list" aria-label="Test Modules">
      <button
        v-for="mod in modules"
        :key="mod.key"
        class="module-btn"
        :class="{ active: current === mod.key }"
        @click="$emit('switch', mod.key)"
      >
        {{ mod.label }}
      </button>
    </nav>
  </aside>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

export default defineComponent({
  name: 'SidebarNav',
  props: {
    modules: {
      type: Array as PropType<Array<{ key: string; label: string }>>,
      required: true,
    },
    current: {
      type: String,
      required: true,
    },
    currentLabel: {
      type: String,
      required: true,
    },
  },
  emits: ['switch'],
});
</script>
