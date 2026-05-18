import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: ['./bundled/openapi.json'],
  output: '../packages/api-gen/src/',
  plugins: [
    {
      name: '@hey-api/client-next',
      includeInEntry: true,
    },
    '@hey-api/transformers',
    {
      name: '@hey-api/typescript',
      enums: 'javascript',
      includeInEntry: true,
    },
    'zod',
    {
      name: '@hey-api/sdk',
      validator: false, // temp: because our seeded data bruh the id
      transformer: true,
    },
    {
      name: '@tanstack/react-query',
      queryKeys: true,
      useQuery: true,
      useMutation: true,
      exportFromIndex: true,
    },
  ],
});
