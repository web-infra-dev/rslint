import { defineConfig } from '@rslib/core';

/**
 * Build configuration for `@rslint/eslint-plugin-runner`.
 *
 * Each library is its own `lib` block so Rspack treats them as
 * independent builds — that means each output file gets the entire
 * reachable module graph inlined and there are no `dist/<hash>.js`
 * shared chunks between entries. Crucial for the worker entry, which
 * is loaded via `new Worker(workerEntryPath)`: the worker spawns a
 * fresh V8 isolate and pays a filesystem-open + parse cost for every
 * additional chunk it has to load. Splitting `lint-worker.js` into a
 * thin 10 KB stub plus a 158 KB shared chunk doubled the worker's
 * boot-time disk reads for no shared-cache benefit (each worker is
 * its own isolate; modules can't be reused across them).
 *
 * `oxc-parser` is the only runtime dep that MUST stay external — it
 * uses a NAPI loader (`src-js/bindings.js`) that resolves its
 * platform-specific `.node` binary relative to `import.meta.url`;
 * bundling its loader would repoint that URL at our bundle file and
 * the native binary lookup would fail.
 *
 * Five outputs map 1:1 to the package.json `exports` field:
 *
 *   - `index.js`        — main entry, host integration
 *   - `lint-worker.js`  — worker_threads entry (resolved by sibling
 *                          lookup inside `worker-pool.ts`, so the file
 *                          MUST land at `dist/lint-worker.js`)
 *   - `ipc-client.js`   — Go-side stdio framing helpers
 *   - `types.js`        — shared type re-exports
 */
const baseLib = {
  format: 'esm' as const,
  bundle: true,
  // All three categories external — runtime deps come from the user's
  // node_modules (correct version resolution + tree-shareable), peerDeps
  // are explicitly host-supplied. Only devDependencies bundle into dist.
  //
  // Two source-level idioms tacitly rely on this:
  //   - `@typescript-eslint/scope-manager` / `visitor-keys` are loaded
  //     via `createRequire(import.meta.url)` (CJS packages on an ESM
  //     entry). Static `import` would tempt a future contributor and
  //     `autoExternal: { dependencies: false }` would silently bundle
  //     them.
  //   - `oxc-parser` is ESM with a NAPI loader that resolves its
  //     platform `.node` next to `import.meta.url`; bundling repoints
  //     that URL at our dist file and breaks the native binary lookup.
  autoExternal: true,
  output: {
    target: 'node' as const,
  },
  source: {
    tsconfigPath: './tsconfig.build.json',
  },
};

export default defineConfig({
  lib: [
    {
      ...baseLib,
      source: { ...baseLib.source, entry: { index: './src/index.ts' } },
      // Bundle dts only on the main entry — the other entries either
      // re-export from `index` or are tiny standalone modules; emitting
      // independent dts for each would duplicate the same types five
      // times. The main `dist/index.d.ts` covers all public surface.
      dts: { bundle: true },
    },
    {
      ...baseLib,
      source: {
        ...baseLib.source,
        entry: { 'lint-worker': './src/lint-worker.ts' },
      },
      dts: false,
    },
    {
      ...baseLib,
      source: {
        ...baseLib.source,
        entry: { 'ipc-client': './src/ipc-client.ts' },
      },
      dts: { bundle: true },
    },
    {
      ...baseLib,
      source: { ...baseLib.source, entry: { types: './src/types.ts' } },
      dts: { bundle: true },
    },
  ],
});
