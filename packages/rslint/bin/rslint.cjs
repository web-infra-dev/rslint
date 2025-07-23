#!/usr/bin/env node
const path = require('node:path');
const os = require('node:os');
function getBinPath() {
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

