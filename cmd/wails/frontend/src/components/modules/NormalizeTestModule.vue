<template>
  <section class="module-grid module-grid-normalize">
    <article class="panel">
      <h2>归一化输入</h2>
      <label class="field">
        <span class="field-label">原始文本</span>
        <textarea class="text-area" rows="8" v-model="text" />
      </label>
    </article>

    <article class="panel">
      <h2>配置</h2>
      <ScaleInput v-model="scale" />
      <label class="field inline">
        <input type="checkbox" v-model="ignoreCase" />
        <span>Ignore Case</span>
      </label>
      <label class="field inline">
        <input type="checkbox" v-model="ignoreWidth" />
        <span>Ignore Width</span>
      </label>
      <button class="run-btn" :disabled="loading" @click="runTest">
        {{ loading ? '执行中...' : '执行归一化测试' }}
      </button>
    </article>

    <article class="panel panel-wide">
      <ExecutionTimeBar
        :elapsed-ns="response.elapsed_ns || 0"
        :elapsed-ms="response.elapsed_ms || 0"
        :error="response.error || ''"
      />
      <ResultPanel title="归一化结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { normalizeTest } from '../../api';
import type { NormalizeTestResponse } from '../../types';
import ExecutionTimeBar from '../ExecutionTimeBar.vue';
import ResultPanel from '../ResultPanel.vue';
import ScaleInput from '../ScaleInput.vue';

export default defineComponent({
  name: 'NormalizeTestModule',
  components: { ScaleInput, ExecutionTimeBar, ResultPanel },
  data() {
    return {
      text: 'ＡBC badword Hello 世界',
      ignoreCase: true,
      ignoreWidth: true,
      scale: 1,
      loading: false,
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<NormalizeTestResponse>,
    };
  },
  methods: {
    async runTest() {
      this.loading = true;
      this.response = await normalizeTest({
        text: this.text,
        ignore_case: this.ignoreCase,
        ignore_width: this.ignoreWidth,
        scale: this.scale,
      });
      this.loading = false;
    },
  },
});
</script>
