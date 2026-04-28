const fs = require('node:fs');
const childProcess = require('node:child_process');
const { syncBuiltinESMExports } = require('node:module');
const { performance } = require('node:perf_hooks');

const stampFdRaw = process.env.RSLINT_BENCH_STAMP_FD;
const stampFd = Number.parseInt(stampFdRaw ?? '', 10);

if (Number.isNaN(stampFd)) {
  return;
}

const originalExecFileSync = childProcess.execFileSync;
let stamped = false;

childProcess.execFileSync = function patchedExecFileSync(file, args, options) {
  if (!stamped) {
    stamped = true;
    try {
      fs.writeSync(
        stampFd,
        JSON.stringify({
          interceptedAtMs: performance.timeOrigin + performance.now(),
          file: typeof file === 'string' ? file : null,
        }) + '\n',
      );
    } catch {
      // Swallow pipe write errors so benchmark instrumentation never changes
      // CLI behavior.
    }
  }

  return originalExecFileSync.call(this, file, args, options);
};

syncBuiltinESMExports();
