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
  process.exit(exitCode);
}

main().catch((err) => {
  process.stderr.write(`rslint: ${err}\n`);
  process.exit(1);
});
