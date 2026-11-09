import { defineConfig } from 'oxlint';

export default defineConfig({
  options: {
    typeAware: process.env.IS_EDITOR === 'true' ? false : true,
  },
  plugins: ['typescript', 'unicorn', 'eslint'],
  rules: {
    'typescript/require-await': 'error',
    'typescript/return-await': 'error',
    'typescript/no-deprecated': 'warn',
  },
  settings: {
    'better-tailwindcss': {
      detectComponentClasses: true,
    },
  },
});
