import { defineConfig } from '@rslib/core';

export default defineConfig({
  lib: [
    {
      format: 'esm',
      // Avoid using api-extractor to bundle d.ts as it
      // pulls in the typescript-go submodule sources and
      // triggers extractor internal errors. Generate plain
      // declarations instead.
      dts: {
        bundle: false,
      },
    },
  ],
  source: {
    tsconfigPath: './tsconfig.build.json',
  },
  resolve: {
    alias: {
      '@typescript/libsyncrpc': false,
    },
  },
  tools: {
    rspack(config) {
      if (!config.resolve?.conditionNames) {
        config.resolve.conditionNames = ['...'];
      }
      config.resolve.conditionNames.unshift('@typescript/source');
      return config;
    },
  },
});
