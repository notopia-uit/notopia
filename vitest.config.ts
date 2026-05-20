// vitest.config.ts
import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    projects: ['packages/*', 'apps/*'],
    reporters: [
      'default',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true' ? ['github-actions'] : []),
      new CtrfReporter({
        outputFile: 'vitest-ctrf.json',
        outputDir: './coverage',
        appName: 'vitest',
      }),
    ],
    outputFile: {
      junit: './coverage/vitest-junit.xml',
    },
    coverage: {
      enabled: true,
    },
    passWithNoTests: true,
  },
});
