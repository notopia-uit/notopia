import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: ['./note/dist/openapi.json', './edit/dist/openapi.json'],
  output: '../packages/notopia-api/src/',
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
