import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  test: {
    name: 'ui',
    environment: 'node',
    include: ['src/lib/**/*.spec.ts'],
    watch: false,
    reporters: [
      'default',
      'junit',
      // oxlint-disable-next-line node/no-process-env
      ...(process.env.GITHUB_ACTIONS === 'true'
        ? ([['github-actions', { jobSummary: { enabled: false } }]] as any)
        : []),
      new CtrfReporter({
        outputFile: 'tests-ctrf.json',
        outputDir: '../../coverage/ui',
        appName: 'ui',
      }),
    ],
    coverage: {
      enabled: true,
      reportsDirectory: '../../coverage/ui',
    },
  },
});
