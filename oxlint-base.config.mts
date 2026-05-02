import { defineConfig } from 'oxlint';

export default defineConfig({
  plugins: ['import', 'typescript', 'unicorn', 'eslint'],
  rules: {
    'import/no-cycle': ['error', { maxDepth: 3 }],
  },
  options: { typeAware: true },
});
