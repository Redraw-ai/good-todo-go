import { defineConfig } from 'orval';

export default defineConfig({
  publicApi: {
    input: {
      target: '../backend/openapi/openapi-public.yaml',
    },
    output: {
      mode: 'tags-split',
      target: './src/api/public',
      schemas: './src/api/public/model',
      client: 'react-query',
      httpClient: 'axios',
      clean: true,
      prettier: true,
      override: {
        mutator: {
          path: './src/api/axios-instance.ts',
          name: 'customInstance',
        },
      },
    },
  },
});
