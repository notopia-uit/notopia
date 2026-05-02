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
      },
    },
  ],
  settings: {
    'better-tailwindcss': {
      entryPoint: './app/globals.css',
    },
  },
});
