import nx from '@nx/eslint-plugin';

import baseConfig from '../../eslint.config.mjs';

export default [
  ...baseConfig,
  ...nx.configs['flat/react'],
  {
    files: ['./src/lib/auth.ts'],
    rules: {
      '@typescript-eslint/no-unsafe-assignment': 'off',
    },
  },
  {
    ignores: ['**/out-tsc'],
  },
];
