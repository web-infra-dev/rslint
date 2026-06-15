#!/usr/bin/env node
const startTime = Date.now();
const path = require('node:path');
const { pathToFileURL } = require('node:url');
const os = require('node:os');

function getBinPath() {
  // The Go binary lives in the @rslint/native-{tuple} platform package, reached
  // via its `./bin` export. Resolution is identical in dev and prod: `pnpm build`
  // drops the host binary into npm/rslint/{tuple}/, and npm installs only the
  // subpackage matching the host os/cpu/libc. On linux we just try gnu then musl
  // and use whichever resolved — no libc sniffing (Go binaries are static, the
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
  // process.exit() would tear down before async-buffered stdout writes (pipes,
  // Windows TTYs) flush, truncating the lint tail. Setting exitCode lets the
  // event loop drain naturally; run()'s cleanup guarantees nothing keeps it
  // alive.
  process.exitCode = exitCode;
}

main().catch((err) => {
  process.stderr.write(`rslint: ${err}\n`);
  process.exitCode = 1;
});
