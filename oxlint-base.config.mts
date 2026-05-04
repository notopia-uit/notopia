import { defineConfig } from 'oxlint';

export default defineConfig({
  plugins: ['typescript', 'unicorn', 'eslint'],
  options: {
    typeAware: true,
  },
  settings: {
    'better-tailwindcss': {
      detectComponentClasses: true,
    },
  },
});
