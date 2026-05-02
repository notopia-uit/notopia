// https://vitepress.dev/guide/custom-theme
import type { Theme } from 'vitepress';
import { useData } from 'vitepress';
import { createMermaidRenderer } from 'vitepress-mermaid-renderer';
import { theme } from 'vitepress-openapi/client';

import 'vitepress-openapi/dist/style.css';
import DefaultTheme from 'vitepress/theme';
import { h, nextTick, watch } from 'vue';

// sort-imports-ignore
import '@catppuccin/vitepress/theme/mocha/lavender.css';

export default {
  extends: DefaultTheme,
  Layout: () => {
    const { isDark } = useData();
    const initMermaid = () => {
      const mermaidRenderer = createMermaidRenderer({
        theme: isDark.value ? 'dark' : 'null',
        startOnLoad: false,
        sequence: {
          useMaxWidth: true,
        },
      });
      mermaidRenderer.setToolbar({
        showLanguageLabel: false,
      });
    };
    nextTick(initMermaid);
    watch(() => isDark.value, initMermaid);

    return h(DefaultTheme.Layout, null, {
      // https://vitepress.dev/guide/extending-default-theme#layout-slots
    });
  },
  async enhanceApp(ctx) {
    await theme.enhanceApp(ctx);
  },
} satisfies Theme;
