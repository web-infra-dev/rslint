const fs = require('node:fs');
const path = require('node:path');
const childProcess = require('node:child_process');
const { syncBuiltinESMExports } = require('node:module');
const { performance } = require('node:perf_hooks');

const stampFdRaw = process.env.RSLINT_BENCH_STAMP_FD;
const stampFd = Number.parseInt(stampFdRaw ?? '', 10);

if (Number.isNaN(stampFd)) {
  return;
}

// The CLI launches the Go binary through `spawn` (see runEngine in
// packages/rslint/src/engine.ts), so that is the call to stamp. Match on the
// binary name instead of stamping the first call so unrelated spawns the CLI
// may grow later never shift the measurement.
const GO_BINARY_NAME = process.platform === 'win32' ? 'rslint.exe' : 'rslint';

const originalSpawn = childProcess.spawn;
let stamped = false;

childProcess.spawn = function patchedSpawn(...args) {
  const file = args[0];
  if (
    !stamped &&
    typeof file === 'string' &&
    path.basename(file) === GO_BINARY_NAME
  ) {
    stamped = true;
    try {
      fs.writeSync(
        stampFd,
        JSON.stringify({
          interceptedAtMs: performance.timeOrigin + performance.now(),
          file,
        }) + '\n',
      );
    } catch {
      // Swallow pipe write errors so benchmark instrumentation never changes
      // CLI behavior.
    }
  }

  return originalSpawn.apply(this, args);
};

syncBuiltinESMExports();
