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
 *    runtime and JavaScript language modules from file discovery). A dts build
 *    is a TS project, so it must
 *    not share its `tsBuildInfoFile` with the tsgo `typecheck` over the same
 *    `src` — the two tools' incremental formats clash. Hence a tsconfig per
 *    consumer: `tsconfig.lib.json` (here), `tsconfig.worker.json` (below), and
 *    `tsconfig.build.json` (typecheck). `output.autoExternal` externalizes the
 *    optional `jiti` peer. The `globals` catalog is a devDep; a Rspack plugin
 *    emits one package-internal JSON asset per environment so the root can
 *    load only selected maps.
 *    Consumers do not install `globals`. One `lib` block with all entries:
 *    the surface modules share a graph, so shared chunks between entries are
 *    fine here.
 *
 * 2. Plugin host and worker → `dist/eslint-plugin/` (`tsconfig.worker.json`,
 *    which includes `src/eslint-plugin/**` and `src/languages/js/**`). Each
 *    entry is its own `lib` block
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
      tsconfigPath: './tsconfig.worker.json',
    },
  };

  return {
    lib: [
      librarySurface,
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          entry: { index: './src/eslint-plugin/index.ts' },
        },
        // Bundle dts only on the main entry — the others re-export from `index`
        // or are tiny standalone modules; per-entry dts would duplicate types.
        dts: { bundle: true },
      },
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          entry: { 'lint-worker': './src/eslint-plugin/lint-worker.ts' },
        },
        dts: false,
      },
      {
        ...workerBase,
        source: {
          ...workerBase.source,
          entry: { types: './src/eslint-plugin/types.ts' },
        },
        dts: { bundle: true },
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
