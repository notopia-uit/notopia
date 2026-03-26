import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: ['./bundled/openapi.json'],
  output: '../packages/api-gen/src/',
  plugins: [
    '@hey-api/client-next',
    '@hey-api/transformers',
    {
      name: '@hey-api/typescript',
      enums: 'javascript',
    },
    'zod',
    {
      name: '@hey-api/sdk',
      validator: true,
      transformer: true,
    },
    '@tanstack/react-query',
  ],
});
