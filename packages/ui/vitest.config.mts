import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  test: { name: 'ui', environment: 'node', include: ['src/lib/**/*.spec.ts'], watch: false },
});
