import { define } from 'rstack';
import { fileURLToPath } from 'node:url';

define.lib({
  format: 'esm',
  dts: {
    bundle: true,
    tsgo: true,
    typescriptPath: fileURLToPath(
      new URL('../../tools/pinned-typescript/version.cjs', import.meta.url),
    ),
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
