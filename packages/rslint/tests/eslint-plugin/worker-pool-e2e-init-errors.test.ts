import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';

/**
 * WorkerPool end-to-end — init-error paths: a failing config surfaces a
 * helpful error, repeated failures don't crash the host (windows
 * terminate-race canary), an async worker fault during the init-error
 * window stays contained, and the init-failure branch awaits in-flight
 * respawns before throwing (symmetric with shutdown).
 */

// Skipped on windows: tearing down a worker that has oxc (a napi addon)
// loaded aborts below the JS layer there (nodejs/node#34567) and crashes
// the rstest worker running this file. These e2e tests spawn real
// workers and tear them down, so they are windows-skipped; they still
// run on linux/macOS.
describe.skipIf(process.platform === 'win32')(
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
        workerInitTimeoutMs: 5000,
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
          workerInitTimeoutMs: 5000,
        });
        await expect(pool.init()).rejects.toThrow(
          /eslint-plugin-this-does-not-exist/,
        );
      }
    });

    test('init failure + async worker fault keeps the host alive (error safety net)', async () => {
      // The fixture fails init AND schedules an async throw that fires after
      // sendInitError, making the worker emit an 'error' event. An unhandled
      // Worker 'error' is re-thrown by Node as an uncaught exception in the
      // host — so the pool must keep an 'error' listener through the
      // init-error window. Capture host uncaughtException to assert none leaks.
      const faults: string[] = [];
      const onUncaught = (e: Error) => {
        faults.push(e.message);
      };
      process.on('uncaughtException', onUncaught);
      try {
        const cfgPath = path.resolve(
          __dirname,
          'fixtures',
          'init-error-async-fault.config.mjs',
        );
        const pool = new WorkerPool({
          configs: [
            { configPath: cfgPath, configDirectory: path.dirname(cfgPath) },
          ],
          workerCount: 1,
          workerInitTimeoutMs: 5000,
        });
        await expect(pool.init()).rejects.toThrow(/init eval failure/);
        // Let the worker's scheduled async fault (20ms) fire + propagate.
        await new Promise((r) => setTimeout(r, 200));
        expect(faults).toEqual([]);
      } finally {
        process.off('uncaughtException', onUncaught);
      }
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
      const missingPath = path.resolve(
        __dirname,
        'fixtures',
        'missing-plugin.config.mjs',
      );
      const pool = new WorkerPool({
        configs: [
          {
            configPath: missingPath,
            configDirectory: path.dirname(missingPath),
          },
        ],
        workerCount: 1,
        workerInitTimeoutMs: 5_000,
      });

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const internals = pool as any;

      // Inject an in-flight respawn that NEVER settles on its own — only
      // we resolve it. This makes the assertion fully deterministic with
      // no reliance on timer ordering: with the fix, init() blocks at
      // `await Promise.allSettled([...this.respawns])` and cannot reject
      // until we resolve this; without the fix, init() rejects the moment
      // the (missing-plugin) spawn fails, never awaiting this promise.
      let resolveRespawn!: () => void;
      let respawnSettled = false;
      const respawnTeardown = new Promise<void>((res) => {
        resolveRespawn = res;
      }).then(() => {
        respawnSettled = true;
      });
      internals.respawns.add(respawnTeardown);

      // Kick off init() without awaiting. It will fail its spawn (the
      // config imports a non-existent plugin) within ~tens of ms.
      let initSettled = false;
      let initRejected = false;
      let settledAtInitReject = false;
      const initP = pool
        .init()
        .then(
          () => {
            /* unexpected resolve — asserted below */
          },
          () => {
            initRejected = true;
            // Capture whether the injected respawn had ALREADY settled at
            // the instant init() rejected. With the fix this is true (init
            // awaited it); without the fix it is false (init threw first).
            settledAtInitReject = respawnSettled;
          },
        )
        .finally(() => {
          initSettled = true;
        });

      // Wait well beyond the measured ~32ms spawn-failure latency so the
      // spawn has DEFINITELY failed. After this, the only thing that can
      // keep init() pending is the fix's `await` on the injected respawn.
      await new Promise((r) => setTimeout(r, 800));

      // Discriminating assertion: the spawn has failed, yet init() is
      // still pending — it is blocked awaiting the injected respawn.
      // Without the fix init() would have rejected during this window.
      expect(initSettled).toBe(false);
      expect(respawnSettled).toBe(false);

      // Release the injected respawn; init() can now finish and throw.
      resolveRespawn();
      await initP;

      expect(initRejected).toBe(true);
      // init() awaited the respawn's teardown BEFORE throwing.
      expect(settledAtInitReject).toBe(true);
      // Post-conditions of the init-failure branch still hold.
      expect(internals.closed).toBe(true);
      expect(internals.workers).toHaveLength(0);
    }, 15_000);
  },
);
