/// <reference types='vitest' />
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig(() => ({
  resolve: {
    tsconfigPaths: true,
  },
  root: import.meta.dirname,
  cacheDir: 'node_modules/.vite/apps/test-editor',
  server: {
    port: 4201,
    host: true,
  },
  preview: {
    port: 4201,
    host: true,
  },
  plugins: [tailwindcss(), react()],
  build: {
    outDir: './dist',
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
  },
}));
