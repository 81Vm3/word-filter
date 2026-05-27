<template>
  <section class="result-panel">
    <header class="result-header">
      <h3>{{ title }}</h3>
      <button class="ghost-btn" @click="toggleView">
        {{ showPretty ? '原始 JSON' : '格式化 JSON' }}
      </button>
    </header>
    <pre class="result-body">{{ rendered }}</pre>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'ResultPanel',
  props: {
    title: {
      type: String,
      default: '结果',
    },
    payload: {
      type: [Object, Array, String, Number, Boolean],
      default: () => ({}),
    },
  },
  data() {
    return {
      showPretty: true,
    };
  },
  computed: {
    rendered(): string {
      if (this.showPretty) {
        return JSON.stringify(this.payload, null, 2);
      }
      return JSON.stringify(this.payload);
    },
  },
  methods: {
    toggleView() {
      this.showPretty = !this.showPretty;
    },
  },
});
</script>
