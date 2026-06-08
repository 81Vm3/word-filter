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
      <label class="field">
        <span class="field-label">结果展示模式</span>
        <select class="text-input" v-model="displayMode">
          <option value="replace">屏蔽替换</option>
          <option value="highlight">高亮标记</option>
        </select>
      </label>
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
      <p class="hint" v-if="displayMode === 'highlight'">
        高亮模式下不展示替换文本，但仍会把替换字符参数传给后端。
      </p>
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
      <section class="visual-result">
        <header class="visual-result-header">
          <h3>可视化结果</h3>
          <span class="visual-mode-tag">{{ displayModeLabel }}</span>
        </header>
        <p class="hint" v-if="!executedText">暂无可视化结果，请先执行一次测试。</p>
        <div class="visual-text" v-else-if="displayMode === 'replace'">{{ replaceViewText }}</div>
        <div class="visual-text visual-text-highlight" v-else>
          <template v-for="(segment, idx) in highlightSegments" :key="idx">
            <mark v-if="segment.highlight" class="hit-mark">{{ segment.text }}</mark>
            <span v-else>{{ segment.text }}</span>
          </template>
        </div>
      </section>
      <ResultPanel title="过滤结果" :payload="response.result" />
    </article>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { filterTest } from '../../api';
import type { FilterHit, FilterTestResponse } from '../../types';
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
      displayMode: 'replace',
      ignoreCase: true,
      ignoreWidth: true,
      replaceRune: '*',
      scale: 1,
      loading: false,
      executedText: '',
      response: {
        elapsed_ns: 0,
        elapsed_ms: 0,
        error: '',
        result: {},
      } as Partial<FilterTestResponse>,
    };
  },
  computed: {
    displayModeLabel(): string {
      return this.displayMode === 'highlight' ? '高亮标记模式' : '屏蔽替换模式';
    },
    resultData(): FilterTestResponse['result'] {
      return (this.response.result ?? {}) as FilterTestResponse['result'];
    },
    replaceViewText(): string {
      const text = this.resultData?.sanitized;
      return typeof text === 'string' ? text : '';
    },
    mergedHighlightRanges(): Array<{ start: number; end: number }> {
      const rawHits = this.resultData?.hits;
      if (!Array.isArray(rawHits) || rawHits.length === 0 || this.executedText === '') {
        return [];
      }
      const runeLen = Array.from(this.executedText).length;
      const ranges: Array<{ start: number; end: number }> = [];
      for (const hit of rawHits as FilterHit[]) {
        const start = Math.max(0, Math.min(runeLen, Number(hit.rune_start)));
        const end = Math.max(0, Math.min(runeLen, Number(hit.rune_end)));
        if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) {
          continue;
        }
        ranges.push({ start, end });
      }
      if (ranges.length === 0) {
        return [];
      }
      ranges.sort((a, b) => {
        if (a.start !== b.start) {
          return a.start - b.start;
        }
        return a.end - b.end;
      });
      const merged: Array<{ start: number; end: number }> = [];
      let current = { ...ranges[0] };
      for (let i = 1; i < ranges.length; i++) {
        const next = ranges[i];
        if (next.start <= current.end) {
          current.end = Math.max(current.end, next.end);
          continue;
        }
        merged.push(current);
        current = { ...next };
      }
      merged.push(current);
      return merged;
    },
    highlightSegments(): Array<{ text: string; highlight: boolean }> {
      if (this.executedText === '') {
        return [];
      }
      const runes = Array.from(this.executedText);
      if (this.mergedHighlightRanges.length === 0) {
        return [{ text: this.executedText, highlight: false }];
      }
      const segments: Array<{ text: string; highlight: boolean }> = [];
      let cursor = 0;
      for (const range of this.mergedHighlightRanges) {
        if (range.start > cursor) {
          segments.push({
            text: runes.slice(cursor, range.start).join(''),
            highlight: false,
          });
        }
        segments.push({
          text: runes.slice(range.start, range.end).join(''),
          highlight: true,
        });
        cursor = range.end;
      }
      if (cursor < runes.length) {
        segments.push({
          text: runes.slice(cursor).join(''),
          highlight: false,
        });
      }
      return segments.filter((seg) => seg.text !== '');
    },
  },
  methods: {
    normalizeScale(scale: number): number {
      if (scale < 1) {
        return 1;
      }
      if (scale > 10000) {
        return 10000;
      }
      return scale;
    },
    buildScaledText(input: string, scale: number): string {
      const normalizedScale = this.normalizeScale(scale);
      if (normalizedScale === 1 || input === '') {
        return input;
      }
      const parts = new Array<string>(normalizedScale);
      for (let i = 0; i < normalizedScale; i++) {
        parts[i] = input;
      }
      return parts.join('\n');
    },
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
      const sourceText = payloadFileContent !== '' ? payloadFileContent : payloadInputText;
      this.executedText = this.buildScaledText(sourceText, this.scale);
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
