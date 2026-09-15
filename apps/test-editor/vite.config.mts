import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig(() => ({
  resolve: {
    tsconfigPaths: true,
  },
  root: import.meta.dirname,
  server: {
    port: 4201,
    host: true,
  },
  preview: {
    port: 4201,
    host: true,
  },
  plugins: [tailwindcss() as any, react()], // tailwindcss vite doesn't conform vite v:
  build: {
    outDir: './dist',
    emptyOutDir: true,
    reportCompressedSize: true,
    commonjsOptions: {
      transformMixedEsModules: true,
    },
  },
}));
