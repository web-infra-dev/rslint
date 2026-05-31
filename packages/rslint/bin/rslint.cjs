#!/usr/bin/env node
const startTime = Date.now();
const path = require('node:path');
const { pathToFileURL } = require('node:url');
const os = require('node:os');
const fs = require('node:fs');

function getBinPath() {
  // dev / local `build:bin` output sits next to this script
  if (fs.existsSync(path.resolve(__dirname, './rslint'))) {
    return path.resolve(__dirname, './rslint');
  }
  if (fs.existsSync(path.resolve(__dirname, './rslint.exe'))) {
    return path.resolve(__dirname, './rslint.exe');
  }
  // published: the Go binary lives in the @rslint/native-{tuple} subpackage,
  // reached via its `./bin` export. npm installs only the subpackage matching
  // the host os/cpu/libc, so on linux we just try gnu then musl and use
  // whichever got installed — no libc sniffing (Go binaries are static, the
  // gnu/musl distinction doesn't matter to them).
  const arch = os.arch();
  const tuples =
    process.platform === 'linux'
      ? [`linux-${arch}-gnu`, `linux-${arch}-musl`]
      : process.platform === 'win32'
        ? [`win32-${arch}-msvc`]
        : [`${process.platform}-${arch}`];
  for (const tuple of tuples) {
    try {
      return require.resolve(`@rslint/native-${tuple}/bin`);
    } catch {}
  }
  throw new Error(
    `rslint: no native binary for ${process.platform}-${arch} ` +
      `(looked for @rslint/native-{${tuples.join(',')}})`,
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
