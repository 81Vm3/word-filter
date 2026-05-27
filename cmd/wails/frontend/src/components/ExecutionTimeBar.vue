<template>
  <div class="time-bar">
    <div>
      <strong>执行时间</strong>
      <span class="time-value">{{ formattedTime }}</span>
    </div>
    <div v-if="error" class="time-error">{{ error }}</div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'ExecutionTimeBar',
  props: {
    elapsedNs: {
      type: Number,
      default: 0,
    },
    elapsedMs: {
      type: Number,
      default: 0,
    },
    error: {
      type: String,
      default: '',
    },
  },
  computed: {
    formattedTime(): string {
      const ns = Number(this.elapsedNs) || 0;
      const ms = Number(this.elapsedMs) || 0;
      if (ms >= 1) {
        return `${ms.toFixed(3)} ms`;
      }
      if (ms > 0) {
        return `${(ms * 1000).toFixed(3)} us`;
      }
      if (ns > 0) {
        if (ns >= 1000) {
          return `${(ns / 1000).toFixed(3)} us`;
        }
        return `${ns} ns`;
      }
      return '0 ns';
    },
  },
});
</script>
