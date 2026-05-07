import { defineConfig } from 'oxlint';

export default defineConfig({
  plugins: ['typescript', 'unicorn', 'eslint'],
  options: {
    typeAware: true,
  },
  rules: {
    'typescript/require-await': 'error',
    'typescript/return-await': 'error',
  },
  settings: {
    'better-tailwindcss': {
      detectComponentClasses: true,
    },
  },
});
