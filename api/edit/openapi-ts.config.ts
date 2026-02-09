import { defineConfig } from '@hey-api/openapi-ts';

export default defineConfig({
  input: './dist/openapi.json',
  output: '../../packages/notopia-api/src/edit/',
});
