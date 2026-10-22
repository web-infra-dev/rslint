import { define } from 'rstack';

define.lib({
  format: 'esm',
  dts: {
    bundle: true,
    // Declaration generation uses the pinned JavaScript TypeScript compiler.
    tsgo: false,
  },
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
