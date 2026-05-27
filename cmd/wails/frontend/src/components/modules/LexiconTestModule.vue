<template>
  <section class="module-grid module-grid-lexicon">
    <article class="panel">
      <h2>词库操作</h2>
      <label class="field">
        <span class="field-label">操作</span>
        <select class="text-input" v-model="action">
          <option value="import">Import TXT/CSV</option>
          <option value="add">Add</option>
          <option value="update">Update</option>
          <option value="delete">Delete</option>
          <option value="list">List</option>
          <option value="save">Save words.txt</option>
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
      <h2>导入与分页</h2>
      <label class="field">
        <span class="field-label">导入文件（TXT/CSV）</span>
        <input type="file" class="text-input" accept=".txt,.csv,text/plain,text/csv" @change="onFileChange" />
      </label>
      <p class="hint" v-if="fileName">导入文件: {{ fileName }}</p>
      <ScaleInput v-model="scale" />
      <label class="field">
        <span class="field-label">页码</span>
        <input class="text-input" type="number" min="1" v-model.number="page" />
      </label>
      <label class="field">
        <span class="field-label">每页数量</span>
        <input class="text-input" type="number" min="1" max="200" v-model.number="pageSize" />
      </label>
      <button class="run-btn" :disabled="loading" @click="runTest">
        {{ loading ? '执行中...' : '执行词库测试' }}
      </button>
    </article>

    <article class="panel panel-wide">
      <ExecutionTimeBar
        :elapsed-ns="response.elapsed_ns || 0"
        :elapsed-ms="response.elapsed_ms || 0"
        :error="response.error || ''"
      />
      <ResultPanel title="词库结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { lexiconTest } from '../../api';
import type { LexiconTestResponse } from '../../types';
import ExecutionTimeBar from '../ExecutionTimeBar.vue';
import ResultPanel from '../ResultPanel.vue';
import ScaleInput from '../ScaleInput.vue';

export default defineComponent({
  name: 'LexiconTestModule',
  components: { ScaleInput, ExecutionTimeBar, ResultPanel },
  data() {
    return {
      action: 'list',
      word: 'newword',
      severity: 0,
      extraJSON: '',
      fileName: '',
      fileContent: '',
      scale: 1,
      page: 1,
      pageSize: 20,
      loading: false,
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<LexiconTestResponse>,
    };
  },
  methods: {
    async onFileChange(event: Event) {
      const target = event.target as HTMLInputElement;
      const file = target.files?.[0];
      if (!file) {
        return;
      }
      this.fileName = file.name;
      this.fileContent = await file.text();
    },
    async runTest() {
      this.loading = true;
      this.response = await lexiconTest({
        action: this.action,
        word: this.word,
        severity: this.severity,
        extra_json: this.extraJSON,
        scale: this.scale,
        input_file_name: this.fileName,
        input_file_content: this.fileContent,
        page: this.page,
        page_size: this.pageSize,
      });
      this.loading = false;
    },
  },
});
</script>
