import { fileURLToPath } from 'node:url';

import swc from 'unplugin-swc';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: { alias: { '#': fileURLToPath(new URL('./src', import.meta.url)) } },
  plugins: [swc.vite({ tsconfigFile: 'tsconfig.app.json', module: { type: 'es6' } }) as any],
  test: { name: 'ai', environment: 'node', include: ['tests/**/*.spec.ts'], watch: false },
});
