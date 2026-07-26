import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import { Worker } from 'node:worker_threads';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';

/**
 * WorkerPool end-to-end — init-error paths: a failing config surfaces a
 * helpful error, repeated failures don't crash the host (windows
 * terminate-race canary), an async worker fault during the init-error
 * window stays contained, and the init-failure branch awaits in-flight
 * respawns before throwing (symmetric with shutdown).
 */

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    test('init failure surfaces with helpful error', async () => {
      const missingPath = path.resolve(
        __dirname,
        'fixtures',
        'missing-plugin.config.mjs',
      );
      const missingDir = path.dirname(missingPath);
      const pool = new WorkerPool({
        configs: [{ configPath: missingPath, configDirectory: missingDir }],
        workerCount: 1,
      });

      // The fixture config imports a non-existent plugin package; the
      // worker's `loadPluginsFromConfigs` throws on the failed import
      // and the WorkerPool surfaces the underlying error. We assert the
      // missing specifier appears in the message so users get a
      // pointer to the broken config entry.
      await expect(pool.init()).rejects.toThrow(
        /eslint-plugin-this-does-not-exist/,
      );
    });

    test('repeated init failures each reject without crashing the host (windows terminate-race canary)', async () => {
      // Each failure drives the init-error path: the worker self-exits
      // (`process.exitCode = 1`) and the pool now lets it exit
      // cooperatively instead of racing it with `terminate()`. Looping it
      // stresses exactly that path — on windows-latest a
      // terminate-vs-self-exit race aborts BELOW the JS layer ("Worker
      // exited unexpectedly"), so the cooperative exit must hold across
      // many reps. On macOS/Linux this is a plain regression check that the
      // helpful error still surfaces on every attempt.
      const missingPath = path.resolve(
        __dirname,
        'fixtures',
        'missing-plugin.config.mjs',
      );
      const missingDir = path.dirname(missingPath);
      for (let i = 0; i < 8; i++) {
        const pool = new WorkerPool({
          configs: [{ configPath: missingPath, configDirectory: missingDir }],
          workerCount: 1,
        });
        await expect(pool.init()).rejects.toThrow(
          /eslint-plugin-this-does-not-exist/,
        );
      }
    });

    test('init failure retains the Worker error safety net', async () => {
      // Capture the actual Worker when spawnWorker installs its init-phase
      // error listener. After init-error rejection that listener must remain:
      // otherwise a later Worker error is re-thrown by Node in the host.
      const originalOnce = Worker.prototype.once;
      let failedWorker: Worker | undefined;
      let markWorkerExited!: () => void;
      const workerExited = new Promise<void>((resolve) => {
        markWorkerExited = resolve;
      });
      Worker.prototype.once = function (event, listener) {
        if (event === 'error' && !failedWorker) {
          failedWorker = this;
          originalOnce.call(this, 'exit', () => {
            markWorkerExited();
          });
        }
        return originalOnce.call(this, event, listener);
      } as typeof Worker.prototype.once;
      let initPromise: Promise<void>;
      try {
        const cfgPath = path.resolve(
          __dirname,
          'fixtures',
          'missing-plugin.config.mjs',
        );
        const pool = new WorkerPool({
          configs: [
            { configPath: cfgPath, configDirectory: path.dirname(cfgPath) },
          ],
          workerCount: 1,
        });
        // Worker construction and init-phase listener installation are
        // synchronous up to pool.init() returning its promise. Restore the
        // shared prototype before any async wait so a stalled init cannot
        // pollute later tests after Rstest's outer timeout.
        initPromise = pool.init();
      } finally {
        Worker.prototype.once = originalOnce;
      }
      await expect(initPromise!).rejects.toThrow(
        /eslint-plugin-this-does-not-exist/,
      );
      expect(failedWorker).toBeDefined();
      expect(failedWorker!.listenerCount('error')).toBeGreaterThan(0);
      let handled = false;
      expect(() => {
        handled = failedWorker!.emit(
          'error',
          new Error('synthetic late init fault'),
        );
      }).not.toThrow();
      expect(handled).toBe(true);
      await workerExited;
    });

    test('init-failure path awaits in-flight respawns before throwing (symmetric with shutdown)', async () => {
      // Regression for Fix C. `shutdown()` awaits
      // `Promise.allSettled([...this.respawns])`; the init-failure branch
      // did not. If a ready worker crashed during the initial spawn window
      // (before `closed` flips), its exit handler registers an in-flight
      // respawn in `this.respawns`; that respawn's `.then` sees
      // `closed===true` and self-terminates the freshly-spawned worker —
      // but `init()` would `throw` BEFORE that orphan thread was reaped.
      // Fix: the init-failure branch mirrors shutdown and awaits the set.
      const pool = new WorkerPool({
        configs: [{ configPath: 'synthetic', configDirectory: 'synthetic' }],
        workerCount: 1,
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const internals = pool as any;
      internals.spawnWorker = async () => {
        throw new Error('synthetic init failure');
      };

      let releaseRespawn!: () => void;
      const respawnGate = new Promise<void>((resolve) => {
        releaseRespawn = resolve;
      });
      let markSubscribed!: () => void;
      const subscribed = new Promise<void>((resolve) => {
        markSubscribed = resolve;
      });
      const observedThenable = {
        then(
          onFulfilled: (value: void) => unknown,
          onRejected: (reason: unknown) => unknown,
        ) {
          markSubscribed();
          return respawnGate.then(onFulfilled, onRejected);
        },
      };
      internals.respawns.add(observedThenable);

      let initSettled = false;
      let initError: Error | undefined;
      const initP = pool
        .init()
        .catch((error: Error) => {
          initError = error;
        })
        .finally(() => {
          initSettled = true;
        });

      // Promise.allSettled calling the injected thenable proves init reached
      // the exact respawn-await boundary; it must remain pending until release.
      await subscribed;
      expect(initSettled).toBe(false);
      releaseRespawn();
      await initP;

      expect(initError?.message).toBe('synthetic init failure');
      // Post-conditions of the init-failure branch still hold.
      expect(internals.closed).toBe(true);
      expect(internals.workers).toHaveLength(0);
    });
  },
);
