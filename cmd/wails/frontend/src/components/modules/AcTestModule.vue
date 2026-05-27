<template>
  <section class="module-grid module-grid-ac">
    <article class="panel">
      <h2>AC 自动机输入</h2>
      <label class="field">
        <span class="field-label">测试文本</span>
        <textarea class="text-area" rows="9" v-model="text" />
      </label>
    </article>

    <article class="panel">
      <h2>参数</h2>
      <ScaleInput v-model="scale" />
      <button class="run-btn" :disabled="loading" @click="runTest">
        {{ loading ? '执行中...' : '执行 AC 测试' }}
      </button>
      <p class="hint">该模块逐字符驱动 Step，展示命中序列与统计。</p>
    </article>

    <article class="panel panel-wide">
      <ExecutionTimeBar
        :elapsed-ns="response.elapsed_ns || 0"
        :elapsed-ms="response.elapsed_ms || 0"
        :error="response.error || ''"
      />
      <ResultPanel title="AC 测试结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { acTest } from '../../api';
import type { ACTestResponse } from '../../types';
import ExecutionTimeBar from '../ExecutionTimeBar.vue';
import ResultPanel from '../ResultPanel.vue';
import ScaleInput from '../ScaleInput.vue';

export default defineComponent({
  name: 'AcTestModule',
  components: { ScaleInput, ExecutionTimeBar, ResultPanel },
  data() {
    return {
      text: 'badword and 傻逼 and ＢＡＤＷＯＲＤ',
      scale: 1,
      loading: false,
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<ACTestResponse>,
    };
  },
  methods: {
    async runTest() {
      this.loading = true;
      this.response = await acTest({
        text: this.text,
        scale: this.scale,
      });
      this.loading = false;
    },
  },
});
</script>
