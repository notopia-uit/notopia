import { ChildProcess, spawn } from 'node:child_process';

import getPort from 'get-port';
import { DefaultTheme, UserConfig, defineConfig } from 'vitepress';
import { type Plugin } from 'vitepress';
import {
  BuildTimeDiagramPluginOptions,
  DiagramPluginOptions,
  configureDiagramsPlugin,
  createBuildTimeDiagramsPlugin,
} from 'vitepress-plugin-diagrams';
import { pagefindPlugin } from 'vitepress-plugin-pagefind';
import { withSidebar } from 'vitepress-sidebar';
import { VitePressSidebarOptions } from 'vitepress-sidebar/types';
import waitOn from 'wait-on';

const krokiPort = await getPort({ port: 8000 });

const diagramPluginOptions = {
  diagramsDir: 'src/public/diagrams',
  publicPath: '/notopia/diagrams',
  excludedDiagramTypes: ['mermaid'],
  krokiServerUrl: `http://localhost:${krokiPort}`,
} satisfies DiagramPluginOptions & BuildTimeDiagramPluginOptions;

type KrokiWrapperOptions = {
  port?: number;
  docker?: boolean;
};

export function createDiagramsWithKroki(options: KrokiWrapperOptions = {}): Plugin {
  let krokiProcess: ChildProcess | null = null;
  let krokiUrl: string;
  let started = false;

  async function startKroki() {
    if (started) return;
    started = true;
    krokiUrl = `http://localhost:${options.port}`;
    const healthWaitOn = `http-get://localhost:${options.port}/health`;

    if (options.docker) {
      krokiProcess = spawn(
        'docker',
        [
          'run',
          '--rm',
          '-e',
          'DEBUG=true', // for d2
          '-p',
          `${options.port}:8000`,
          'yuzutech/kroki',
        ],
        {
          stdio: ['ignore', 'ignore', 'inherit'],
        }
      );
    } else {
      krokiProcess = spawn('kroki', {
        stdio: ['ignore', 'ignore', 'inherit'],
        env: {
          ...process.env,
          KROKI_PORT: String(options.port),
          DEBUG: 'true', // for d2
        },
      });
    }

    await waitOn({
      resources: [healthWaitOn],
      timeout: 10000,
      verbose: true,
    });

    console.log(`🟢 Kroki started at ${krokiUrl}`);
  }

  function stopKroki() {
    if (krokiProcess) {
      krokiProcess.kill();
      krokiProcess = null;
      console.log('🔴 Kroki stopped');
    }
  }

  return {
    name: 'vitepress-diagrams-kroki',

    async configureServer(server) {
      await startKroki();
      server.httpServer?.once('close', stopKroki);
    },

    async buildStart() {
      await startKroki();
    },

    closeBundle() {
      stopKroki();
    },
  };
}

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
    plugins: [
      createDiagramsWithKroki({ port: krokiPort }),
      pagefindPlugin(),
      createBuildTimeDiagramsPlugin(diagramPluginOptions),
    ],
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

export default defineConfig(withSidebar(vitePressOptions, vitePressSidebarOptions));
