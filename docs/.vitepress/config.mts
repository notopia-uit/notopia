import { type ChildProcess, spawn } from 'node:child_process';
import { statSync } from 'node:fs';
import path from 'node:path';

import spec from '@notopia-uit/api/openapi' with { type: 'json' };
import getPort from 'get-port';
import { DefaultTheme, UserConfig, defineConfig } from 'vitepress';
import { type Plugin } from 'vitepress';
import { generateLlmsFiles, llmsDevServer } from 'vitepress-openapi';
import {
  BuildTimeDiagramPluginOptions,
  DiagramPluginOptions,
  createBuildTimeDiagramsPlugin,
} from 'vitepress-plugin-diagrams';
import llmstxt from 'vitepress-plugin-llms';
import { pagefindPlugin } from 'vitepress-plugin-pagefind';
import { tabsMarkdownPlugin } from 'vitepress-plugin-tabs';
import { withSidebar } from 'vitepress-sidebar';
import { VitePressSidebarOptions } from 'vitepress-sidebar/types';
import waitOn from 'wait-on';

import { restLlmsOptions } from './rest-llms';
import { getRestReference } from './rest-reference';
import { copyScalarOutput } from './scalar-output';
import { validateAuthorship } from './theme/authorship';

const COMPOSE_KROKI_URL = 'http://localhost:8002';
const PUBLIC_KROKI_URL = 'https://kroki.io';
const KROKI_PROBE_TIMEOUT = 3_000;
const KROKI_START_TIMEOUT = 30_000;
const WIN32_EXECUTABLE_EXTENSIONS = ['.exe', '.cmd', '.bat', '.com'];

function isFile(candidate: string): boolean {
  try {
    return statSync(candidate).isFile();
  } catch {
    return false;
  }
}

function resolveOnPath(bin: string): string | null {
  const dirs = (process.env.PATH ?? '').split(path.delimiter).filter(Boolean);
  const names =
    process.platform === 'win32'
      ? [bin, ...WIN32_EXECUTABLE_EXTENSIONS.map((ext) => `${bin}${ext}`)]
      : [bin];

  for (const dir of dirs) {
    for (const name of names) {
      const candidate = path.join(dir, name);
      if (isFile(candidate)) return candidate;
    }
  }
  return null;
}

const configuredKrokiUrl =
  process.env.DOCS_KROKI_URL?.trim() || process.env.KROKI_SERVER_URL?.trim() || undefined;

const krokiBin = configuredKrokiUrl ? null : resolveOnPath('kroki');

const krokiPort = krokiBin ? await getPort({ port: 8000 }) : undefined;

function krokiHealthResource(url: string): string {
  return `${url.replace(/^https?/, (scheme) => `${scheme}-get`)}/health`;
}

async function isKrokiHealthy(url: string, timeout = KROKI_PROBE_TIMEOUT): Promise<boolean> {
  try {
    await waitOn({ resources: [krokiHealthResource(url)], timeout });
    return true;
  } catch {
    return false;
  }
}

async function resolveKrokiServerUrl(): Promise<string | undefined> {
  if (configuredKrokiUrl) return configuredKrokiUrl;

  if (krokiPort) return `http://localhost:${krokiPort}`;

  if (await isKrokiHealthy(COMPOSE_KROKI_URL)) return COMPOSE_KROKI_URL;

  console.warn(
    `⚠️  No local Kroki at ${COMPOSE_KROKI_URL}, falling back to ${PUBLIC_KROKI_URL}. ` +
      'Diagram sources leave your machine and rendering is rate limited; run ' +
      '`docker compose --profile docs up -d kroki` or set DOCS_KROKI_URL to use a Kroki you control.'
  );
  return undefined;
}

const krokiServerUrl = await resolveKrokiServerUrl();

const base = process.env.DOCS_BASE || '/notopia/';

const diagramPluginOptions = {
  diagramsDir: 'src/public/diagrams',
  publicPath: `${base}diagrams`,
  diagramsDistDir: 'diagrams',
  excludedDiagramTypes: ['mermaid'],
  ...(krokiServerUrl ? { krokiServerUrl } : {}),
} satisfies DiagramPluginOptions & BuildTimeDiagramPluginOptions;

const diagrams = createBuildTimeDiagramsPlugin(diagramPluginOptions);

function waitForKroki(url: string): Plugin {
  const health = krokiHealthResource(url);

  return {
    name: 'vitepress-diagrams-kroki-wait',
    apply: 'build',

    async buildStart() {
      try {
        await waitOn({ resources: [health], timeout: KROKI_START_TIMEOUT });
      } catch {
        throw new Error(
          `Kroki is not reachable at ${url}, configured via DOCS_KROKI_URL / KROKI_SERVER_URL. ` +
            'Start that server, point the variable at a running one, or unset it to fall back to ' +
            `${PUBLIC_KROKI_URL}.`
        );
      }
    },
  };
}

type KrokiWrapperOptions = {
  bin: string;
  port: number;
};

export function createDiagramsWithKroki({ bin, port }: KrokiWrapperOptions): Plugin {
  const krokiUrl = `http://localhost:${port}`;
  let krokiProcess: ChildProcess | null = null;
  let started = false;

  async function startKroki() {
    if (started) return;
    started = true;

    const child = spawn(bin, {
      stdio: ['ignore', 'ignore', 'inherit'],
      env: {
        ...process.env,
        KROKI_PORT: String(port),
        DEBUG: 'true', // for d2
      },
    });
    krokiProcess = child;

    const exitedEarly = new Promise<never>((_, reject) => {
      child.once('error', reject);
      child.once('exit', (code, signal) =>
        reject(new Error(`Kroki (${bin}) exited with ${code ?? signal} before becoming healthy`))
      );
    });

    await Promise.race([
      waitOn({ resources: [krokiHealthResource(krokiUrl)], timeout: KROKI_START_TIMEOUT }),
      exitedEarly,
    ]);

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

function krokiPlugins(): Plugin[] {
  if (krokiBin && krokiPort) return [createDiagramsWithKroki({ bin: krokiBin, port: krokiPort })];
  if (configuredKrokiUrl) return [waitForKroki(configuredKrokiUrl)];
  return [];
}

// https://vitepress.dev/reference/site-config
const vitePressOptions = {
  title: 'Notopia',
  description: 'Utopia of Notes',
  lang: 'en-GB',
  base,
  srcDir: 'src',
  async buildEnd(siteConfig) {
    await copyScalarOutput(
      path.resolve(import.meta.dirname, '../../apps/api-web/dist'),
      siteConfig.outDir
    );
    const { outDir, pages, site } = siteConfig;
    await generateLlmsFiles({ outDir, pages, site }, restLlmsOptions);
  },
  transformPageData(pageData) {
    validateAuthorship(pageData.frontmatter.authorship, pageData.relativePath);
    if (pageData.relativePath.startsWith('api/rest/operations/') && pageData.params?.title) {
      pageData.title = pageData.params.title;
    }
  },
  markdown: {
    theme: {
      light: 'catppuccin-latte',
      dark: 'catppuccin-mocha',
    },
    config: (md) => {
      diagrams.configureMarkdown(md);
      md.use(tabsMarkdownPlugin);
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
        link: '/api/openapi-scalar/index.html',
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
    // Pagefind reads Markdown Git history during configuration; CI keeps historical trees.
    plugins: [
      ...krokiPlugins(),
      pagefindPlugin(),
      diagrams.vitePlugin(),
      llmsDevServer(restLlmsOptions),
      llmstxt({
        ignoreFiles: ['api/rest/operations/**'],
        ignoreFilesPerOutput: {
          pages: ['api/rest.md'],
          llmsFullTxt: ['api/rest.md'],
        },
      }),
    ],
  },
  ignoreDeadLinks: [`${base}api/openapi-scalar/index.html`],
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
  excludeByGlobPattern: ['[[]operationId].md'],
} satisfies VitePressSidebarOptions;

const config = withSidebar(vitePressOptions, vitePressSidebarOptions);
const restSidebar = getRestReference(spec).groups.map(({ tag, operations }) => ({
  text: tag,
  collapsed: true,
  items: operations.map(({ title, href }) => ({ text: title, link: href })),
}));

export default defineConfig({
  ...config,
  themeConfig: {
    ...config.themeConfig,
    sidebar: {
      '/': config.themeConfig?.sidebar as DefaultTheme.SidebarItem[],
      '/api/rest.md': restSidebar,
      '/api/rest/': restSidebar,
    },
  },
});
