import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [react()],
  test: {
    name: 'web',
    watch: false,
    environment: 'node',
    include: ['app/**/*.test.ts'],
    passWithNoTests: true,
    reporters: [
      'default',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true'
        ? ([['github-actions', { jobSummary: { enabled: false } }]] as any)
        : []),
      new CtrfReporter({
        outputFile: 'tests-ctrf.json',
        outputDir: '../../coverage/web',
        appName: 'web',
      }),
    ],
    coverage: {
      enabled: true,
      reportsDirectory: '../../coverage/web',
    },
  },
});
