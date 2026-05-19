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
    reporters: [
      'default',
      'vitest-ctrf-json-reporter',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true' ? ['github-actions'] : []),
    ],
    outputFile: {
      junit: 'test-report.junit.xml',
    },
    coverage: {
      enabled: true,
    },
    passWithNoTests: true,
  },
}));
