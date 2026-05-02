import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['react', 'react-perf'],
  env: {
    browser: true,
  },
});
