import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
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
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true' ? ['github-actions'] : []),
      new CtrfReporter({
        outputFile: 'tests-ctrf.json',
        outputDir: '../../coverage/document',
        appName: 'document',
      }),
    ],
    coverage: {
      enabled: true,
      reportsDirectory: '../../coverage/document',
    },
    passWithNoTests: true,
  },
}));
