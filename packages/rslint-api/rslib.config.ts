import { defineConfig } from '@rslib/core';

export default defineConfig({
  lib: [
    {
      format: 'esm',
      dts: {
        bundle: true,
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
