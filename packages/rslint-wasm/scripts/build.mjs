import esbuild from 'esbuild';
import fs from 'fs/promises';
import path from 'path';
import {polyfillNode} from 'esbuild-plugin-polyfill-node';
import { fileURLToPath } from 'url';

async function main() {
  await buildWorker();
  await buildBrowser();
}
async function buildWorker() {
 await esbuild.build({
    entryPoints: ['./src/worker.ts'],
    bundle: true,
    outfile: './dist/worker.js',
    platform: 'browser',
    target: 'es2020',
    write:true,
    format: 'iife',
    sourcemap: 'inline',
    plugins: [ polyfillNode()]
  });
}
async function buildBrowser() {
  const WEB_WORKER_SOURCE_CODE = await fs.readFile(path.resolve(import.meta.dirname, '../dist/worker.js'), 'utf8');
  await esbuild.build({
    entryPoints: ['./src/browser.ts'],
    outfile: './dist/index.mjs',
    bundle: true,
    platform: 'browser',
    target: 'es2020',
    format: 'esm',
    splitting: false,
    define: {
      WEB_WORKER_SOURCE_CODE: JSON.stringify(WEB_WORKER_SOURCE_CODE),
    },
  });
}

main();