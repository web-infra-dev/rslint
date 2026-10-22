#!/usr/bin/env node
/** Build native npm artifacts exclusively through the Rust addon pipeline. */
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { main as buildRust } from './tsrs/build-rust.mjs';

export function main(args = process.argv.slice(2)) {
  const [product, ...options] = args;
  if (product !== 'rslint') {
    throw new Error(
      product === 'tsgo' || product === 'tsgo-server'
        ? 'The standalone semantic server is not supported by the Rust+JS rsbuild migration. No Go server or compiler-service artifact is available.'
        : 'Usage: node scripts/build-npm.mjs rslint [--platform TUPLE] [--debug] [--plan]',
    );
  }
  // The default is the current host. Release matrices must select their target
  // explicitly; no unrelated platform artifacts are reused or synthesized.
  buildRust(options);
}
if (
  process.argv[1] &&
  pathToFileURL(path.resolve(process.argv[1])).href === import.meta.url
) {
  try {
    main();
  } catch (error) {
    process.stderr.write(
      `Native npm build failed: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    process.exitCode = 1;
  }
}
