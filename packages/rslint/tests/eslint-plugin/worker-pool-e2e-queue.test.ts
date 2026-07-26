import { describe, test, expect } from '@rstest/core';
import { EventEmitter } from 'node:events';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type { LintTask } from '../../src/eslint-plugin/worker-pool.js';

import { LOCAL_CONFIG_DIR, localConfigs } from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';
import {
  runPoolScenario,
  formatScenarioFailure,
  POOL_SCENARIO_OUTER_TIMEOUT_MS,
} from './pool-isolation/harness.js';

class QueueFakeWorker extends EventEmitter {
  readonly posted: unknown[] = [];

  postMessage(message: unknown): void {
    this.posted.push(message);
    if (
      (message as { kind?: string }).kind === 'shutdown' &&
      this.listenerCount('exit') > 0
    ) {
      queueMicrotask(() => this.emit('exit', 0));
    }
  }

  terminate(): Promise<number> {
    return Promise.resolve(0);
  }
}

/**
 * WorkerPool end-to-end — queue model: tasks wait in `pendingQueue`
 * for an idle worker. Pins the design properties of the queue
 * refactor — per-task timers start at dispatch (not enqueue), large
 * batches / backpressure complete, kickQueue is idempotent — plus the
 * teardown invariants: queued tasks settle as parseError:'shutdown' on
 * shutdown, and an exhausted-retryCap pool drains queued / future
 * batches as parseError:'pool_degraded' instead of hanging.
 */

// win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (see that file for the
// nodejs/node#34567 rationale); the flag is false so these run on win32 too.
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
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

      // lintBatch enqueues synchronously before returning its Promise. Assert
      // that state directly instead of sleeping and hoping the scheduler ran.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      expect((pool as any).pendingQueue.length).toBe(5);

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
    });

    // Driving every slot past its respawn cap (crashCount=cap + terminate)
    // drains the in-flight batch as parseError:pool_degraded. The terminate of
    // an oxc-napi worker can native-abort on Windows — run it isolated (see
    // ./pool-isolation). The in-child asserts pin the pool_degraded drain.
    test(
      'all workers degraded → pendingQueue drains as parseError:pool_degraded (no hang)',
      async () => {
        const r = await runPoolScenario('all-degraded');
        expect(r.verdict, formatScenarioFailure(r)).toBe('PASS');
      },
      POOL_SCENARIO_OUTER_TIMEOUT_MS,
    );

    // Queue-model regression suite — five tests pinning the design
    // properties that the Finding 3 refactor introduced.

    test('queue: timers start only when each queued task is dispatched', async () => {
      // Pre-refactor every queued task armed its timer synchronously.
      // The queue model must arm only the one dispatched task; the remaining
      // 19 receive a timer only when a worker actually takes them. An
      // in-memory Worker drives each result event deterministically here; the
      // real-worker backpressure test below separately covers integration.
      const nonFiringTaskTimeoutMs = 24 * 60 * 60 * 1_000;
      const pool = new WorkerPool({
        configs: [],
        taskTimeoutMs: nonFiringTaskTimeoutMs,
      });
      const internals = pool as any;
      const worker = new QueueFakeWorker();
      const slot = {
        id: 0,
        worker,
        ready: true,
        exited: false,
        respawning: false,
        inflight: new Map(),
        crashCount: 0,
      };
      internals.opts.workerCount = 1;
      internals.workers = [slot];
      internals.attachOngoingHandlers(slot);

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
      // Instrument this pool instance rather than retaining a global timer
      // patch across an async boundary. Every dequeue enters dispatch
      // synchronously, so each wrapper captures its exact product timer and
      // restores the global before yielding.
      const originalDispatch = internals.dispatchToWorker;
      const taskTimers: NodeJS.Timeout[] = [];
      let dispatchCalls = 0;
      internals.dispatchToWorker = function (...args: unknown[]) {
        dispatchCalls++;
        const originalSetTimeout = globalThis.setTimeout;
        globalThis.setTimeout = ((
          callback: (...callbackArgs: any[]) => void,
          delay?: number,
          ...callbackArgs: any[]
        ) => {
          const handle = originalSetTimeout(callback, delay, ...callbackArgs);
          if (delay === nonFiringTaskTimeoutMs) {
            handle.unref();
            taskTimers.push(handle);
          }
          return handle;
        }) as typeof setTimeout;
        try {
          return originalDispatch.apply(this, args);
        } finally {
          globalThis.setTimeout = originalSetTimeout;
        }
      };

      try {
        const resultsPromise = pool.lintBatch(tasks);
        expect(dispatchCalls).toBe(1);
        expect(taskTimers).toHaveLength(1);
        expect(internals.pendingQueue).toHaveLength(19);

        for (let i = 0; i < tasks.length; i++) {
          expect(taskTimers).toHaveLength(i + 1);
          expect(internals.pendingQueue).toHaveLength(tasks.length - i - 1);
          expect(worker.posted).toHaveLength(i + 1);
          const message = worker.posted[i] as {
            kind: string;
            taskId: number;
          };
          expect(message.kind).toBe('task');
          worker.emit('message', {
            kind: 'result',
            taskId: message.taskId,
            result: {
              filePath: tasks[i].filePath,
              diagnostics: [{ ruleName: 'local/no-null' }],
              fixes: [],
              suggestionsCount: 0,
              cancelled: false,
            },
          });
          expect(
            (
              taskTimers[i] as unknown as {
                _destroyed?: boolean;
              }
            )._destroyed,
          ).toBe(true);
        }

        const results = await resultsPromise;
        expect(dispatchCalls).toBe(20);
        expect(taskTimers).toHaveLength(20);
        expect(
          taskTimers.every(
            (handle) =>
              (handle as unknown as { _destroyed?: boolean })._destroyed ===
              true,
          ),
        ).toBe(true);
        expect(results).toHaveLength(20);
        for (const result of results) {
          expect(result.parseError).toBeUndefined();
          expect(result.diagnostics).toHaveLength(1);
        }
      } finally {
        internals.dispatchToWorker = originalDispatch;
        for (const handle of taskTimers) clearTimeout(handle);
        await pool.shutdown();
      }
    });

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
    });

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
    });

    // A SECOND batch issued AFTER the pool settled into the terminal degraded
    // state must also resolve pool_degraded (not hang) — Fix A. Reaching that
    // state needs a forced terminate of an oxc-napi worker, which can
    // native-abort on Windows, so run it isolated (see ./pool-isolation). The
    // in-child asserts pin the terminal-state sanity checks and both drains.
    test(
      'lintBatch issued AFTER the pool settled into the terminal degraded state resolves as pool_degraded (does not hang)',
      async () => {
        const r = await runPoolScenario('lint-batch-after-degraded');
        expect(r.verdict, formatScenarioFailure(r)).toBe('PASS');
      },
      POOL_SCENARIO_OUTER_TIMEOUT_MS,
    );
  },
);
