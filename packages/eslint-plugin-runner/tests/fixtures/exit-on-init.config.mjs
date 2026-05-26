// Fixture for the spawn-phase `exit` listener test.
//
// This config module hard-exits the worker at import time — simulating
// a plugin whose module-top-level code calls `process.exit()` (or that
// segfaults / OOMs) DURING init, before the worker can send a `ready`
// or `init-error` message. The worker emits only an `'exit'` event;
// without a spawn-phase exit listener the pool would hang on the
// 60 s init timer. With the fix, `init()` rejects promptly with the
// exit code.
process.exit(42);

// Unreachable, but keeps the module shape valid for tooling.
export default [];
