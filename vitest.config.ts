// vitest.config.ts
import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    tsconfigPaths: true,
  },
  test: {
    watch: false,
    projects: ['apps/*', 'packages/*'],
    coverage: {
      enabled: true,
      reportsDirectory: './coverage/vitest',
    },
    passWithNoTests: true,
  },
});
