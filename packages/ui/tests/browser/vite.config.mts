import { createRequire } from 'node:module';

import { defineConfig } from 'vite';

const require = createRequire(new URL('../../package.json', import.meta.url));

export default defineConfig({
  root: new URL('../../../..', import.meta.url).pathname,
  resolve: {
    alias: [
      '@blocknote/core/yjs',
      '@blocknote/core',
      '@blocknote/xl-ai',
      '@ai-sdk/react',
      'ai',
      'yjs',
      'y-prosemirror',
    ].map((name) => ({
      find: new RegExp(`^${name}$`),
      replacement:
        name === 'y-prosemirror'
          ? require.resolve(name).replace(/dist\/y-prosemirror\.cjs$/, 'src/y-prosemirror.js')
          : require.resolve(name).replace(/\.cjs$/, name === 'yjs' ? '.mjs' : '.js'),
    })),
  },
  server: { host: '127.0.0.1', port: 5188, strictPort: true },
});
