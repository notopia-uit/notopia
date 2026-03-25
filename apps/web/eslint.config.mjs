import baseConfig from '../../eslint.config.mjs';
import nextEslintPluginNext from '@next/eslint-plugin-next';
import nx from '@nx/eslint-plugin';
import eslintPluginBetterTailwindcss from 'eslint-plugin-better-tailwindcss';
import reactHooks from 'eslint-plugin-react-hooks';
import { defineConfig } from 'eslint/config';
import { parser as eslintParserTypeScript } from 'typescript-eslint';

export default defineConfig(
  ...baseConfig,
  reactHooks.configs.flat.recommended,
  {
    plugins: {
      '@next/next': nextEslintPluginNext,
    },
  },
  ...nx.configs['flat/react-typescript'],
  {
    extends: [eslintPluginBetterTailwindcss.configs.recommended],
    settings: {
      'better-tailwindcss': {
        entryPoint: './src/app/global.css',
      },
    },
  },
  {
    files: ['**/*.{ts,tsx,cts,mts}'],
    languageOptions: {
      parser: eslintParserTypeScript,
      parserOptions: {
        project: true,
      },
    },
  },
  {
    files: ['**/*.{jsx,tsx}'],
    languageOptions: {
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
  },
  {
    ignores: [
      '.next/**/*',
      '**/out-tsc',
      './next.config.ts',
      './tailwind.config.ts',
    ],
  }
);
