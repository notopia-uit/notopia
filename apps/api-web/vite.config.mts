import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig(({ command, isPreview }) => ({
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
  base: `${process.env.DOCS_BASE || '/notopia/'}api/${command === 'build' || isPreview ? 'openapi-scalar/' : ''}`,
  build: {
    outDir: './dist',
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
    rollupOptions: {
      output: {
        chunkFileNames: 'assets/c-[name]-[hash].js',
        assetFileNames: 'assets/c-[hash][extname]',
      },
    },
  },
}));
