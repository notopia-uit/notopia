import { DefaultTheme, UserConfig, defineConfig } from 'vitepress';
import {
  DiagramPluginOptions,
  configureDiagramsPlugin,
} from 'vitepress-plugin-diagrams';
import { pagefindPlugin } from 'vitepress-plugin-pagefind';
import { withSidebar } from 'vitepress-sidebar';
import { VitePressSidebarOptions } from 'vitepress-sidebar/types';

const diagramPluginOptions = {
  diagramsDir: 'src/public/diagrams',
  publicPath: '/notopia/diagrams',
  excludedDiagramTypes: ['mermaid'],
  krokiServerUrl: process.env.CI ? undefined : process.env.KROKI_SERVER_URL,
} satisfies DiagramPluginOptions;

// https://vitepress.dev/reference/site-config
const vitePressOptions = {
  title: 'Notopia',
  description: 'Utopia of Notes',
  lang: 'en-GB',
  base: '/notopia/',
  srcDir: 'src',
  markdown: {
    theme: {
      light: 'catppuccin-latte',
      dark: 'catppuccin-mocha',
    },
    config: (md) => {
      configureDiagramsPlugin(md, diagramPluginOptions);
    },
  },
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      {
        text: 'Home',
        link: '../',
        target: '_self',
        rel: 'noopener',
      },
      {
        text: 'Scalar API',
        link: '/api/index.html',
        target: '_blank',
        rel: 'noopener',
      },
    ],

    socialLinks: [
      {
        icon: 'github',
        link: 'https://github.com/notopia-uit/notopia',
      },
    ],
  },
  vite: {
    plugins: [pagefindPlugin()],
  },
  ignoreDeadLinks: ['/notopia/api/index.html'],
} satisfies UserConfig<NoInfer<DefaultTheme.Config>>;

const vitePressSidebarOptions = {
  documentRootPath: 'src',
  useTitleFromFileHeading: true,
  useTitleFromFrontmatter: true,
  useFolderLinkFromIndexFile: true,
  useFolderTitleFromIndexFile: true,
  sortMenusByFrontmatterOrder: true,
  collapsed: true,
  collapseDepth: 2,
} satisfies VitePressSidebarOptions;

export default defineConfig(
  withSidebar(vitePressOptions, vitePressSidebarOptions)
);
