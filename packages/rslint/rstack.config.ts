import { define } from 'rstack';
import type { Rspack } from 'rstack/lib';
import {
  bundleGlobalsTypesPlugin,
  emitGlobalsAssetsPlugin,
} from './plugins/globals.ts';
import { generateRuleOptionTypesPlugin } from './plugins/generate-rule-option-types.ts';

/**
 * Single rslib build for all of `@rslint/core`'s JS: the public library surface
 * plus the internal `eslint-plugin` worker runtime. Replaces the former split
 * (tsgo `build:js` + rslib `build:worker`) — one `build:js` now emits both.
 *
 * Two groups of `lib` blocks:
 *
 * 1. Library surface → `dist/` (`tsconfig.lib.json`, which excludes the plugin
 *    runtime and JavaScript language modules from file discovery). Its
 *    declarations use a separate build info file from the tsgo check in
 *    `tsconfig.build.json` because the cache formats differ.
 *    `output.autoExternal` externalizes the
 *    optional `jiti` peer. The `globals` catalog is a devDep; a Rspack plugin
 *    emits one package-internal JSON asset per environment so the root can
 *    load only selected maps.
 *    Consumers do not install `globals`. One `lib` block with all entries:
 *    the surface modules share a graph, so shared chunks between entries are
 *    fine here.
 *
 * 2. Plugin host and worker → `dist/eslint-plugin/`. The public plugin entries
 *    bundle declarations from `tsconfig.plugin-api.json`. The worker checks
 *    its full project and compiler syntax reference with `dts.build`, keeping
 *    those private declarations in the build cache. Each runtime entry has
 *    its own `lib` block (the public index shares its block with a type-only entry)
 *    so Rspack inlines each output's full module graph with NO shared chunks —
 *    crucial for the worker (`new Worker(...)` spawns a fresh V8 isolate that
 *    pays a filesystem-open + parse cost per extra chunk; modules can't be
 *    reused across isolates). The ESLint-compat libs (`@typescript-eslint/
 *    scope-manager`, `eslint-scope`, `esquery`) are devDependencies imported
 *    statically so they bundle in; consumers need none at runtime. The native byte
 *    transport loader (`src/native/binding.ts`) bundles into
 *    the worker and full public index, but not the lightweight host. The
 *    platform `.node` it loads stays external: the loader selects the
 *    `@rslint/native-<tuple>` package at runtime via `createRequire`, which
 *    rspack can't statically follow (so the binary is never inlined — intended).
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
        service: './src/service/service.ts',
        internal: './src/internal/node.ts',
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
    dts: { bundle: true },
    plugins: [bundleGlobalsTypesPlugin(), generateRuleOptionTypesPlugin()],
  };

  const workerBase = {
    format: 'esm' as const,
    bundle: true,
    output: {
      autoExternal: true,
      target: 'node' as const,
      distPath: { root: './dist/eslint-plugin' },
    },
    source: {
      tsconfigPath: './tsconfig.plugin-api.json',
    },
    tools: {
      rspack(config: Rspack.Configuration) {
        config.resolve ??= {};
        config.resolve.conditionNames = [
          '@typescript/source',
          ...(config.resolve.conditionNames ?? ['...']),
        ];
      },
    },
  };

  return {
    lib: [
      librarySurface,
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          entry: {
            index: './src/eslint-plugin/index.ts',
            types: './src/eslint-plugin/types.ts',
          },
        },
        dts: { bundle: true },
      },
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          tsconfigPath: './tsconfig.worker.json',
          entry: { 'lint-worker': './src/eslint-plugin/lint-worker.ts' },
        },
        dts: { build: true },
      },
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          entry: { host: './src/eslint-plugin/host.ts' },
        },
        // Keep this host bundle limited to worker coordination. The CLI loads
        // its native memory arena separately through the common binding loader.
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
  exclude: ['pkg/mod', 'node_modules'],
});
