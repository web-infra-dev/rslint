import { describe, test, expect } from '@rstest/core';

import { WorkerPool } from '../src/worker-pool.js';
import type { LintTask } from '../src/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';

/**
 * WorkerPool end-to-end — queue model: tasks wait in `pendingQueue`
 * for an idle worker. Pins the design properties of the queue
 * refactor — per-task timers start at dispatch (not enqueue), large
 * batches / backpressure complete, kickQueue is idempotent — plus the
 * teardown invariants: queued tasks settle as parseError:'shutdown' on
 * shutdown, and an exhausted-retryCap pool drains queued / future
 * batches as parseError:'pool_degraded' instead of hanging.
 */

// Skipped on windows: tearing down a worker that has oxc (a napi addon)
// loaded aborts below the JS layer there (nodejs/node#34567) and crashes
// the rstest worker running this file. These e2e tests spawn real
// workers and tear them down, so they are windows-skipped; they still
// run on linux/macOS.
describe.skipIf(process.platform === 'win32')(
  'WorkerPool end-to-end with a local fixture plugin',
  () => {
    // R1 (re-purposed for the queue model): the original R1 finding
    // — `Promise.all` rejecting the whole batch when no worker was
    // available — was eliminated by the queue refactor. Tasks now wait
    // in `pendingQueue` for an idle worker rather than failing
    // synchronously; transient "all workers not-ready" (mid-respawn,
    // mid-cancel) becomes a brief backlog stall instead of a batch-
    // wide reject.
    //
    // The new equivalent invariant: when the pool shuts down WHILE
    // tasks are queued, those queued tasks must still settle (not
    // hang) with a `parseError: 'shutdown'` marker — the same
    // result-shaped failure inflight tasks get. This guards against
    // a regression where the queue path drops settlement of queued
    // promises during teardown.
    test('R1: queued tasks resolve as parseError:shutdown when pool shuts down mid-batch', async () => {
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
      });
      await pool.init();
      // Force the lone worker to look busy (not idle) so freshly-
      // enqueued tasks stay in `pendingQueue` instead of being
      // dispatched immediately. Without this the worker would grab
      // them before shutdown lands.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const ws = (pool as any).workers as Array<{ ready: boolean }>;
      ws[0].ready = false;

      // Kick off a 5-file batch. With the worker not-ready, all 5
      // tasks land in pendingQueue and the Promise.all stays pending.
      const batchP = pool.lintBatch(
        [1, 2, 3, 4, 5].map((i) => ({
          filePath: `q${i}.ts`,
          text: 'const x = null;\n',
          rules: { 'local/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: LOCAL_CONFIG_DIR,
        })),
      );

      // Let the enqueue actually happen before shutting down.
      await new Promise((r) => setTimeout(r, 30));

      // Tear down. shutdown() must drain pendingQueue and resolve
      // every queued task with the 'shutdown' marker.
      await pool.shutdown();
      const result = await batchP;
      expect(result).toHaveLength(5);
      for (const r of result) {
        expect(r.parseError).toBe('shutdown');
        expect(r.diagnostics).toEqual([]);
        expect(r.cancelled).toBe(false);
      }
      // No leaked worker slot.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect((pool as any).workers.length).toBe(0);
      // No leaked queue entries.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect((pool as any).pendingQueue.length).toBe(0);
    }, 15_000);

    test('all workers degraded → pendingQueue drains as parseError:pool_degraded (no hang)', async () => {
      // Regression for review P0 #2. When every worker exhausts its
      // respawn cap (default 3), the 'exit' handler used to JUST log;
      // `kickQueue` skips not-ready slots, so anything in
      // `pendingQueue` waited forever — `await pool.lintBatch(tasks)`
      // never returned and the host had no path to recover.
      //
      // Fix: in the cap-reached branch, drain `pendingQueue` with
      // `parseError: 'pool_degraded'` once every slot is `ready=false`.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
        retryCap: 1,
      });
      await pool.init();

      // Wedge: enqueue a couple of tasks but hold the worker non-ready
      // so kickQueue can't dispatch.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const internals = pool as any;
      internals.workers[0].ready = false;

      const batchP = pool.lintBatch(
        [1, 2].map((i) => ({
          filePath: `q${i}.ts`,
          text: 'const x = null;\n',
          rules: { 'local/no-null': { options: [] } },
          collectFixes: false,
          suggestionsMode: 'off',
          configKey: LOCAL_CONFIG_DIR,
        })),
      );
      await new Promise((r) => setTimeout(r, 30));

      // Simulate the terminal "all workers exhausted retryCap" state by
      // bumping crashCount and re-running the exit handler. The exit
      // handler is wired during `attachOngoingHandlers`; the simplest
      // path is to terminate the worker (triggers a real 'exit') after
      // setting crashCount to the cap.
      internals.workers[0].crashCount = internals.opts.retryCap;
      await internals.workers[0].worker.terminate();

      const result = await batchP;
      expect(result).toHaveLength(2);
      for (const r of result) {
        expect(r.parseError).toBe('pool_degraded');
        expect(r.cancelled).toBe(false);
        expect(r.diagnostics).toEqual([]);
      }
      // Pending queue fully drained.
      expect(internals.pendingQueue.length).toBe(0);

      // Tear down cleanly — the pool is in a degraded state but
      // shutdown() should still work.
      await pool.shutdown();
    }, 15_000);

    // Queue-model regression suite — five tests pinning the design
    // properties that the Finding 3 refactor introduced.

    test('queue: large batch on a single worker completes without queue-time timeout', async () => {
      // 20 tasks on 1 worker × 600ms timer = 12 s total work. Pre-
      // refactor, the timer started at lintBatch enqueue time, so
      // tasks 2-20 raced their 600ms deadline against a 1-worker
      // serial backlog and most would land as task_timeout. Post-
      // refactor, each task's timer starts only when the worker
      // actually takes it off the queue — so every task gets its
      // full 600ms execution budget regardless of position.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
        taskTimeoutMs: 600,
      });
      await pool.init();
      try {
        const tasks: LintTask[] = [];
        for (let i = 0; i < 20; i++) {
          tasks.push({
            filePath: `f${i}.ts`,
            text: 'const x = null;\n',
            rules: { 'local/no-null': { options: [] } },
            collectFixes: false,
            suggestionsMode: 'off',
            configKey: LOCAL_CONFIG_DIR,
          });
        }
        const results = await pool.lintBatch(tasks);
        expect(results).toHaveLength(20);
        // Every file has exactly ONE `null` literal → exactly ONE
        // `local/no-null` diagnostic. `toBe(1)` (not `> 0`) proves the
        // task ran AND catches a queue-reuse regression where a file is
        // dispatched twice and its diagnostics merged (→ 2). None should
        // be task_timeout: a queue-time timeout regression would land
        // most as parseError.
        for (const r of results) {
          expect(r.parseError).toBeUndefined();
          expect(r.diagnostics.length).toBe(1);
        }
      } finally {
        await pool.shutdown();
      }
    }, 30_000);

    test('queue: kickQueue is idempotent + safe with empty queue', async () => {
      // kickQueue gets called from multiple async paths (result
      // handler, exit handler, lintBatch, postMessage_failed); a
      // bug that double-dispatches a task or trips on an empty queue
      // would surface as either an extra postMessage (file double-
      // linted) or an exception. Hammer it.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 2,
      });
      await pool.init();
      try {
        // No queued tasks — repeated kicks must be no-ops.
        for (let i = 0; i < 20; i++) {
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (pool as any).kickQueue();
        }
        // Now actually lint — pool should still work normally.
        const r = await pool.lintBatch([
          {
            filePath: 'a.ts',
            text: 'const x = null;\n',
            rules: { 'local/no-null': { options: [] } },
            collectFixes: false,
            suggestionsMode: 'off',
            configKey: LOCAL_CONFIG_DIR,
          },
        ]);
        expect(r).toHaveLength(1);
        // Single file, single `null` literal → exactly ONE diagnostic.
        // `toBe(1)` (not `> 0`) catches a kickQueue double-dispatch that
        // would re-lint the file and surface duplicate diagnostics.
        expect(r[0].diagnostics.length).toBe(1);
      } finally {
        await pool.shutdown();
      }
    }, 15_000);

    test('queue: when batch size > worker count, all tasks complete (backpressure)', async () => {
      // 30 tasks on 2 workers — only 2 inflight at a time, the
      // remaining 28 sit in pendingQueue and get dispatched as
      // workers complete. Pre-refactor, all 30 were postMessage'd
      // immediately, the worker postMessage queue grew, later tasks
      // raced their timer. This test verifies backpressure works
      // without artificial timing assumptions.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 2,
      });
      await pool.init();
      try {
        const tasks: LintTask[] = [];
        for (let i = 0; i < 30; i++) {
          tasks.push({
            filePath: `bp${i}.ts`,
            text: 'const x = null;\n',
            rules: { 'local/no-null': { options: [] } },
            collectFixes: false,
            suggestionsMode: 'off',
            configKey: LOCAL_CONFIG_DIR,
          });
        }
        const results = await pool.lintBatch(tasks);
        expect(results).toHaveLength(30);
        // Every task ran successfully — no parseError, and each file's
        // single `null` literal yields exactly ONE diagnostic. `toBe(1)`
        // (not `> 0`) additionally catches a backpressure/queue-reuse
        // regression that double-dispatches a file (→ 2 diagnostics).
        for (const r of results) {
          expect(r.parseError).toBeUndefined();
          expect(r.diagnostics.length).toBe(1);
        }
      } finally {
        await pool.shutdown();
      }
    }, 30_000);

    test('lintBatch issued AFTER the pool settled into the terminal degraded state resolves as pool_degraded (does not hang)', async () => {
      // Regression for Fix A. The existing "all workers degraded →
      // pendingQueue drains as parseError:pool_degraded" test only covers
      // tasks ALREADY in-flight when the terminal exit fires. A
      // `lintBatch` issued AFTER the pool has settled into the degraded
      // state was never drained: it passed the `closed` guard (pool is
      // degraded, not closed) and the `workers.length === 0` guard, then
      // `kickQueue` skipped the not-ready slot and NO future event would
      // ever call the drain again → `Promise.all` never settled →
      // permanent hang. Fix: `lintBatch` calls
      // `drainQueueIfAllSlotsDegraded()` after the enqueue/kick so a
      // batch handed to an already-terminal pool resolves immediately.
      const pool = new WorkerPool({
        configs: localConfigs,
        workerCount: 1,
        retryCap: 1,
      });
      await pool.init();

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const internals = pool as any;

      // Drive the pool to the terminal degraded state EXACTLY the way the
      // existing in-flight-drain test does: hold the slot non-ready,
      // enqueue a first batch, push crashCount to the cap, and terminate
      // the worker so its real 'exit' handler takes the cap-reached
      // branch and drains that first batch as pool_degraded.
      internals.workers[0].ready = false;
      const firstBatch = pool.lintBatch([
        task('first.ts', 'const x = null;\n'),
      ]);
      await new Promise((r) => setTimeout(r, 30));
      internals.workers[0].crashCount = internals.opts.retryCap;
      await internals.workers[0].worker.terminate();

      const firstResult = await firstBatch;
      expect(firstResult).toHaveLength(1);
      expect(firstResult[0].parseError).toBe('pool_degraded');

      // Pool is now terminal: slot is ready=false, respawning=false,
      // EXITED (the worker thread really died — this is the discriminator
      // that separates this from a benign not-ready-but-alive worker),
      // crashCount>=cap, and NOT closed. Sanity-check that state so the
      // test is exercising the intended scenario.
      expect(internals.closed).toBe(false);
      expect(internals.workers).toHaveLength(1);
      expect(internals.workers[0].ready).toBe(false);
      expect(internals.workers[0].respawning).toBe(false);
      expect(internals.workers[0].exited).toBe(true);

      // Issue a SECOND batch against the already-degraded pool. With the
      // fix it must RESOLVE (every result pool_degraded). Without the fix
      // it hangs forever — race it against a rejecting timer so the
      // negative-check produces a clean, fast failure instead of wedging
      // the whole suite on the per-test timeout.
      const secondBatch = pool.lintBatch([
        task('second-a.ts', 'const y = null;\n'),
        task('second-b.ts', 'const z = null;\n'),
      ]);
      const guard = new Promise<never>((_, rej) =>
        setTimeout(
          () => rej(new Error('second lintBatch hung — Fix A regression')),
          4_000,
        ),
      );
      const secondResult = await Promise.race([secondBatch, guard]);
      expect(secondResult).toHaveLength(2);
      for (const r of secondResult) {
        expect(r.parseError).toBe('pool_degraded');
        expect(r.cancelled).toBe(false);
        expect(r.diagnostics).toEqual([]);
      }
      expect(internals.pendingQueue.length).toBe(0);

      await pool.shutdown();
    }, 15_000);
  },
);
