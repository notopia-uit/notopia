import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: ['../bundled/share.json'],
  output: '../../packages/api-share-gen/src/',
  plugins: [
    '@hey-api/transformers',
    {
      name: '@hey-api/typescript',
      enums: 'javascript',
    },
  ],
});
