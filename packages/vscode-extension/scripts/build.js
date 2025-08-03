const esbuild = require('esbuild');
const path = require('node:path');
const os = require('node:os');
const fs = require('node:fs');
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
      name: 'copy-files',
      setup(build) {
        build.onStart(() => {
          console.log('📁 Copy file start');
        });
        build.onEnd(() => {
          const binPath = require.resolve(
            `@rslint/${os.platform()}-${os.arch()}/bin`,
          );
          const binaryName = path.basename(binPath);
          fs.cpSync(binPath, path.join(__dirname, `../dist/${binaryName}`), {
            recursive: true,
          });
          console.log('✅ Copy file done');
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
