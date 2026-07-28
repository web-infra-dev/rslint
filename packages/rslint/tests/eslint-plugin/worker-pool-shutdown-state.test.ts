import { EventEmitter } from 'node:events';

import { describe, expect, test } from '@rstest/core';

import { WorkerPool } from '../../src/eslint-plugin/worker-pool.js';

class FakePipe {
  constructor(
    private readonly name: string,
    private readonly events: string[],
  ) {}

  destroy(): void {
    this.events.push(`${this.name}:destroy`);
  }
}

class FakeWorker extends EventEmitter {
  readonly stdout: FakePipe;
  readonly stderr: FakePipe;
  readonly posted: unknown[] = [];
  terminateCalls = 0;
  throwOnPost = false;
  onPostMessage?: () => void;

  constructor(readonly events: string[]) {
    super();
    this.stdout = new FakePipe('stdout', events);
    this.stderr = new FakePipe('stderr', events);
  }

  postMessage(message: unknown): void {
    this.events.push('postMessage');
    if (this.throwOnPost) throw new Error('postMessage failed');
    this.posted.push(message);
    this.onPostMessage?.();
  }

  terminate(): Promise<number> {
    this.terminateCalls++;
    this.events.push('terminate');
    return Promise.resolve(1);
  }
}

interface TestSlot {
  id: number;
  worker: FakeWorker;
  ready: boolean;
  exited: boolean;
  respawning: boolean;
  inflight: Map<number, never>;
  crashCount: number;
}

interface TestPoolState {
  opts: { workerCount: number };
  workers: TestSlot[];
  closed: boolean;
}

function makeSlot(worker: FakeWorker, id: number): TestSlot {
  return {
    id,
    worker,
    ready: true,
    exited: false,
    respawning: false,
    inflight: new Map(),
    crashCount: 0,
  };
}

function makePool(...workers: FakeWorker[]): {
  pool: WorkerPool;
  state: TestPoolState;
  slots: TestSlot[];
} {
  // Empty configs ensure construction itself cannot create a real Worker. The
  // tests then install structurally faithful in-memory slots.
  const pool = new WorkerPool({ configs: [] });
  const state = pool as unknown as TestPoolState;
  const slots = workers.map<TestSlot>(makeSlot);
  for (const slot of slots) {
    slot.worker.on('exit', () => {
      slot.ready = false;
      slot.exited = true;
    });
  }
  state.opts.workerCount = workers.length;
  state.workers = slots;
  return { pool, state, slots };
}

interface CapturedTimer {
  delay: number;
  cleared: boolean;
  fired: boolean;
  fire(): void;
}

function beginShutdownWithCapturedTimers(pool: WorkerPool): {
  shutdown: Promise<void>;
  timers: CapturedTimer[];
  restore(): void;
} {
  const originalSetTimeout = globalThis.setTimeout;
  const originalClearTimeout = globalThis.clearTimeout;
  const timers: CapturedTimer[] = [];

  globalThis.setTimeout = ((
    callback: (...args: unknown[]) => void,
    delay?: number,
    ...args: unknown[]
  ) => {
    const timer: CapturedTimer = {
      delay: delay ?? 0,
      cleared: false,
      fired: false,
      fire() {
        if (timer.cleared || timer.fired) return;
        timer.fired = true;
        callback(...args);
      },
    };
    timers.push(timer);
    return timer as unknown as NodeJS.Timeout;
  }) as typeof setTimeout;
  globalThis.clearTimeout = ((handle: NodeJS.Timeout) => {
    const timer = handle as unknown as CapturedTimer;
    if (timers.includes(timer)) timer.cleared = true;
  }) as typeof clearTimeout;

  try {
    const shutdown = pool.shutdown();
    return {
      shutdown,
      timers,
      restore() {
        globalThis.setTimeout = originalSetTimeout;
        globalThis.clearTimeout = originalClearTimeout;
      },
    };
  } catch (err) {
    globalThis.setTimeout = originalSetTimeout;
    globalThis.clearTimeout = originalClearTimeout;
    throw err;
  }
}

describe('WorkerPool shutdown state machine (no real Worker)', () => {
  test('worker exit after waiter registration cancels the fallback', async () => {
    const events: string[] = [];
    const worker = new FakeWorker(events);
    const { pool, state } = makePool(worker);
    const capture = beginShutdownWithCapturedTimers(pool);
    try {
      expect(capture.timers).toHaveLength(1);
      expect(capture.timers[0].delay).toBe(5_000);
      expect(worker.posted).toEqual([{ kind: 'shutdown' }]);

      worker.emit('exit', 0);
      expect(capture.timers[0].cleared).toBe(true);
    } finally {
      capture.restore();
    }

    await capture.shutdown;
    expect(worker.terminateCalls).toBe(0);
    expect(state.workers).toEqual([]);
  });

  test('exit while posting shutdown is observed without arming a stale timer', async () => {
    const events: string[] = [];
    const worker = new FakeWorker(events);
    const { pool, state } = makePool(worker);
    worker.onPostMessage = () => worker.emit('exit', 0);

    const capture = beginShutdownWithCapturedTimers(pool);
    capture.restore();
    await capture.shutdown;

    expect(capture.timers).toEqual([]);
    expect(worker.terminateCalls).toBe(0);
    expect(state.workers).toEqual([]);
  });

  test('fallback fires at the configured grace and closes pipes first', async () => {
    const events: string[] = [];
    const worker = new FakeWorker(events);
    const { pool, state } = makePool(worker);
    const capture = beginShutdownWithCapturedTimers(pool);
    try {
      expect(capture.timers).toHaveLength(1);
      expect(capture.timers[0].delay).toBe(5_000);
      capture.timers[0].fire();
    } finally {
      capture.restore();
    }

    await capture.shutdown;
    expect(events).toEqual([
      'postMessage',
      'stdout:destroy',
      'stderr:destroy',
      'terminate',
    ]);
    expect(worker.terminateCalls).toBe(1);
    expect(state.workers).toEqual([]);
  });

  test('postMessage failure still leaves a bounded terminate fallback', async () => {
    const events: string[] = [];
    const worker = new FakeWorker(events);
    worker.throwOnPost = true;
    const { pool, state } = makePool(worker);
    const capture = beginShutdownWithCapturedTimers(pool);
    try {
      expect(capture.timers).toHaveLength(1);
      expect(capture.timers[0].delay).toBe(5_000);
      capture.timers[0].fire();
    } finally {
      capture.restore();
    }

    await capture.shutdown;
    expect(worker.terminateCalls).toBe(1);
    expect(state.workers).toEqual([]);
  });

  test('already-exited slots and repeated shutdown do not arm timers', async () => {
    const events: string[] = [];
    const worker = new FakeWorker(events);
    const { pool, state, slots } = makePool(worker);
    slots[0].exited = true;
    slots[0].ready = false;

    const first = beginShutdownWithCapturedTimers(pool);
    first.restore();
    await first.shutdown;
    expect(first.timers).toEqual([]);
    expect(worker.terminateCalls).toBe(0);
    expect(state.workers).toEqual([]);

    const second = beginShutdownWithCapturedTimers(pool);
    second.restore();
    await second.shutdown;
    expect(second.timers).toEqual([]);
    expect(worker.terminateCalls).toBe(0);
  });

  test('empty pool shutdown arms no timer', async () => {
    const { pool, state } = makePool();
    const capture = beginShutdownWithCapturedTimers(pool);
    capture.restore();
    await capture.shutdown;

    expect(capture.timers).toEqual([]);
    expect(state.workers).toEqual([]);
  });

  test('a pending sibling respawn survives another respawn rejection and services the queue', async () => {
    const firstWorker = new FakeWorker([]);
    const secondWorker = new FakeWorker([]);
    const replacementWorker = new FakeWorker([]);
    const { pool, state, slots } = makePool(firstWorker, secondWorker);
    const internals = pool as any;
    internals.attachOngoingHandlers(slots[0]);
    internals.attachOngoingHandlers(slots[1]);
    slots[0].ready = false;
    slots[1].ready = false;

    const batch = pool.lintBatch([
      {
        filePath: 'queued.ts',
        text: '',
        rules: {},
        collectFixes: false,
        suggestionsMode: 'off',
      },
    ]);
    let batchSettled = false;
    void batch.then(
      () => {
        batchSettled = true;
      },
      () => {
        batchSettled = true;
      },
    );
    expect(internals.pendingQueue).toHaveLength(1);

    const replacementSlot = makeSlot(replacementWorker, 1);
    internals.attachOngoingHandlers(replacementSlot);
    let releaseSecondRespawn!: () => void;
    const secondRespawnGate = new Promise<void>((resolve) => {
      releaseSecondRespawn = resolve;
    });
    internals.spawnWorker = async (id: number) => {
      if (id === 0) throw new Error('first respawn rejected');
      expect(id).toBe(1);
      await secondRespawnGate;
      return replacementSlot;
    };

    // Start B's pending respawn first, then reject A's. Both transitions run
    // through the real ongoing exit handler; no state is injected manually.
    secondWorker.emit('exit', 42);
    firstWorker.emit('exit', 43);
    const respawns = [...internals.respawns] as Promise<void>[];
    expect(respawns).toHaveLength(2);
    await respawns[1];

    expect(slots[0].respawning).toBe(false);
    expect(slots[1].respawning).toBe(true);
    expect(batchSettled).toBe(false);
    expect(internals.pendingQueue).toHaveLength(1);

    releaseSecondRespawn();
    await Promise.all(respawns);
    expect(state.workers[1]).toBe(replacementSlot);
    expect(replacementWorker.posted).toHaveLength(1);
    expect(internals.pendingQueue).toEqual([]);

    const posted = replacementWorker.posted[0] as { taskId: number };
    replacementWorker.emit('message', {
      kind: 'result',
      taskId: posted.taskId,
      result: {
        filePath: 'queued.ts',
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
      },
    });
    await expect(batch).resolves.toMatchObject([
      {
        filePath: 'queued.ts',
        cancelled: false,
      },
    ]);
  });

  test('a worker published during init can be replaced before sibling spawn settles', async () => {
    const pool = new WorkerPool({
      configs: [{ configPath: 'fake', configDirectory: 'fake' }],
      workerCount: 2,
    });
    const internals = pool as any;
    const first = new FakeWorker([]);
    const second = new FakeWorker([]);
    const replacement = new FakeWorker([]);
    const firstSlot = makeSlot(first, 0);
    const secondSlot = makeSlot(second, 1);
    const replacementSlot = makeSlot(replacement, 0);
    internals.attachOngoingHandlers(firstSlot);
    internals.attachOngoingHandlers(secondSlot);
    internals.attachOngoingHandlers(replacementSlot);

    let releaseSecond!: () => void;
    const secondGate = new Promise<void>((resolve) => {
      releaseSecond = resolve;
    });
    let markFirstPublished!: () => void;
    const firstPublished = new Promise<void>((resolve) => {
      markFirstPublished = resolve;
    });
    const workers: TestSlot[] = [];
    internals.workers = new Proxy(workers, {
      get(target, property, receiver) {
        if (property === 'push') {
          return (...values: TestSlot[]) => {
            const length = target.push(...values);
            if (values.includes(firstSlot)) markFirstPublished();
            return length;
          };
        }
        return Reflect.get(target, property, receiver);
      },
    });

    let spawnCall = 0;
    let markRespawnStarted!: () => void;
    const respawnStarted = new Promise<void>((resolve) => {
      markRespawnStarted = resolve;
    });
    internals.spawnWorker = async (id: number) => {
      spawnCall++;
      if (spawnCall === 1) return firstSlot;
      if (spawnCall === 2) {
        await secondGate;
        return secondSlot;
      }
      expect(id).toBe(0);
      markRespawnStarted();
      return replacementSlot;
    };

    const init = pool.init();
    await firstPublished;
    expect(workers).toEqual([firstSlot]);
    first.emit('exit', 42);
    await respawnStarted;
    await Promise.all([...internals.respawns]);
    expect(workers[0]).toBe(replacementSlot);

    releaseSecond();
    await init;
    expect(workers).toHaveLength(2);
    expect(workers.every((slot) => slot.ready)).toBe(true);
  });

  test('mixed slots wait for graceful exit and forced exit together', async () => {
    const firstWorker = new FakeWorker([]);
    const secondWorker = new FakeWorker([]);
    const { pool, state } = makePool(firstWorker, secondWorker);
    const capture = beginShutdownWithCapturedTimers(pool);
    try {
      expect(capture.timers.map((timer) => timer.delay)).toEqual([
        5_000, 5_000,
      ]);
      firstWorker.emit('exit', 0);
      capture.timers[1].fire();
    } finally {
      capture.restore();
    }

    await capture.shutdown;
    expect(firstWorker.terminateCalls).toBe(0);
    expect(secondWorker.terminateCalls).toBe(1);
    expect(state.workers).toEqual([]);
  });
});
