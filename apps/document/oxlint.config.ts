import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['node', 'vitest'],
  ignorePatterns: ['dist/**', 'coverage/**', 'out-tsc/**'],
  env: {
    node: true,
  },
  rules: {
    'oxc/no-barrel-file': 'error',
  },
});
