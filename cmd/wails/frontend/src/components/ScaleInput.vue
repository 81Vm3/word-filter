<template>
  <label class="field">
    <span class="field-label">数据量级</span>
    <input
      class="text-input"
      type="number"
      min="1"
      step="1"
      :value="modelValue"
      @input="onInput"
    />
  </label>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'ScaleInput',
  props: {
    modelValue: {
      type: Number,
      required: true,
    },
  },
  emits: ['update:modelValue'],
  methods: {
    onInput(event: Event) {
      const target = event.target as HTMLInputElement;
      const next = Number.parseInt(target.value, 10);
      if (Number.isNaN(next) || next < 1) {
        this.$emit('update:modelValue', 1);
        return;
      }
      this.$emit('update:modelValue', next);
    },
  },
});
</script>
