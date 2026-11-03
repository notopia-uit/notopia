import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  root: import.meta.dirname,
  resolve: { tsconfigPaths: true },
  plugins: [react()],
  test: {
    name: 'web',
    watch: false,
    environment: 'node',
    include: ['app/**/*.test.ts'],
  },
});
