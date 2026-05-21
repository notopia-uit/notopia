import { nxViteTsPaths } from '@nx/vite/plugins/nx-tsconfig-paths.plugin';
import swc from 'unplugin-swc';
import { defineProject } from 'vitest/config';

export default defineProject({
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
    globals: true,
    environment: 'node',
    include: ['{src,tests,database}/**/*.{test,spec}.{ts,tsx}'],
  },
});
