import { defineConfig, type RsbuildPlugin } from '@rslib/core';
import { generateRuleOptionTypes } from '../../scripts/generate-rule-option-types.mjs';

/**
 * Splices typed rule options into `dist/index.d.ts` once the whole
 * (multi-`lib`-entry) build finishes — `onAfterBuild` fires once for the
 * entire rslib build, not per entry, so by the time it runs `librarySurface`'s
 * `dist/index.d.ts` already exists. See scripts/generate-rule-option-types.mjs.
 *
 * A missing rule-schemas dump only skips with a warning, rather than failing
 * the build: unlike `build:bin`-driven builds (local dev, CI's Go-toolchain
 * jobs) and CI jobs that fetch the prebuilt artifact, `build:website`'s
 * `build:js` (no Go, no artifact fetch — it only needs the JS runtime, not
 * typed rule options) never produces or fetches that file.
 */
const generateRuleOptionTypesPlugin = (): RsbuildPlugin => ({
  name: 'generate-rule-option-types',
  setup(api) {
    api.onAfterBuild(async () => {
      try {
        const count = await generateRuleOptionTypes();
        api.logger.log(
          `generate-rule-option-types: injected ${count} typed rule(s) into dist/index.d.ts`,
        );
      } catch (err) {
        if (err.code === 'RULE_SCHEMAS_NOT_FOUND') {
          api.logger.warn(
            `generate-rule-option-types: skipped — ${err.message}`,
          );
          return;
        }
        throw err;
      }
    });
  },
});

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
 *    `tsconfig.build.json` (typecheck). `autoExternal` externalizes `dependencies`
 *    (`picomatch`) + `peerDependencies` (`jiti`); `tinyglobby` is a devDep so it
 *    bundles in. But `tinyglobby`'s `fdir` loads `picomatch` via `createRequire`,
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
const librarySurface = {
  format: 'esm' as const,
  bundle: true,
  autoExternal: true,
  output: {
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
  dts: { bundle: true },
  plugins: [generateRuleOptionTypesPlugin()],
};

const workerBase = {
  format: 'esm' as const,
  bundle: true,
  autoExternal: true,
  output: {
    target: 'node' as const,
    distPath: { root: './dist/eslint-plugin' },
  },
  source: {
    tsconfigPath: './tsconfig.worker.json',
  },
};

export default defineConfig({
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
});
