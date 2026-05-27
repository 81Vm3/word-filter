<template>
  <section class="module-grid module-grid-filter">
    <article class="panel">
      <h2>过滤器输入</h2>
      <label class="field">
        <span class="field-label">输入方式</span>
        <select class="text-input" v-model="inputMode">
          <option value="manual">手动输入</option>
          <option value="txt">TXT 导入</option>
        </select>
      </label>

      <label class="field" v-if="inputMode === 'manual'">
        <span class="field-label">手动输入内容</span>
        <textarea v-model="inputText" class="text-area" rows="8" />
      </label>

      <label class="field" v-else>
        <span class="field-label">导入文件（TXT）</span>
        <input type="file" class="text-input" accept=".txt,text/plain" @change="onFileChange" />
      </label>
      <p class="hint" v-if="inputMode === 'txt' && fileName">已加载文件: {{ fileName }}</p>
    </article>

    <article class="panel">
      <h2>测试参数</h2>
      <ScaleInput v-model="scale" />
      <label class="field inline">
        <input type="checkbox" v-model="ignoreCase" />
        <span>忽略大小写</span>
      </label>
      <label class="field inline">
        <input type="checkbox" v-model="ignoreWidth" />
        <span>忽略全角半角</span>
      </label>
      <label class="field">
        <span class="field-label">替换字符</span>
        <input class="text-input" v-model="replaceRune" maxlength="1" />
      </label>
      <button class="run-btn" :disabled="loading" @click="runTest">
        {{ loading ? '执行中...' : '执行过滤器测试' }}
      </button>
    </article>

    <article class="panel panel-wide">
      <ExecutionTimeBar
        :elapsed-ns="response.elapsed_ns || 0"
        :elapsed-ms="response.elapsed_ms || 0"
        :error="response.error || ''"
      />
      <ResultPanel title="过滤结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { filterTest } from '../../api';
import type { FilterTestResponse } from '../../types';
import ExecutionTimeBar from '../ExecutionTimeBar.vue';
import ResultPanel from '../ResultPanel.vue';
import ScaleInput from '../ScaleInput.vue';

export default defineComponent({
  name: 'FilterTestModule',
  components: { ScaleInput, ExecutionTimeBar, ResultPanel },
  data() {
    return {
      inputMode: 'manual',
      inputText: '你这个ＢＡＤＷＯＲＤ真的不行',
      fileName: '',
      fileContent: '',
      ignoreCase: true,
      ignoreWidth: true,
      replaceRune: '*',
      scale: 1,
      loading: false,
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<FilterTestResponse>,
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
      const payloadInputText = this.inputMode === 'manual' ? this.inputText : '';
      const payloadFileName = this.inputMode === 'txt' ? this.fileName : '';
      const payloadFileContent = this.inputMode === 'txt' ? this.fileContent : '';
      this.response = await filterTest({
        input_text: payloadInputText,
        input_file_name: payloadFileName,
        input_file_content: payloadFileContent,
        ignore_case: this.ignoreCase,
        ignore_width: this.ignoreWidth,
        replace_rune: this.replaceRune,
        scale: this.scale,
      });
      this.loading = false;
    },
  },
});
</script>
