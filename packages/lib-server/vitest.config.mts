import CtrfReporter from '@d2t/vitest-ctrf-json-reporter';
import swc from 'unplugin-swc';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [
    swc.vite({
      tsconfigFile: 'tsconfig.spec.json',
      module: {
        type: 'es6',
      },
    }) as any, // because unplugin broken on new vitest. Until it fix
  ],
  test: {
    name: 'lib-server',
    watch: false,
    globals: true,
    environment: 'node',
    include: ['src/**/*.spec.ts'],
    reporters: [
      'default',
      'junit',
      ...(process.env.GITHUB_ACTIONS === 'true' ? ['github-actions'] : []),
      new CtrfReporter({
        outputFile: 'tests-ctrf.json',
        outputDir: '../../coverage/lib-server',
        appName: 'lib-server',
      }),
    ],
    coverage: {
      enabled: true,
      reportsDirectory: '../../coverage/lib-server',
    },
    passWithNoTests: true,
  },
});
