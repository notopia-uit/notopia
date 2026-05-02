import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss';
import { defineConfig } from 'oxlint';

import baseConfig from '../../oxlint-base.config.mts';

export default defineConfig({
  extends: [baseConfig],
  plugins: ['jsx-a11y', 'react', 'react-perf'],
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
            ignore: [
              'animate-in',
              'fade-in',
              'font-heading',
              'notopia-reference',
              'notopia-tag',
              'slide-in-from-bottom-4',
              'tree-container',
            ],
          },
        ],
        'better-tailwindcss/enforce-consistent-line-wrapping': 'off',
        'better-tailwindcss/enforce-consistent-class-order': 'off',
      },
    },
  ],
  ignorePatterns: ['**/shadcn/**'],
  settings: {
    'better-tailwindcss': {
      entryPoint: `${import.meta.dirname}/src/globals.css`, // Because it is ran by nx from workspace root
    },
  },
});
