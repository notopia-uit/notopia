// vitest.config.ts
import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    tsconfigPaths: true,
  },
  test: {
    watch: false,
    projects: ['apps/document'],
    reporters: [
      'default',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true' ? ['github-actions'] : []),
      new CtrfReporter({
        outputFile: 'ctrf.json',
        outputDir: './coverage/vitest',
        appName: 'vitest',
      }),
    ],
    outputFile: {
      junit: './coverage/vitest/junit.xml',
    },
    coverage: {
      enabled: true,
      reportsDirectory: './coverage/vitest',
    },
    passWithNoTests: true,
  },
});
