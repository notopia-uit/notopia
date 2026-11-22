import { fileURLToPath } from 'node:url';

import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import swc from 'unplugin-swc';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: {
    alias: {
      '#': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  plugins: [
    swc.vite({
      tsconfigFile: 'tsconfig.app.json',
      module: {
        type: 'es6',
      },
    }),
  ],
  test: {
    name: 'ai',
    environment: 'node',
    include: ['tests/**/*.spec.ts'],
    watch: false,
    reporters: [
      'default',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true'
        ? ([['github-actions', { jobSummary: { enabled: false } }]] as any)
        : []),
      new CtrfReporter({
        outputFile: 'tests-ctrf.json',
        outputDir: '../../coverage/ai',
        appName: 'ai',
      }),
    ],
    coverage: {
      enabled: true,
      reportsDirectory: '../../coverage/ai',
    },
  },
});
