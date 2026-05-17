import { nxViteTsPaths } from '@nx/vite/plugins/nx-tsconfig-paths.plugin';
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
  plugins: [nxViteTsPaths(), react()],
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
        chunkFileNames: 'assets/c-[name]-[hash].js',
        assetFileNames: 'assets/c-[hash][extname]',
      },
    },
  },
}));
