// https://vitepress.dev/guide/custom-theme
import type { Theme } from 'vitepress';
import { useData } from 'vitepress';
import { createMermaidRenderer } from 'vitepress-mermaid-renderer';
import { theme } from 'vitepress-openapi/client';
import { enhanceAppWithTabs } from 'vitepress-plugin-tabs/client';

import 'vitepress-openapi/dist/style.css';
import DefaultTheme from 'vitepress/theme';
import { h, nextTick, watch } from 'vue';

import Authorship from './Authorship.vue';

// sort-imports-ignore
import '@catppuccin/vitepress/theme/mocha/lavender.css';

export default {
  extends: DefaultTheme,
  Layout: () => {
    const { isDark } = useData();
    const initMermaid = () => {
      const mermaidRenderer = createMermaidRenderer({
        theme: isDark.value ? 'redux-dark-color' : 'redux-color',
        startOnLoad: false,
        sequence: {
          showSequenceNumbers: true,
          useMaxWidth: true,
        },
      });
      mermaidRenderer.setToolbar({
        showLanguageLabel: false,
        downloadFormat: 'svg',
        desktop: {
          download: 'enabled',
        },
        mobile: {
          download: 'enabled',
        },
      });
    };
    nextTick(initMermaid);
    watch(() => isDark.value, initMermaid);

    return h(DefaultTheme.Layout, null, {
      'doc-top': () => h(Authorship),
    });
  },
  async enhanceApp(ctx) {
    await theme.enhanceApp(ctx);
    enhanceAppWithTabs(ctx.app);
  },
} satisfies Theme;
