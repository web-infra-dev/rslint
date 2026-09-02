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
 * 1. Library surface → `dist/` (`tsconfig.lib.json`, which inherits root's
 *    exclude of `src/eslint-plugin/**`). A dts build is a TS project, so it must
 *    not share its `tsBuildInfoFile` with the tsgo `typecheck` over the same
 *    `src` — the two tools' incremental formats clash. Hence a tsconfig per
 *    consumer: `tsconfig.lib.json` (here), `tsconfig.worker.json` (below), and
 *    `tsconfig.build.json` (typecheck). `output.autoExternal` externalizes
 *    `dependencies` (`picomatch`) + `peerDependencies` (`jiti`); `tinyglobby` is
 *    a devDep, so its code bundles in. The public `globals` catalog is also a
 *    devDep, but a Rspack plugin emits one package-internal JSON asset per
 *    environment so the root can load only selected maps. Consumers install
 *    neither devDependency.
 *    `tinyglobby`'s `fdir` loads `picomatch` via `createRequire`,
 *    which rspack can't statically follow — so `picomatch` can't be bundled away
 *    and stays a runtime dep. One `lib` block with all entries: the surface
 *    modules share a graph, so shared chunks between entries are fine here.
 *
 * 2. eslint-plugin worker → `dist/eslint-plugin/` (`tsconfig.worker.json`,
 *    which includes `src/eslint-plugin/**`). Each entry is its own `lib` block
 *    so Rspack inlines each output's full module graph with NO shared chunks —
 *    crucial for the worker (`new Worker(...)` spawns a fresh V8 isolate that
 *    pays a filesystem-open + parse cost per extra chunk; modules can't be
 *    reused across isolates). The ESLint-compat libs (`@typescript-eslint/
 *    scope-manager`, `eslint-scope`, `esquery`) are devDependencies imported
 *    statically so they bundle in; consumers need none at runtime. The native
 *    parser loader (`src/eslint-plugin/native/load-binding.ts`) bundles in too,
 *    but the platform `.node` it loads stays external: the loader selects the
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
    ],
  };
});

define.test({
  // Keep the standalone Rstest behavior from before the configs were merged.
  extends: {},
  testEnvironment: 'node',
  globals: true,
  // Each eslint-plugin test process can spawn up to eight Worker threads.
  // Run Windows test files serially so Rstest's CPU-count process pool cannot
  // multiply that inner concurrency and starve Worker startup/shutdown.
  pool: process.platform === 'win32' ? { maxWorkers: 1 } : undefined,
  // Normal completion is event-driven. This is only the final in-process
  // deadlock sentinel, deliberately later than the 30-minute child watchdogs.
  // The same semantic boundary applies on every platform.
  testTimeout: 35 * 60_000,
  // The eslint-plugin worker tests spawn the built `dist/eslint-plugin/
  // lint-worker.js` (worker_threads can't run TS); this setup file points the
  // pool at it via setWorkerEntryForTests(). Run `pnpm build` once before testing.
  setupFiles: ['./tests/eslint-plugin/setup-worker-entry.ts'],
  // this is temporary folder generated by go build
  exclude: ['pkg/mod', 'node_modules'],
});
