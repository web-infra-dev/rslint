#!/usr/bin/env node
const path = require('node:path');
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
function main() {
  const binPath = getBinPath();
  try {
    require('child_process').execFileSync(binPath, process.argv.slice(2), {
      stdio: 'inherit',
    });
  } catch (error) {
    // Preserve the exit code from the child process
    if (error.status != null) {
      process.exit(error.status);
    } else {
      console.error(`Failed to execute ${binPath}: ${error.message}`);
      process.exit(1);
    }
  }
}
main();
