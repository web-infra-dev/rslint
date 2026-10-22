import { define } from 'rstack';
import type { Rspack } from 'rstack/lib';
import {
  bundleGlobalsTypesPlugin,
  emitGlobalsAssetsPlugin,
} from './plugins/globals.ts';
import { generateRuleOptionTypesPlugin } from './plugins/generate-rule-option-types.ts';

/** Build the direct Rust CLI/API and explicit compatibility rejections.
 * Legacy service and worker entrypoints expose the same unsupported shim as
 * package exports. They never bundle process launchers or a compiler service.
 */
define.lib(() => {
  const librarySurface = {
    format: 'esm' as const,
    bundle: true,
    output: {
      autoExternal: true,
      target: 'node' as const,
      distPath: { root: './dist' },
    },
    source: {
      tsconfigPath: './tsconfig.lib.json',
      entry: {
        index: './src/index.ts',
        'config-loader': './src/config/config-loader.ts',
        cli: './src/cli/cli.ts',
      },
    },
    tools: {
      rspack(config: Rspack.Configuration) {
        config.plugins ??= [];
        config.plugins.push(emitGlobalsAssetsPlugin());
      },
    },
    dts: { bundle: true, tsgo: false },
    plugins: [bundleGlobalsTypesPlugin(), generateRuleOptionTypesPlugin()],
  };

  return {
    lib: [
      librarySurface,
      {
        format: 'esm',
        bundle: true,
        output: {
          autoExternal: true,
          target: 'node',
          distPath: { root: './dist' },
        },
        source: {
          entry: {
            service: './unsupported.mjs',
            internal: './unsupported.mjs',
          },
        },
        // Public subpaths use the checked-in unsupported.d.ts declarations.
        dts: false,
      },
    ],
  };
});

define.test({
  extends: {},
  testEnvironment: 'node',
  globals: true,
  pool: process.platform === 'win32' ? { maxWorkers: 1 } : undefined,
  testTimeout: 35 * 60_000,
  exclude: ['node_modules'],
});
