import { resolve } from 'path';

import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss';
import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['jsx-a11y', 'react', 'react-perf', 'nextjs'],
  env: {
    browser: true,
    node: true,
  },
  overrides: [
    {
      files: ['**/*.{ts,tsx}'],
      jsPlugins: ['eslint-plugin-better-tailwindcss'],
      rules: {
        ...eslintPluginBetterTailwindcss.configs.recommended.rules,
        'better-tailwindcss/no-unknown-classes': [
          'error',
          {
            ignore: ['notopia-reference', 'notopia-tag', 'global-graph-outer'],
          },
        ],
        'better-tailwindcss/enforce-consistent-line-wrapping': 'off',
        'better-tailwindcss/enforce-consistent-class-order': 'off',
        'better-tailwindcss/enforce-consistent-variant-order': 'off',
      },
    },
  ],
  ignorePatterns: ['**/shadcn/**'],
  settings: {
    'better-tailwindcss': {
      entryPoint: resolve(import.meta.dirname, 'src/globals.css'),
      tsconfig: resolve(import.meta.dirname, 'tsconfig.lib.json'),
      cwd: 'packages/ui',
      detectComponentClasses: true, // may oxlint doesnt work
    },
  },
});
