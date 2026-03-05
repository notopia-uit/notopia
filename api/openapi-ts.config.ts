import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: ['./bundled/openapi.json'],
  output: '../packages/api-gen/src/',
  plugins: [
    '@hey-api/client-next',
    {
      name: '@hey-api/typescript',
      enums: 'javascript',
    },
    'zod',
    {
      name: '@hey-api/sdk',
      validator: true,
    },
    '@tanstack/react-query',
  ],
});
