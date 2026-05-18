import { nxViteTsPaths } from '@nx/vite/plugins/nx-tsconfig-paths.plugin';
import swc from 'unplugin-swc';
import { defineConfig } from 'vitest/config';

export default defineConfig(() => ({
  root: __dirname,
  plugins: [
    nxViteTsPaths(),
    swc.vite({
      module: {
        type: 'es6',
      },
    }),
  ],
  test: {
    name: 'document',
    watch: false,
    globals: true,
    environment: 'node',
    include: ['{src,tests,database}/**/*.{test,spec}.{ts,tsx}'],
    coverage: {
      enabled: true,
    },
    passWithNoTests: true,
  },
}));
