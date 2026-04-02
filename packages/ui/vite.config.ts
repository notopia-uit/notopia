import { resolve } from 'path';
import { defineConfig } from 'vite';
import dts from 'vite-plugin-dts';

export default defineConfig({
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [
    dts({
      tsconfigPath: resolve(__dirname, 'tsconfig.lib.json'),
    }),
  ],
  build: {
    lib: {
      entry: resolve(__dirname, 'src/index.ts'),
      name: 'NotopiaUi',
      formats: ['es'],
      fileName: 'index',
    },
    rollupOptions: {
      external: [
        'react',
        'react-dom',
        'react/jsx-runtime',
        ...Object.keys(require('./package.json').peerDependencies || {}),
      ],
      output: {
        entryFileNames: '[name].mjs',
        chunkFileNames: '[name].mjs',
      },
    },
  },
});
