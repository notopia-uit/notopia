/// <reference types='vitest' />
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig(() => ({
  resolve: {
    tsconfigPaths: true,
  },
  root: import.meta.dirname,
  server: {
    port: 9080,
    host: true,
  },
  preview: {
    port: 9080,
    host: true,
  },
  plugins: [react()],
  // Uncomment this if you are using workers.
  // worker: {
  //  plugins: [],
  // },
  base: '/notopia/api/',
  build: {
    outDir: './dist',
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
    rollupOptions: {
      output: {
        chunkFileNames: (chunkInfo) => {
          const name = chunkInfo.name.replace(/^_/, '');
          return `assets/${name}-[hash].js`;
        },
      },
    },
  },
}));
