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
// platform package carrying the napi `.node`) into `dist/eslint-plugin/`. Unlike
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
 *   3. a nested `@rslint/native-<tuple>` platform package (the napi `.node` +
 *      a minimal package.json), which the worker's loader resolves at runtime
 *      via Node's node_modules walk-up from the worker file.
 *
 * Resolution MUST be anchored on @rslint/core's directory: the vscode-extension
 * package doesn't declare the platform packages, so a bare `require.resolve`
 * from here throws MODULE_NOT_FOUND.
 */
function stageRuntimeAssets() {
  const distDir = path.join(__dirname, '../dist');
  const coreDir = path.dirname(require.resolve('@rslint/core/package.json'));

  // 1. Launcher (`rslint.js`) from core/bin + the Go LSP binary. build:bin now
  //    lands the Go binary in the host platform package (not core/bin), so copy
  //    it from there. In CI, publish-marketplace.mjs overwrites dist/rslint with
  //    the target-platform binary; locally this makes `pnpm build` dev-ready.
  fs.cpSync(path.join(coreDir, 'bin'), distDir, { recursive: true });
  const goBin = process.platform === 'win32' ? 'rslint.exe' : 'rslint';
  try {
    const goSrc = require.resolve(`@rslint/native-${hostTuple()}/bin`, {
      paths: [coreDir],
    });
    fs.copyFileSync(goSrc, path.join(distDir, goBin));
  } catch {
    console.warn(
      'build.js: no local Go LSP binary in the host platform package; the LSP ' +
        'will not start in the dev host until `pnpm --filter @rslint/core build:bin` is run.',
    );
  }

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

  // 4. Nested platform package `@rslint/native-<tuple>`: the worker's native
  //    loader (bundled into lint-worker.js from core sources) resolves it at
  //    runtime via createRequire walk-up to load the `.node`. The vsix is
  //    per-platform; CI (publish-marketplace.mjs) re-stages this as the TARGET
  //    platform's package. Dev stages the host's so a local `pnpm build` works
  //    in the dev extension host.
  const tuple = hostTuple();
  const pkgBase = `native-${tuple}`;
  const nodeFile = `rslint.${tuple}.node`;
  const nativeDest = path.join(epDest, 'node_modules', '@rslint', pkgBase);
  fs.mkdirSync(nativeDest, { recursive: true });
  let hostNodePath = null;
  try {
    const srcPkgDir = path.dirname(
      require.resolve(`@rslint/${pkgBase}/package.json`, { paths: [coreDir] }),
    );
    const candidate = path.join(srcPkgDir, nodeFile);
    if (fs.existsSync(candidate)) hostNodePath = candidate;
  } catch {
    // platform package not resolvable for this host — fall through to the warning
  }
  if (hostNodePath) {
    fs.copyFileSync(hostNodePath, path.join(nativeDest, nodeFile));
    fs.writeFileSync(
      path.join(nativeDest, 'package.json'),
      `${JSON.stringify(
        { name: `@rslint/${pkgBase}`, exports: { '.': `./${nodeFile}` } },
        null,
        2,
      )}\n`,
    );
  } else {
    console.warn(
      `build.js: no local ${nodeFile} found; the eslint-plugin worker will not ` +
        'load in the dev host until `pnpm --filter @rslint/native build` is run.',
    );
  }
}

// The platform-package tuple the worker's loader computes on THIS host. Electron
// embeds glibc, so linux is always gnu here (the loader's musl probe never
// selects musl under an Electron host).
function hostTuple() {
  const { platform, arch } = process;
  if (platform === 'win32') return `win32-${arch}-msvc`;
  if (platform === 'linux') return `linux-${arch}-gnu`;
  return `${platform}-${arch}`;
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
