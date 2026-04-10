import nx from '@nx/eslint-plugin';
import { defineConfig } from 'eslint/config';

import baseConfig from '../../eslint.config.mjs';

export default defineConfig(...baseConfig, ...nx.configs['flat/react'], {
  files: ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
});
