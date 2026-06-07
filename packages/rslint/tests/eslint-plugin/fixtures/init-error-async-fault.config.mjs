// Fixture for the init-error error-safety-net regression test. Two
// things happen when the worker imports this config:
//
//   1. A timer is scheduled, then a top-level `throw` makes the worker's
//      `loadPluginsFromConfigs` fail → the worker takes its
//      `sendInitError` path and `pool.init()` rejects.
//   2. The scheduled timer fires AFTER init-error (the pending timer keeps
//      the worker's event loop alive past the failed eval), throwing
//      uncaught inside the worker → the worker emits an 'error' event.
//
// The WorkerPool must keep an 'error' listener through the init-error
// window: an unhandled 'error' on a Worker is re-thrown by Node as an
// uncaught exception in the host. See worker-pool-e2e-init-errors.test.ts.
setTimeout(() => {
  throw new Error('async worker fault after init-error');
}, 20);

throw new Error('intentional init eval failure');
