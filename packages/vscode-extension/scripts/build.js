const esbuild = require('esbuild');
const fs = require('fs');
const path = require('path');
const isWatchMode = process.argv.includes('--watch');
const distDirectory = path.resolve(__dirname, '../dist');

const config = {
  entryPoints: ['src/main.ts'],
  outfile: 'dist/main.js',
  format: 'cjs',
  bundle: true,

  sourcemap: true,
  platform: 'node',
  // Core imports in extension source are type-only. Runtime modules are loaded
  // by absolute path from the user's project and must never enter this bundle.
  external: ['@rslint/core', 'vscode'],
};

async function main() {
  // A universal VSIX must never retain native assets from an older build.
  fs.rmSync(distDirectory, { recursive: true, force: true });
  if (isWatchMode) {
    const ctx = await esbuild.context(config);
    await ctx.watch();
  } else {
    await esbuild.build(config);
  }
}

main();
