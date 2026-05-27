<template>
  <div class="app-shell">
    <SidebarNav
      :modules="modules"
      :current="activeModule"
      :current-label="activeLabel"
      @switch="switchModule"
    />
    <section class="main-panel">
      <header class="topbar">
        <h1 class="topbar-title">{{ activeLabel }}</h1>
      </header>
      <main class="content">
        <component :is="activeComponent" :key="activeModule" />
      </main>
    </section>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import SidebarNav from './components/SidebarNav.vue';
import AcTestModule from './components/modules/AcTestModule.vue';
import FilterTestModule from './components/modules/FilterTestModule.vue';
import LexiconTestModule from './components/modules/LexiconTestModule.vue';
import NormalizeTestModule from './components/modules/NormalizeTestModule.vue';
import TrieTestModule from './components/modules/TrieTestModule.vue';
import { wordsPath } from './api';

export default defineComponent({
  name: 'App',
  components: {
    SidebarNav,
    FilterTestModule,
    TrieTestModule,
    AcTestModule,
    NormalizeTestModule,
    LexiconTestModule,
  },
  data() {
    return {
      activeModule: 'filter',
      loadedWordsPath: '',
      modules: [
        { key: 'filter', label: '过滤器测试', component: 'FilterTestModule' },
        { key: 'trie', label: '字典树测试', component: 'TrieTestModule' },
        { key: 'ac', label: 'AC 自动机测试', component: 'AcTestModule' },
        { key: 'normalize', label: '归一化测试', component: 'NormalizeTestModule' },
        { key: 'lexicon', label: '词库测试', component: 'LexiconTestModule' },
      ] as Array<{ key: string; label: string; component: string }>,
    };
  },
  computed: {
    activeComponent(): string {
      const current = this.modules.find((item) => item.key === this.activeModule);
      return current?.component || 'FilterTestModule';
    },
    activeLabel(): string {
      const current = this.modules.find((item) => item.key === this.activeModule);
      return current?.label || '过滤器测试';
    },
  },
  methods: {
    switchModule(next: string) {
      this.activeModule = next;
    },
    syncWindowTitle(title: string) {
      document.title = title;
      const runtime = (window as Window & {
        runtime?: { WindowSetTitle?: (value: string) => void };
      }).runtime;
      runtime?.WindowSetTitle?.(title);
    },
  },
  async mounted() {
    try {
      this.loadedWordsPath = await wordsPath();
    } catch (_error) {
      this.loadedWordsPath = '';
    }
  },
});
</script>
