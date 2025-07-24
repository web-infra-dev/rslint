#!/usr/bin/env node
const path = require('node:path');
const os = require('node:os');
const fs = require('node:fs');
function getBinPath() {
  if (fs.existsSync(path.resolve(__dirname, './rslint'))) {
    return path.resolve(__dirname, './rslint');
  }
  let platformKey = `${process.platform}-${os.arch()}`;
  return path.resolve(__dirname, `./rslint-${platformKey}`);
}
function main() {
  const binPath = getBinPath();
  require('child_process').execFileSync(binPath, process.argv.slice(2), {
    stdio: 'inherit',
  });
}
main();
