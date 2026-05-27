<template>
  <section class="module-grid module-grid-trie">
    <article class="panel">
      <h2>字典树操作</h2>
      <label class="field">
        <span class="field-label">操作</span>
        <select class="text-input" v-model="action">
          <option value="add">Add</option>
          <option value="update">Update</option>
          <option value="delete">Delete</option>
          <option value="has">Has</option>
          <option value="get">Get</option>
          <option value="walk">Walk</option>
        </select>
      </label>
      <label class="field">
        <span class="field-label">词条</span>
        <input class="text-input" v-model="word" />
      </label>
      <label class="field">
        <span class="field-label">Severity</span>
        <input class="text-input" type="number" v-model.number="severity" />
      </label>
      <label class="field">
        <span class="field-label">Extra JSON</span>
        <textarea class="text-area" rows="3" v-model="extraJSON" />
      </label>
    </article>

    <article class="panel">
      <h2>执行参数</h2>
      <ScaleInput v-model="scale" />
      <button class="run-btn" :disabled="loading" @click="runTest">
        {{ loading ? '执行中...' : '执行字典树测试' }}
      </button>
      <p class="hint">Add/Update/Delete 在 scale > 1 时会自动生成带后缀词条。</p>
    </article>

    <article class="panel panel-wide">
      <ExecutionTimeBar
        :elapsed-ns="response.elapsed_ns || 0"
        :elapsed-ms="response.elapsed_ms || 0"
        :error="response.error || ''"
      />
      <ResultPanel title="字典树结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { trieTest } from '../../api';
import type { TrieTestResponse } from '../../types';
import ExecutionTimeBar from '../ExecutionTimeBar.vue';
import ResultPanel from '../ResultPanel.vue';
import ScaleInput from '../ScaleInput.vue';

export default defineComponent({
  name: 'TrieTestModule',
  components: { ScaleInput, ExecutionTimeBar, ResultPanel },
  data() {
    return {
      action: 'add',
      word: 'badword',
      severity: 1,
      extraJSON: '',
      scale: 1,
      loading: false,
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<TrieTestResponse>,
    };
  },
  methods: {
    async runTest() {
      this.loading = true;
      this.response = await trieTest({
        action: this.action,
        word: this.word,
        severity: this.severity,
        extra_json: this.extraJSON,
        scale: this.scale,
      });
      this.loading = false;
    },
  },
});
</script>
