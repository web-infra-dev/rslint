#!/usr/bin/env node
const startTime = Date.now();
const path = require('node:path');
const { pathToFileURL } = require('node:url');
const os = require('node:os');
const fs = require('node:fs');

function getBinPath() {
  if (fs.existsSync(path.resolve(__dirname, './rslint'))) {
    return path.resolve(__dirname, './rslint');
  }
  if (fs.existsSync(path.resolve(__dirname, './rslint.exe'))) {
    return path.resolve(__dirname, './rslint.exe');
  }
  let platformKey = `${process.platform}-${os.arch()}`;

  return require.resolve(
    `@rslint/${platformKey}/rslint${process.platform === 'win32' ? '.exe' : ''}`,
  );
}

async function main() {
  const binPath = getBinPath();
  const { run } = await import(
    pathToFileURL(path.resolve(__dirname, '../dist/cli.js')).href
  );
  const exitCode = await run(binPath, process.argv.slice(2), startTime);
  // Do NOT call process.exit(): a piped `process.stdout` is async, and a
  // hard exit discards Node's still-unflushed buffer, silently truncating
  // large lint output (`| jq`, `| tee`, CI capturing stdout). Setting
  // `exitCode` lets the event loop drain stdout (pending writes keep the
  // loop alive until the bytes are handed to the OS), then the process
  // exits naturally. This is sound only because runEngine returns after the
  // Go child has exited, the worker pool is terminated, and IPC is closed
  // (engine.ts §5-6) — i.e. no leaked handle keeps the loop alive past the
  // flush. (Verified: a `--start-time`-only run exits promptly.)
  process.exitCode = exitCode;
}

main().catch((err) => {
  process.stderr.write(`rslint: ${err}\n`);
  process.exit(1);
});
