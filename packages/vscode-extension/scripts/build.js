const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');
const isWatchMode = process.argv.includes('--watch');

// The eslint-plugin host is loaded at runtime via a RELATIVE dynamic import
// `import('./eslint-plugin/index.js')` (PluginLintPool.ts). It MUST stay
// external: it is an ESM module that spawns its sibling `lint-worker.js` via
// `new Worker(new URL('./lint-worker.js', import.meta.url))`, so it has to keep
// living next to that worker inside the staged `dist/eslint-plugin/`. Bundling
// it into `dist/main.js` would break that sibling resolution. The `copy-files`
// plugin below stages the whole worker payload (worker bundle + nested
// `@rslint/native` loader/.node) into `dist/eslint-plugin/`. Unlike
// `config-loader`, which IS bundled into `main.js`, the host is never bundled.
const config = {
  entryPoints: ['src/main.ts'],
  outfile: 'dist/main.js',
  format: 'cjs',
  bundle: true,

  sourcemap: true,
  platform: 'node',
  external: ['@rslint/core', 'vscode'],
  loader: {
    '': 'file',
  },
  plugins: [
    {
      name: 'bundle-config-loader',
      setup(build) {
        // Override @rslint/core external for config-loader subpath,
        // so it gets bundled into the extension
        build.onResolve({ filter: /^@rslint\/core\/config-loader$/ }, () => ({
          path: require.resolve('@rslint/core/config-loader'),
        }));
      },
    },
    {
      name: 'external-eslint-plugin-host',
      setup(build) {
        // Keep the relative host import verbatim (emitted as a real ESM
        // `import("./eslint-plugin/index.js")`). Marking it external here —
        // rather than via the top-level `external` array — keeps the exact
        // relative specifier so it resolves against `dist/main.js` at runtime.
        build.onResolve(
          { filter: /^\.\/eslint-plugin\/index\.js$/ },
          (args) => ({
            path: args.path,
            external: true,
          }),
        );
      },
    },
    {
      name: 'copy-files',
      setup(build) {
        build.onStart(() => {
          console.info('start rebuild');
        });
        build.onEnd((result) => {
          // Don't stage a stale/partial payload on top of a failed build.
          if (result.errors.length > 0) return;
          stageRuntimeAssets();
          console.log('rebuild done');
        });
      },
    },
  ],
};

/**
 * Stage everything the packaged extension needs at runtime that esbuild does
 * not bundle into `dist/main.js`:
 *   1. the Go LSP binary + launcher (`@rslint/core/bin`)
 *   2. the eslint-plugin worker bundle (`@rslint/core/dist/eslint-plugin/`)
 *   3. a nested `@rslint/native` (a tiny fixed-name shim loader + the napi
 *      `.node`), so the worker can `import "@rslint/native"` via Node's
 *      node_modules walk-up from the worker file.
 *
 * Resolution of `@rslint/native` MUST be anchored on @rslint/core's directory:
 * the vscode-extension package doesn't declare it, so a bare `require.resolve`
 * from here throws MODULE_NOT_FOUND.
 */
function stageRuntimeAssets() {
  const distDir = path.join(__dirname, '../dist');
  const coreDir = path.dirname(require.resolve('@rslint/core/package.json'));

  // 1. Go LSP binary (`rslint`/`rslint.exe`) + launcher (`rslint.cjs`). In CI,
  //    publish-marketplace.mjs overwrites `dist/rslint` with the per-platform
  //    binary; locally this copy makes `pnpm build` self-sufficient for dev.
  fs.cpSync(path.join(coreDir, 'bin'), distDir, { recursive: true });

  // 2. eslint-plugin worker bundle — copy the WHOLE dir, never enumerate file
  //    names: the Rspack runtime chunk (`<id>.js`) has a non-deterministic id
  //    and both `index.js` and `lint-worker.js` hard-import it as a sibling.
  const epSrc = path.join(coreDir, 'dist', 'eslint-plugin');
  const epDest = path.join(distDir, 'eslint-plugin');
  fs.rmSync(epDest, { recursive: true, force: true });
  fs.cpSync(epSrc, epDest, { recursive: true });

  // 3. ESM marker. The worker bundle is ESM and `@rslint/core` is
  //    `type:module`; the copied subdir loses that ancestry, so without this
  //    marker Node would treat the `.js` files as CommonJS and fail.
  fs.writeFileSync(
    path.join(epDest, 'package.json'),
    `${JSON.stringify({ type: 'module' }, null, 2)}\n`,
  );

  // 4. Nested `@rslint/native`: a minimal shim (./native-loader-shim.cjs) that
  //    requires a FIXED-name `./rslint.node`, replacing napi-rs's auto-generated
  //    loader. We control exactly which per-platform `.node` ships in each vsix,
  //    so the loader's platform/libc/universal probing is unnecessary and — in
  //    the Electron host — a liability (see native-loader-shim.cjs for why).
  const nativeDest = path.join(epDest, 'node_modules', '@rslint', 'native');
  fs.mkdirSync(nativeDest, { recursive: true });
  fs.copyFileSync(
    path.join(__dirname, 'native-loader-shim.cjs'),
    path.join(nativeDest, 'index.js'),
  );
  fs.writeFileSync(
    path.join(nativeDest, 'package.json'),
    `${JSON.stringify(
      { name: '@rslint/native', type: 'commonjs', main: 'index.js' },
      null,
      2,
    )}\n`,
  );
  // Dev: stage the current host's `.node` as the fixed `rslint.node` so a
  // local `pnpm build` works in the dev extension host. CI overwrites this
  // single fixed-name file with the target platform's `.node`
  // (publish-marketplace.mjs) — one filename, so no wrong-arch leak is possible.
  const nativeSrcDir = path.dirname(
    require.resolve('@rslint/native', { paths: [coreDir] }),
  );
  const hostNode = fs
    .readdirSync(nativeSrcDir)
    .find((f) => /^rslint\..*\.node$/.test(f));
  if (hostNode) {
    fs.copyFileSync(
      path.join(nativeSrcDir, hostNode),
      path.join(nativeDest, 'rslint.node'),
    );
  } else {
    console.warn(
      'build.js: no local @rslint/native .node found; the eslint-plugin ' +
        'worker will not load in the dev host until `pnpm --filter ' +
        '@rslint/native build` is run.',
    );
  }
}

async function main() {
  if (isWatchMode) {
    const ctx = await esbuild.context(config);
    await ctx.watch();
  } else {
    await esbuild.build(config);
  }
}

main();
