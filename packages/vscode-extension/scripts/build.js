const { build } = require('esbuild');

build({
  entryPoints: ['src/extension.ts'],
  outfile: 'out/extension.js',
  format: 'cjs',
  bundle: true,
  platform: 'node',
  external: ['@rslint/core', 'vscode'],
});
