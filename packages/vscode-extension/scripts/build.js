const { build } = require('esbuild');

build({
  entryPoints: ['src/extension.ts'],
  outfile: 'out/extension.js',
  format: 'cjs',
  bundle: true,
  sourcemap: true,
  platform: 'node',
  external: ['@rslint/core', 'vscode'],
  tsconfig: 'tsconfig.build.json',
});
