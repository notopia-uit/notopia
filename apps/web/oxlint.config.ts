import { resolve } from 'path';

import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss';
import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['jsx-a11y', 'nextjs', 'react', 'react-perf'],
  env: {
    browser: true,
    node: true,
    builtin: true,
  },
  overrides: [
    {
      files: ['**/*.{ts,tsx}'],
      jsPlugins: ['eslint-plugin-better-tailwindcss'],
      rules: {
        ...eslintPluginBetterTailwindcss.configs.recommended.rules,
        'better-tailwindcss/enforce-consistent-line-wrapping': 'off',
      },
    },
  ],
  settings: {
    'better-tailwindcss': {
      entryPoint: resolve(import.meta.dirname, 'app/globals.css'),
      tsconfig: resolve(import.meta.dirname, 'tsconfig.json'),
      cwd: 'apps/web',
    },
  },
});
