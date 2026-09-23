import { describe, test, expect, rs, afterEach } from 'rstack/test';
import { EventEmitter } from 'node:events';
import os from 'node:os';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';
import type {
  LintTask,
  WorkerPoolOptions,
} from '../../src/eslint-plugin/worker-pool.js';

import {
  LOCAL_CONFIG_DIR,
  localConfigs,
  task,
} from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';
import {
  runPoolScenario,
  formatScenarioFailure,
  POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
} from './pool-isolation/harness.js';

class QueueFakeWorker extends EventEmitter {
  readonly posted: unknown[] = [];
  terminateCalls = 0;

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
    this.terminateCalls++;
    queueMicrotask(() => this.emit('exit', 0));
    return Promise.resolve(0);
  }
}

async function controlledPool(options: Partial<WorkerPoolOptions> = {}) {
  const logs: string[] = [];
  const pool = new WorkerPool({
    configs: localConfigs,
    workerCount: 4,
    warmupWorkerCount: 2,
    retryCap: 0,
    ...options,
    onLog: (record) => logs.push(record.text),
  });
  const state = pool as any;
  const starts: Array<{
    id: number;
    worker: QueueFakeWorker;
    ready(): void;
    fail(): void;
  }> = [];
  state.spawnWorker = (id: number) =>
    new Promise((resolve, reject) => {
      const worker = new QueueFakeWorker();
      const slot = {
        id,
        worker,
        ready: true,
        exited: false,
        respawning: false,
        inflight: new Map(),
        crashCount: 0,
      };
      let settled = false;
      starts.push({
        id,
        worker,
        ready() {
          if (settled) return;
          settled = true;
          state.attachOngoingHandlers(slot);
          resolve(slot);
        },
        fail() {
          if (settled) return;
          settled = true;
          reject(new Error('injected expansion failure'));
        },
      });
    });
  const init = pool.init();
  starts.forEach((start) => start.ready());
  await init;
  const finish = (index: number) => {
    const worker = starts[index].worker;
    const message = worker.posted.at(-1) as {
      taskId: number;
      request: LintTask;
    };
    worker.emit('message', {
      kind: 'result',
      taskId: message.taskId,
      result: {
        filePath: message.request.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
      },
    });
  };
  const close = async () => {
    const shutdown = pool.shutdown();
    starts.forEach((start) => start.ready());
    await shutdown;
  };
  return { pool, state, starts, finish, close, logs };
}

describe('WorkerPool demand-driven capacity', () => {
  afterEach(() => rs.restoreAllMocks());

  test.each([
    [1, 1],
    [2, 2],
    [16, 2],
  ])(
    'available parallelism %i warms %i workers and still allows growth to eight',
    async (parallelism, warmup) => {
      rs.spyOn(os, 'availableParallelism').mockReturnValue(parallelism);
      const h = await controlledPool({
        workerCount: undefined,
        warmupWorkerCount: undefined,
      });
      try {
        expect(h.starts).toHaveLength(warmup);
        const batch = h.pool.lintBatch(
          Array.from({ length: 8 }, (_, i) => task(`default${i}.ts`, '')),
        );
        expect(h.starts).toHaveLength(8);
        h.starts.forEach((start) => start.ready());
        await Promise.all([...h.state.startingWorkers]);
        h.starts.forEach((_, i) => h.finish(i));
        expect((await batch).every((result) => !result.parseError)).toBe(true);
      } finally {
        await h.close();
      }
    },
  );

  test.each([
    { options: { workerCount: 1 }, expected: 1 },
    { options: { configs: [] }, expected: 0 },
  ])(
    'default warmup respects single-worker and plugin-free pools: %j',
    async ({ options, expected }) => {
      rs.spyOn(os, 'availableParallelism').mockReturnValue(16);
      const h = await controlledPool({
        workerCount: undefined,
        warmupWorkerCount: undefined,
        ...options,
      });
      try {
        expect(h.starts).toHaveLength(expected);
      } finally {
        await h.close();
      }
    },
  );

  test('warms two workers, reuses idle workers, and caps warmup at the maximum', async () => {
    for (const [maximum, warmup, expected] of [
      [5, 2, 2],
      [5, 4, 4],
      [1, 4, 1],
    ]) {
      const h = await controlledPool({
        workerCount: maximum,
        warmupWorkerCount: warmup,
      });
      try {
        expect(h.starts).toHaveLength(expected);
        const batch = h.pool.lintBatch([task('one.ts', '')]);
        expect(h.starts).toHaveLength(expected);
        h.finish(0);
        await expect(batch).resolves.toMatchObject([{ filePath: 'one.ts' }]);
        expect(h.starts).toHaveLength(expected);
      } finally {
        await h.close();
      }
    }
  });

  test('concurrent batches reserve starting capacity, preserve FIFO, and never exceed the maximum', async () => {
    const h = await controlledPool();
    try {
      const a = h.pool.lintBatch([0, 1, 2].map((i) => task(`a${i}.ts`, '')));
      expect(h.starts).toHaveLength(3);
      const b = h.pool.lintBatch([0, 1, 2].map((i) => task(`b${i}.ts`, '')));
      expect(h.starts).toHaveLength(4);
      for (let i = 0; i < 50; i++) h.state.kickQueue();
      expect(h.starts.map((start) => start.id)).toEqual([0, 1, 2, 3]);

      const starting = [...h.state.startingWorkers] as Promise<void>[];
      // Later spawn completes first; its reservation still occupies one slot.
      h.starts[3].ready();
      await starting[1];
      expect((h.starts[3].worker.posted[0] as any).request.filePath).toBe(
        'a2.ts',
      );
      h.starts[2].ready();
      await starting[0];
      expect((h.starts[2].worker.posted[0] as any).request.filePath).toBe(
        'b0.ts',
      );
      h.finish(0);
      h.finish(1);
      expect((h.starts[0].worker.posted[1] as any).request.filePath).toBe(
        'b1.ts',
      );
      expect((h.starts[1].worker.posted[1] as any).request.filePath).toBe(
        'b2.ts',
      );
      for (let i = 0; i < 4; i++) h.finish(i);
      const [resultA, resultB] = await Promise.all([a, b]);
      expect(resultA.map((r) => r.filePath)).toEqual([
        'a0.ts',
        'a1.ts',
        'a2.ts',
      ]);
      expect(resultB.map((r) => r.filePath)).toEqual([
        'b0.ts',
        'b1.ts',
        'b2.ts',
      ]);
      expect(h.starts).toHaveLength(4);
    } finally {
      await h.close();
    }
  });

  test('already-cancelled queued tasks do not start extra workers', async () => {
    const h = await controlledPool();
    try {
      const active = h.pool.lintBatch([task('a.ts', ''), task('b.ts', '')]);
      const cancelled = h.pool.lintBatch(
        Array.from({ length: 100 }, (_, i) => task(`cancelled${i}.ts`, '')),
        (id) => h.pool.cancelTask(id),
      );
      expect(h.starts).toHaveLength(2);
      h.finish(0);
      h.finish(1);
      await active;
      expect((await cancelled).every((result) => result.cancelled)).toBe(true);
      expect(h.starts).toHaveLength(2);
    } finally {
      await h.close();
    }
  });

  test('crash replacements and growth share the capacity limit across concurrent batches', async () => {
    const h = await controlledPool({ retryCap: 1 });
    try {
      const a = h.pool.lintBatch([0, 1, 2].map((i) => task(`a${i}.ts`, '')));
      h.starts[0].worker.emit('exit', 1);
      const b = h.pool.lintBatch([0, 1].map((i) => task(`b${i}.ts`, '')));
      // Slot 0 is replaced, while new slots 2 and 3 are still starting.
      // Its replacement occupies the original slot's reserved capacity.
      expect(h.starts.map((start) => start.id)).toEqual([0, 1, 2, 0, 3]);
      expect(h.state.startingWorkers.size).toBe(2);
      expect(h.state.respawns.size).toBe(1);
      for (let i = 0; i < 50; i++) h.state.kickQueue();
      expect(h.starts).toHaveLength(5);

      h.starts[3].ready();
      await Promise.all([...h.state.respawns]);
      h.starts[4].ready();
      h.starts[2].ready();
      await Promise.all([...h.state.startingWorkers]);
      expect(h.state.workers).toHaveLength(4);
      expect(new Set(h.state.workers.map((slot: any) => slot.id)).size).toBe(4);
      for (let i = 1; i < 5; i++) h.finish(i);
      const [resultsA, resultsB] = await Promise.all([a, b]);
      expect(resultsA[0].parseError).toMatch(/^worker_crashed/);
      expect(resultsA.slice(1).every((result) => !result.parseError)).toBe(
        true,
      );
      expect(resultsB.map((result) => result.filePath)).toEqual([
        'b0.ts',
        'b1.ts',
      ]);
      expect(resultsB.every((result) => !result.parseError)).toBe(true);
      expect(h.starts).toHaveLength(5);
    } finally {
      await h.close();
    }
  });

  test('failed expansion retains warm workers and does not retry on every task', async () => {
    const h = await controlledPool();
    try {
      const batch = h.pool.lintBatch(
        [0, 1, 2, 3].map((i) => task(`f${i}.ts`, '')),
      );
      const starting = [...h.state.startingWorkers] as Promise<void>[];
      h.starts[2].fail();
      h.starts[3].fail();
      await Promise.allSettled(starting);
      expect(
        h.logs.filter((log) => log.includes('expansion failed')),
      ).toHaveLength(2);
      h.finish(0);
      h.finish(1);
      h.finish(0);
      h.finish(1);
      expect((await batch).every((result) => !result.parseError)).toBe(true);
      const later = h.pool.lintBatch(
        [0, 1, 2].map((i) => task(`later${i}.ts`, '')),
      );
      expect(h.starts).toHaveLength(4);
      h.finish(0);
      h.finish(1);
      h.finish(0);
      await later;
    } finally {
      await h.close();
    }
  });

  test('failed task serialization leaves idle workers and does not expand the pool', async () => {
    const h = await controlledPool();
    try {
      for (const { worker } of h.starts) {
        const post = worker.postMessage.bind(worker);
        worker.postMessage = (message) => {
          if ((message as { kind: string }).kind === 'task') {
            throw Object.assign(new Error('cannot clone task'), {
              name: 'DataCloneError',
            });
          }
          post(message);
        };
      }
      const results = await h.pool.lintBatch(
        Array.from({ length: 100 }, (_, i) => task(`invalid${i}.ts`, '')),
      );
      expect(
        results.every((result) =>
          result.parseError?.startsWith('postMessage_failed'),
        ),
      ).toBe(true);
      expect(h.starts).toHaveLength(2);
    } finally {
      await h.close();
    }
  });

  test('shutdown and repeated shutdown await pending expansion and reap late workers', async () => {
    const h = await controlledPool();
    const batch = h.pool.lintBatch(
      [0, 1, 2, 3].map((i) => task(`f${i}.ts`, '')),
    );
    let closed = false;
    const first = h.pool.shutdown();
    void first.then(() => {
      closed = true;
    });
    expect(h.pool.shutdown()).toBe(first);
    await Promise.resolve();
    expect(closed).toBe(false);
    expect(
      (await batch).every((result) => result.parseError === 'shutdown'),
    ).toBe(true);
    h.starts[2].ready();
    h.starts[3].ready();
    await first;
    expect(
      h.starts.slice(2).map((start) => start.worker.terminateCalls),
    ).toEqual([1, 1]);
    expect(h.state.workers).toEqual([]);
    expect(h.state.startingWorkers.size).toBe(0);
    expect(h.starts).toHaveLength(4);
  });

  test('pending expansion can serve queued work after every warm worker crashes', async () => {
    const h = await controlledPool({ workerCount: 3 });
    try {
      const batch = h.pool.lintBatch(
        [0, 1, 2].map((i) => task(`f${i}.ts`, '')),
      );
      h.starts[0].worker.emit('exit', 1);
      h.starts[1].worker.emit('exit', 1);
      expect(h.state.pendingQueue).toHaveLength(1);
      const starting = [...h.state.startingWorkers] as Promise<void>[];
      h.starts[2].ready();
      await Promise.all(starting);
      h.finish(2);
      const results = await batch;
      expect(
        results
          .slice(0, 2)
          .every((result) => result.parseError?.startsWith('worker_crashed')),
      ).toBe(true);
      expect(results[2].parseError).toBeUndefined();
      expect(h.starts).toHaveLength(3);
    } finally {
      await h.close();
    }
  });

  test('failed expansion drains the queue when all warm workers are also dead', async () => {
    const h = await controlledPool({ workerCount: 3 });
    try {
      const ids: number[] = [];
      const batch = h.pool.lintBatch(
        [0, 1, 2, 3].map((i) => task(`f${i}.ts`, '')),
        (id) => ids.push(id),
      );
      expect(h.pool.cancelTask(ids[3])).toBe(true);
      h.starts[0].worker.emit('exit', 1);
      h.starts[1].worker.emit('exit', 1);
      h.starts[2].fail();
      const results = await batch;
      expect(results[2].parseError).toBe('pool_degraded');
      expect(results[3].cancelled).toBe(true);
      expect(results[3].parseError).toBeUndefined();
      expect(h.state.pendingQueue).toEqual([]);
    } finally {
      await h.close();
    }
  });

  test('rejects invalid capacity instead of admitting unbounded growth', () => {
    for (const workerCount of [-1, 1.5, NaN, Infinity]) {
      expect(
        () => new WorkerPool({ configs: localConfigs, workerCount }),
      ).toThrow(/workerCount/);
    }
    for (const warmupWorkerCount of [0, -1, 1.5, NaN, Infinity]) {
      expect(
        () => new WorkerPool({ configs: localConfigs, warmupWorkerCount }),
      ).toThrow(/warmupWorkerCount/);
    }
  });

  test('shutdown from the enqueue callback settles the rest of the batch without allocating more tasks', async () => {
    const h = await controlledPool();
    try {
      let shutdown: Promise<void> | undefined;
      let callbacks = 0;
      const batch = h.pool.lintBatch(
        Array.from({ length: 100 }, (_, i) => task(`closing${i}.ts`, '')),
        () => {
          callbacks++;
          shutdown ??= h.pool.shutdown();
        },
      );
      expect(h.state.pendingQueue).toHaveLength(0);
      const results = await batch;
      await shutdown;
      expect(callbacks).toBe(1);
      expect(results).toHaveLength(100);
      expect(results.every((result) => result.parseError === 'shutdown')).toBe(
        true,
      );
      expect(h.state.cancelPool.slotInUse.some((used: number) => used)).toBe(
        false,
      );
      expect(h.starts).toHaveLength(2);
      await expect(h.pool.lintBatch([])).rejects.toThrow(/closed/);
    } finally {
      await h.close();
    }
  });

  test('shutdown from a crash log callback prevents a replacement from starting after close', async () => {
    const h = await controlledPool({ workerCount: 2, retryCap: 1 });
    try {
      let shutdown: Promise<void> | undefined;
      h.state.opts.onLog = (record: { text: string }) => {
        if (record.text.includes('respawning')) {
          shutdown = h.pool.shutdown();
        }
      };
      const batch = h.pool.lintBatch([
        task('crash.ts', ''),
        task('busy.ts', ''),
      ]);
      h.starts[0].worker.emit('exit', 1);
      expect(shutdown).toBeDefined();
      expect(h.starts).toHaveLength(2);
      await shutdown;
      const results = await batch;
      expect(results[0].parseError).toMatch(/^worker_crashed/);
      expect(results[1].parseError).toBe('shutdown');
      expect(h.state.respawns.size).toBe(0);
    } finally {
      await h.close();
      await Promise.allSettled([...h.state.respawns]);
    }
  });
});

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
      POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
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
      POOL_SCENARIO_OUTER_DEADLOCK_SENTINEL_MS,
    );
  },
);
