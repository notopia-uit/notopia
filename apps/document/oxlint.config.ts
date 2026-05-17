import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['node'],
  ignorePatterns: ['dist/**', 'coverage/**', 'out-tsc/**'],
  env: {
    node: true,
  },
});
