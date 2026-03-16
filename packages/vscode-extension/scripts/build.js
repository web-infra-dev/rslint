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
