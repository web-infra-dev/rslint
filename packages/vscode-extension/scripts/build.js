const esbuild = require('esbuild');
const path = require('path');
const fs = require('fs');
const isWatchMode = process.argv.includes('--watch');

const config = {
  entryPoints: ['src/main.ts'],
  outfile: 'dist/main.js',
  format: 'cjs',
  bundle: true,

  sourcemap: true,
  platform: 'node',
  // `@rslint/core/eslint-plugin` MUST stay external (esbuild matches externals
  // by exact specifier, so the bare `@rslint/core` entry does not cover this
  // subpath — it needs its own entry). Its WorkerPool spawns the lint worker
  // via `new Worker(new URL('./lint-worker.js', import.meta.url))`, so the ESM
  // `index.js` must keep living next to `lint-worker.js` inside the installed
  // `@rslint/core/dist/eslint-plugin/`. Bundling it into `dist/main.js` would
  // break that sibling resolution. Unlike `config-loader` below, do NOT add a
  // bundle override for it.
  external: ['@rslint/core', '@rslint/core/eslint-plugin', 'vscode'],
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
      name: 'copy-files',
      setup(build) {
        build.onStart(() => {
          console.info('start rebuild');
        });
        build.onEnd(() => {
          const binDir = path.resolve(
            require.resolve('@rslint/core/package.json'),
            '../bin',
          );
          fs.cpSync(binDir, path.join(__dirname, '../dist'), {
            recursive: true,
          });
          console.log('rebuild done');
        });
      },
    },
  ],
};
async function main() {
  if (isWatchMode) {
    const ctx = await esbuild.context(config);
    await ctx.watch();
  } else {
    esbuild.build(config);
  }
}

main();
