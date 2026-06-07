/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * WorkerPool: spawn N worker_threads, each preloaded with the configured
 * ESLint plugins, and dispatch lint tasks to them round-robin. Survives
 * worker crashes by respawning (capped retries) and failing the affected
 * in-flight tasks. Honors `--singleThreaded` (workerCount=1) and supports
 * cooperative cancellation via the SAB-backed cancel-flag pool.
 *
 * Lifecycle:
 *
 *   const pool = new WorkerPool({...})
 *   await pool.init()                       // spawns workers, waits all 'ready'
 *   const results = await pool.lintBatch(tasks)
 *   pool.cancel(taskId)                     // optional, mid-flight
 *   await pool.shutdown()                   // graceful drain + worker exit
 *
 * Failure isolation — each class of failure is contained at the smallest
 * scope it can be, so a per-file rule throw doesn't take down a worker
 * and a worker crash doesn't take down the pool:
 *
 *   - **Plugin import / init fails or hangs**
 *       init() rejects with the worker's reported message and
 *       terminates any surviving workers.
 *   - **Listener throws on a node**
 *       caught inside the worker (ecma-language-plugin), surfaced in
 *       result.ruleErrors; not a pool concern.
 *   - **Parse / normalize fails for one file**
 *       result.parseError set; lintBatch continues with other files.
 *   - **Per-task timeout (default 30s)**
 *       terminate the worker, respawn, mark only that task as plugin-lint-failed.
 *   - **Worker thread crashes**
 *       all in-flight tasks on that worker are rejected with
 *       {parseError: 'worker_crashed'}; pool respawns up to retryCap.
 */

import { Worker, type WorkerOptions } from 'node:worker_threads';
import { fileURLToPath } from 'node:url';
import { StringDecoder } from 'node:string_decoder';

import { CancelFlagPool } from './cancel-flag.js';
import type {
  LintFileRequest,
  LintFileResult,
} from './linter/ecma-language-plugin.js';
import type { ConfigDescriptor } from './types.js';

// ─── Configuration ─────────────────────────────────────────────────────

export interface WorkerPoolOptions {
  /**
   * Each user rslint config file passed directly to the worker; the
   * worker imports each one once at init via
   * `loadPluginsFromConfigs`, caches the resulting `LoadedPlugins` per
   * `configDirectory`, and per-file lint tasks pick the right one via
   * the task's `configKey`.
   *
   * Empty array means "no plugin work" — the pool spawns no workers
   * (`workerCount=0` fast path) and `lintBatch` short-circuits to
   * empty per-file results.
   */
  configs: ConfigDescriptor[];
  /** Worker count. 1 honors --singleThreaded; default min(cpus, 8). */
  workerCount?: number;
  /** Per-task soft deadline (ms). Default 30_000. */
  taskTimeoutMs?: number;
  /** Worker init timeout (ms). Default 60_000. */
  workerInitTimeoutMs?: number;
  /** Max worker respawns per crashed worker before giving up. Default 3. */
  retryCap?: number;
  /** Hook called when a runner-side log is received (`console.*` from plugins / pool diagnostics). */
  onLog?: (rec: {
    level: string;
    source: 'plugin' | 'runner';
    text: string;
  }) => void;
}

/** A single task as the pool sees it; equivalent to LintFileRequest minus cancelFlag (pool injects). */
export type LintTask = Omit<LintFileRequest, 'cancelFlag'>;

// ─── Internal types ────────────────────────────────────────────────────

interface WorkerSlot {
  id: number;
  worker: Worker;
  ready: boolean;
  /**
   * Set to `true` by the `'exit'` handler once the worker thread has
   * fully torn down. `shutdown()` skips waiting on these — the 'exit'
   * event already fired so `once('exit')` would never resolve and the
   * 5s `terminate()` fallback would be the only thing that frees the
   * promise.
   */
  exited: boolean;
  /**
   * `true` between the `spawnWorker(slot.id)` call in the exit handler
   * and that promise settling. `drainQueueIfAllSlotsDegraded` treats
   * a respawning slot as potentially-recoverable, so a transient
   * `ready=false` window during respawn doesn't trigger the terminal
   * `pool_degraded` drain. Without this, a multi-worker pool where
   * slot A's respawn rejects while slot B is mid-respawn would mis-
   * drain the queue even though B is about to come back online.
   */
  respawning: boolean;
  /** Tasks currently dispatched to this worker, by taskId. */
  inflight: Map<number, PendingTask>;
  /** How many times this worker has crashed and been respawned. */
  crashCount: number;
}

interface PendingTask {
  taskId: number;
  cancelSlot: number;
  resolve: (r: LintFileResult) => void;
  /**
   * Reject this in-flight task with a result-shaped failure. `kind`
   * encodes the cause so the resolver can stamp the right parseError
   * prefix:
   *
   *   - `'crash'`     → `parseError: 'worker_crashed: <msg>'`
   *                     (unexpected exit, runtime error, dropped
   *                     `message`/`error` from the worker)
   *   - `'shutdown'`  → `parseError: 'shutdown'`
   *                     (graceful pool.shutdown() while task was
   *                     in-flight — not a real crash)
   *
   * Distinguishing these matters to consumers: 'crash' triggers
   * compat-dispatch error counters / strict-runner exit codes,
   * 'shutdown' is the user/runtime expectedly tearing down the
   * pool and shouldn't be reported as a fault.
   */
  reject: (err: Error, kind?: 'crash' | 'shutdown') => void;
  /** Per-task timeout handle. */
  timer: NodeJS.Timeout;
}

// ─── Implementation ────────────────────────────────────────────────────

/**
 * Worker entry (`lint-worker.js`) path.
 *
 * Production: this module is bundled into `dist/eslint-plugin/index.js`, so the
 * sibling `./lint-worker.js` (relative to `import.meta.url`) resolves straight
 * to `dist/eslint-plugin/lint-worker.js` — nothing is read from the environment.
 *
 * Tests / dev: this runs from `src/eslint-plugin/*.ts` (rstest transforms it on
 * the fly), but `worker_threads` can't execute TypeScript — it needs the built
 * `.js`. The rstest setup file calls `setWorkerEntryForTests()` to point at the
 * rslib output. (Run `pnpm build` once before testing.)
 */
let testWorkerEntry: string | undefined;

/**
 * Test-only override for the worker entry path. The source-run tests (rstest)
 * can't load this module's `.ts` siblings in a worker thread, so the rstest
 * setup file points this at the built `dist/eslint-plugin/lint-worker.js`. Not
 * re-exported from the package entry — production never calls it and always
 * resolves the bundled sibling below.
 */
export function setWorkerEntryForTests(path: string): void {
  testWorkerEntry = path;
}

const resolveWorkerFile = (): string =>
  testWorkerEntry ??
  fileURLToPath(new URL('./lint-worker.js', import.meta.url));

/**
 * Grace period to wait for a worker to exit on its own before forcing
 * `terminate()`. Shared by `shutdown()` (after a cooperative
 * `{kind:'shutdown'}` message) and the init-error path (the worker
 * self-exits via `process.exitCode = 1`). Forcing `terminate()` only as
 * a fallback avoids racing the worker's own native / stdio teardown —
 * see the native-abort note on `terminateWorker`.
 */
const WORKER_EXIT_GRACE_MS = 5_000;

/**
 * Terminate a worker, closing its piped stdout/stderr FIRST.
 *
 * Mitigates a windows-latest-only crash: `worker.terminate()` abruptly
 * kills the thread while its stdio named pipes are still live, and
 * libuv's concurrent pipe teardown during that kill can fault BELOW the
 * JS layer (a native abort — rstest's forked test child intercepts
 * `uncaughtException` / `unhandledRejection` / `process.exit`, so the
 * only way that child can "exit unexpectedly" is a native fault). It
 * surfaced only in the high-terminate-churn `worker-pool-e2e` suite.
 * Closing our read end first makes the pipe teardown deterministic and
 * removes that race — and is sound teardown hygiene on every platform.
 * Trailing worker output is intentionally dropped (the worker is being
 * killed); `destroy()` errors are swallowed by the `ignorePipeError`
 * listeners wired in `spawnWorker`. Returns `terminate()`'s promise so
 * callers can chain `.then` / `.catch` / `.finally` unchanged.
 */
async function terminateWorker(worker: Worker): Promise<number> {
  worker.stdout?.destroy();
  worker.stderr?.destroy();
  return worker.terminate();
}

/**
 * A task that has been received by `lintBatch` but not yet posted to a
 * worker. Each `QueuedTask` gets a `taskId` + `cancelSlot` allocated
 * up-front (so callers can use `cancelTask(taskId)` before the task
 * ever reaches a worker). The pool drains this queue via `kickQueue`:
 * each idle worker takes one queued task, moves it into `slot.inflight`
 * via `dispatchToWorker`, and only THEN does its 30 s timer start —
 * so backlog wait time doesn't count against the per-task timeout.
 */
interface QueuedTask {
  task: LintTask;
  taskId: number;
  cancelSlot: number;
  resolve: (r: LintFileResult) => void;
  /** Set by `cancelTask` when the task is cancelled while still in
   *  the queue (not yet dispatched). `kickQueue` skips cancelled
   *  entries and resolves them as cancelled without ever posting to
   *  a worker. */
  cancelled: boolean;
}

export class WorkerPool {
  private readonly opts: Required<Omit<WorkerPoolOptions, 'onLog'>> & {
    onLog?: WorkerPoolOptions['onLog'];
  };
  private readonly cancelPool: CancelFlagPool;
  private workers: WorkerSlot[] = [];
  private nextTaskId = 1;
  private closed = false;
  /** Pool-level backlog. `kickQueue` moves entries to idle workers
   *  one at a time. Each worker carries at most ONE inflight task —
   *  the cap exists so per-task timeouts only measure actual
   *  execution time, not "stuck behind 30 other tasks on the worker's
   *  postMessage queue" wait time. Otherwise a 100-file batch on 8
   *  workers makes the last ~12 tasks per worker race a 30 s timer
   *  before the worker even reaches them, terminating the worker and
   *  marking everything else inflight on that worker as `worker_crashed`. */
  private pendingQueue: QueuedTask[] = [];
  /**
   * In-flight respawn promises. A crashed worker's `'exit'` handler
   * kicks off `spawnWorker(id)` asynchronously; `shutdown()` awaits
   * these so it doesn't return while a freshly-spawned replacement
   * thread is still booting (which would leave an orphan worker alive
   * past `await pool.shutdown()`). Each promise resolves only after
   * the replacement is either adopted or — when `closed` raced ahead
   * — fully terminated.
   */
  private readonly respawns = new Set<Promise<void>>();

  constructor(opts: WorkerPoolOptions) {
    const cpuCount = (() => {
      try {
        // rslint-disable-next-line @typescript-eslint/no-var-requires
        const os = require('node:os') as { cpus(): unknown[] };
        return Math.max(1, Math.min(8, os.cpus().length));
      } catch {
        return 4;
      }
    })();
    this.opts = {
      configs: opts.configs,
      // Empty `configs` ⇒ no plugin work. Force the effective worker
      // count to 0 so init() / lintBatch() / shutdown() all take their
      // no-worker fast paths, honoring the `configs` JSDoc contract
      // ("Empty array means 'no plugin work' — the pool spawns no
      // workers"). Pre-fix the default `?? cpuCount` ran even for
      // empty configs, so the pool spawned real workers and each one
      // failed init (`lint-worker.ts` rejects an empty `configs[]`),
      // turning the documented no-op into an init crash. An explicit
      // `workerCount` is intentionally ignored when there's no work —
      // a worker with zero configs has nothing to load and would
      // crash on init regardless.
      workerCount:
        opts.configs.length === 0 ? 0 : (opts.workerCount ?? cpuCount),
      taskTimeoutMs: opts.taskTimeoutMs ?? 30_000,
      workerInitTimeoutMs: opts.workerInitTimeoutMs ?? 60_000,
      retryCap: opts.retryCap ?? 3,
      onLog: opts.onLog,
    };
    // Cancel pool: lintBatch dispatches every task in a batch up-front
    // via Promise.all, so the concurrent-in-flight count equals the
    // batch size, NOT the worker count. The previous workerCount × 8
    // formula silently exhausted on batches over ~64 tasks (cancel
    // returned -1 for any task beyond capacity → cancellation no-ops).
    // CancelFlagPool starts at INITIAL_SLOTS (1024) and doubles its
    // backing SharedArrayBuffer on demand up to MAX_CAPACITY, so it
    // scales to arbitrarily large batches without an explicit size hint.
    this.cancelPool = new CancelFlagPool();
  }

  /**
   * Spawn workers and wait until all report 'ready'. Rejects on the
   * first worker init failure — the entire pool is unusable if any
   * plugin fails to load (the user's config references rules from a
   * plugin that didn't import, so every subsequent lintBatch would
   * surface "rule not found").
   *
   * `workerCount=0` is a no-op fast path used when there are zero
   * ESLint plugin entries: the Go side never reaches the dispatcher
   * (no rule has IsEslintPluginRule=true), so a real worker pool is
   * wasted overhead. We still go through the IPC handshake — this lets
   * the CLI keep ONE code path regardless of whether plugins are
   * configured.
   */
  async init(): Promise<void> {
    if (this.workers.length > 0) {
      throw new Error('WorkerPool: init called twice');
    }
    if (this.opts.workerCount === 0) {
      return;
    }
    // Launch every worker in parallel and push each ready slot into
    // `this.workers` AS its spawnWorker promise resolves. The previous
    // shape — `this.workers = fulfilled` AFTER `Promise.allSettled`
    // — opened a window where a ready worker crashing mid-init had
    // its respawn handler `findIndex(this.workers, …)` run against
    // the still-empty array, returning -1; the replacement worker
    // was then terminated (line ~754 in the exit handler) and the
    // original slot left dead. By the time `init()` finished, the
    // pool silently came up with N-1 active workers. Pushing as
    // ready closes the window: the respawn handler always sees the
    // original slot in the list.
    //
    // If even one fails to init (plugin import error / hung init),
    // the pool is unusable; allSettled lets us collect every result,
    // terminate the survivors, then surface the first failure to the
    // caller.
    const settled = await Promise.allSettled(
      Array.from({ length: this.opts.workerCount }, async (_, i) => {
        const slot = await this.spawnWorker(i);
        // If init was aborted between spawn-resolve and this await
        // (e.g. shutdown raced or another worker already failed
        // hard), don't add a now-orphaned worker to the list.
        if (this.closed) {
          terminateWorker(slot.worker).catch(() => {
            /* best-effort */
          });
          return slot;
        }
        this.workers.push(slot);
        return slot;
      }),
    );
    let hasFailure = false;
    let firstFailure: unknown = undefined;
    for (const r of settled) {
      if (r.status === 'rejected' && !hasFailure) {
        // Track failure explicitly — `firstFailure === undefined` was
        // ambiguous with `reject(undefined)` (a rejection IS a failure
        // regardless of the reason carried). Pre-fix any spawn that
        // rejected with undefined silently bypassed the failure path,
        // leaving the pool under-provisioned with no error surfaced.
        hasFailure = true;
        firstFailure = r.reason;
      }
    }
    if (hasFailure) {
      // Mark closed BEFORE terminating so the spawnWorker
      // `attachOngoingHandlers` 'exit' handler (which runs async via
      // the worker_thread exit event) sees `this.closed === true` and
      // skips its respawn branch. Without this flag flip, terminate →
      // 'exit' fires → respawn fires spawnWorker → newSlot.then runs
      // and would leak the new worker as an orphan thread.
      this.closed = true;
      // Terminate every worker that managed to enter `this.workers`.
      await Promise.allSettled(
        this.workers.map(async (w) => terminateWorker(w.worker)),
      );
      // Await any respawn already in flight, mirroring `shutdown()`.
      // If a ready worker crashed during the initial spawn window
      // (before `closed` flipped above), its exit handler registered a
      // respawn in `this.respawns`. That respawn's `.then` now sees
      // `closed === true` and terminates the freshly-spawned
      // replacement — but `init()` would otherwise `throw` before that
      // orphan thread is reaped, leaking a live worker_thread past the
      // rejected `init()`. `shutdown()` awaits the same Set for exactly
      // this reason; the init-failure path must be symmetric.
      await Promise.allSettled([...this.respawns]);
      this.workers = [];
      // Wrap non-Error rejection values into a real Error so the
      // caller's `.catch(err => err.message)` doesn't crash on
      // `undefined.message`. The pool's own spawn paths always reject
      // with Error today; this only matters if a future code path or
      // a runtime hook produces a non-Error rejection.
      throw firstFailure instanceof Error
        ? firstFailure
        : new Error(`worker spawn rejected: ${String(firstFailure)}`);
    }
    // `this.workers` was populated incrementally as each spawnWorker
    // resolved — no need for a final assignment here.
  }

  /**
   * Dispatch tasks round-robin to workers; resolve as a per-task result array.
   *
   * @param onTaskDispatched optional callback invoked synchronously with
   *   each task's internal taskId after the task has been fully tracked
   *   (cancelSlot acquired, `inflight` populated) but BEFORE the task
   *   is posted to the worker. Two use cases:
   *
   *     1. **ID bookkeeping** — callers that need a list of dispatched
   *        ids (e.g. the VS Code extension host mapping LSP
   *        `$/cancelRequest` reqId → taskId) collect them here.
   *
   *     2. **Cancel-before-start** — callers that have observed a
   *        cancel signal mid-dispatch can call `cancelTask(taskId)` from
   *        inside this callback. Because `inflight` is already populated,
   *        the lookup succeeds and the SAB cancel flag is set BEFORE
   *        postMessage delivers the task. The worker sees flag=1 on its
   *        first poll and bails immediately without running any rule.
   *
   *   The callback runs before any await, so the caller's bookkeeping
   *   is guaranteed populated before any task can complete.
   */
  async lintBatch(
    tasks: LintTask[],
    onTaskDispatched?: (taskId: number) => void,
  ): Promise<LintFileResult[]> {
    if (this.closed) throw new Error('WorkerPool: closed');
    if (this.opts.workerCount === 0) {
      // Empty pool: should not normally be invoked (Go side has no compat
      // rules to batch when entries=[]), but tolerate it gracefully so a
      // misconfigured caller doesn't crash. Each task gets a fully
      // populated empty result so downstream merge code is unchanged.
      return tasks.map((t) => ({
        filePath: t.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
      }));
    }
    if (this.workers.length === 0)
      throw new Error('WorkerPool: not initialized');

    // Enqueue every task synchronously (no awaits before this loop
    // finishes — `onTaskDispatched` runs in caller order so the
    // VS Code extension's reqId→taskId map stays consistent with the
    // input array). `cancelTask` works against in-queue entries too,
    // so callers can cancel before the task ever hits a worker.
    const promises = tasks.map(
      async (task) =>
        new Promise<LintFileResult>((resolve) => {
          const taskId = this.nextTaskId++;
          const cancelSlot = this.cancelPool.acquire();
          const q: QueuedTask = {
            task,
            taskId,
            cancelSlot,
            resolve,
            cancelled: false,
          };
          this.pendingQueue.push(q);
          onTaskDispatched?.(taskId);
        }),
    );
    // Drain as many queued tasks as there are idle workers right now.
    // Subsequent drains happen from the result/exit/respawn handlers.
    this.kickQueue();
    // If the pool has ALREADY settled into the terminal degraded state
    // (every worker thread has EXITED on a PRIOR batch — each slot's
    // respawn cap exhausted, or a respawn rejected — so none is ready,
    // respawning, or even alive), there is no future event left to
    // drain THIS freshly-enqueued batch: `kickQueue` above skipped
    // every not-ready slot, and the exit/respawn handlers that normally
    // call `drainQueueIfAllSlotsDegraded` already fired for the earlier
    // crashes and won't fire again. Without draining here the tasks sit
    // in `pendingQueue` forever and `Promise.all(promises)` never
    // settles → permanent hang.
    //
    // Gate strictly on `every slot has exited`. This is the ONLY
    // genuinely-terminal shape:
    //   - a ready/respawning slot ⇒ normal path, leave the batch to the
    //     worker / a later handler (the task's "normal path unchanged"
    //     requirement);
    //   - a slot that is merely `ready=false` but still ALIVE
    //     (`exited=false`) — e.g. a worker briefly busy, or a caller
    //     that toggled readiness — is NOT terminal: `shutdown()` /
    //     `kickQueue` will still service or drain its queued tasks, so
    //     stamping them `pool_degraded` here would be wrong (it would
    //     preempt a pending `cancelTask`/`shutdown` outcome).
    // `drainQueueIfAllSlotsDegraded`'s own guard (`some(ready ||
    // respawning)`) can't make this distinction (it ignores `exited`),
    // so the discriminator lives here at the call site; we still reuse
    // the function for its `pool_degraded` result shape rather than
    // duplicating it.
    if (this.workers.length > 0 && this.workers.every((s) => s.exited)) {
      this.drainQueueIfAllSlotsDegraded();
    }
    return Promise.all(promises);
  }

  /**
   * Move queued tasks onto idle workers. Each worker takes AT MOST ONE
   * task — the per-worker concurrency cap is what makes the per-task
   * timeout meaningful (it measures real execution time, not backlog
   * wait + execution). Called from:
   *
   *   - `lintBatch` after enqueue
   *   - the `result` message handler (`attachOngoingHandlers`) after a
   *     task completes
   *   - the `exit` handler after a respawn, so the replacement worker
   *     can pick up the backlog
   *
   * Idempotent + cheap: bails on the first non-idle worker for each
   * loop pass and on an empty queue.
   *
   * Pre-cancelled entries (set by `cancelTask` while still queued)
   * are resolved here without ever being posted to a worker.
   */
  private kickQueue(): void {
    if (this.closed) return;
    if (this.pendingQueue.length === 0) return;
    for (const slot of this.workers) {
      if (this.pendingQueue.length === 0) return;
      if (!slot.ready) continue;
      if (slot.inflight.size > 0) continue;
      // Skip any cancelled-while-queued entries to find the next real task.
      while (this.pendingQueue.length > 0) {
        const q = this.pendingQueue.shift()!;
        if (q.cancelled) {
          this.cancelPool.release(q.cancelSlot);
          q.resolve({
            filePath: q.task.filePath,
            diagnostics: [],
            fixes: [],
            suggestionsCount: 0,
            cancelled: true,
          });
          continue;
        }
        this.dispatchToWorker(slot, q);
        break;
      }
    }
  }

  /** Cancel a task by taskId. Best-effort.
   *
   *   - In-flight (already on a worker): set the SAB cancel flag —
   *     worker bails at the next per-node visit.
   *   - Queued (not yet dispatched): set `cancelled = true` so
   *     `kickQueue` resolves it as cancelled and never posts to a
   *     worker.
   *   - Otherwise (already completed / never existed): no-op, returns
   *     false.
   *
   * RACE NOTE: a `true` return does NOT guarantee the task's result is
   * suppressed. Cancellation is cooperative (polled per-node in
   * `listener-merge`), so a worker that has already finished its
   * traversal but whose result is still in flight will deliver that
   * result with `cancelled: false` and complete diagnostics — the flag
   * is no longer read. The returned result is itself correct/complete;
   * callers that need to drop a cancelled task's output must key off
   * each result's own `cancelled` field, not this method's return. */
  cancelTask(taskId: number): boolean {
    for (const w of this.workers) {
      const p = w.inflight.get(taskId);
      if (p) {
        this.cancelPool.cancel(p.cancelSlot);
        return true;
      }
    }
    for (const q of this.pendingQueue) {
      if (q.taskId === taskId) {
        q.cancelled = true;
        return true;
      }
    }
    return false;
  }

  /** Graceful shutdown — message all workers, wait for exit. */
  async shutdown(): Promise<void> {
    if (this.closed) return;
    this.closed = true;
    if (this.opts.workerCount === 0) {
      // No workers spawned — nothing to drain.
      return;
    }
    // Reject any still-pending tasks so callers don't hang. The
    // 'shutdown' kind keeps the resolver from labelling these as
    // worker crashes (which would inflate strict-runner counters and
    // mislead operators).
    const reason = new Error('WorkerPool: shutdown requested');
    for (const w of this.workers) {
      for (const [, p] of w.inflight) {
        clearTimeout(p.timer);
        // Release the slot BEFORE inflight.clear() — `w.inflight` is
        // about to be wiped, and the eventual 'exit' handler's
        // release loop iterates an empty Map and runs zero times.
        // The other cleanup paths (exit / result / timeout /
        // postMessage-fail / pendingQueue drain) all release here
        // too; without it, every shutdown-with-in-flight slowly
        // consumes the SAB slot pool until cancel becomes a no-op.
        this.cancelPool.release(p.cancelSlot);
        p.reject(reason, 'shutdown');
      }
      w.inflight.clear();
    }
    // Drain the backlog the same way: each queued task resolves with
    // `parseError: 'shutdown'` so callers awaiting `lintBatch` get a
    // result-shaped failure rather than hanging on a Promise that
    // would never settle (no worker will ever pick it up).
    for (const q of this.pendingQueue) {
      this.cancelPool.release(q.cancelSlot);
      if (q.cancelled) {
        // User called `cancelTask(taskId)` between enqueue and drain.
        // Surface it as a cancellation (matches the kickQueue path
        // for queued+cancelled entries) instead of mislabelling as a
        // shutdown failure — host strict-runner counters and LSP
        // result categorisation distinguish the two.
        q.resolve({
          filePath: q.task.filePath,
          diagnostics: [],
          fixes: [],
          suggestionsCount: 0,
          cancelled: true,
        });
      } else {
        q.resolve({
          filePath: q.task.filePath,
          diagnostics: [],
          fixes: [],
          suggestionsCount: 0,
          cancelled: false,
          parseError: 'shutdown',
        });
      }
    }
    this.pendingQueue = [];
    // Tell each worker to exit.
    for (const w of this.workers) {
      try {
        w.worker.postMessage({ kind: 'shutdown' });
      } catch {
        /* ignore */
      }
    }
    // Wait for actual termination (grace then terminate). Skip slots
    // that already fired 'exit' — for those, `once('exit')` never
    // resolves (event already happened) and the 5s `terminate()`
    // fallback would be the only thing freeing the promise, blocking
    // shutdown on every dead slot.
    const exitWaits = this.workers.map(async (w) => {
      if (w.exited) return;
      return new Promise<void>((resolveOk) => {
        const t = setTimeout(() => {
          void terminateWorker(w.worker).finally(() => {
            resolveOk();
          });
        }, WORKER_EXIT_GRACE_MS);
        w.worker.once('exit', () => {
          clearTimeout(t);
          resolveOk();
        });
      });
    });
    // Also await any in-flight respawns kicked off by a crash that
    // raced this shutdown. `this.closed` is already true, so each
    // respawn's `.then` takes the terminate-the-new-worker branch and
    // resolves only once that thread is dead — without this await,
    // `shutdown()` could return while a replacement worker was still
    // booting (running plugin imports), leaving an orphan thread that
    // keeps the process alive past `await pool.shutdown()`.
    await Promise.all([
      Promise.all(exitWaits),
      Promise.allSettled([...this.respawns]),
    ]);
    this.workers = [];
  }

  // ── private ──────────────────────────────────────────────────────

  /**
   * Spawn one worker and wait for its 'ready' (or 'init-error') message.
   * On init error or init-timeout, throws — caller (init) reports up.
   */
  private async spawnWorker(id: number): Promise<WorkerSlot> {
    return new Promise((resolveOk, reject) => {
      const workerOptions: WorkerOptions = {
        workerData: {
          cancelSab: this.cancelPool.sharedBuffer,
          configs: this.opts.configs,
        },
        // Capture worker stdout/stderr instead of letting them inherit
        // the parent's. Native console behavior inside plugin code is
        // preserved (plugins still call the unpatched `console.log`),
        // but the resulting output goes into a per-worker readable
        // stream that we forward to `onLog` below — keeping the
        // parent's stdout (the rslint CLI / LSP JSON-RPC channel)
        // pristine. The previous design monkey-patched `console.*` in
        // the worker, which altered plugin-visible behavior; this
        // approach does not.
        stdout: true,
        stderr: true,
      };
      let worker: Worker;
      try {
        worker = new Worker(resolveWorkerFile(), workerOptions);
      } catch (err) {
        reject(err as Error);
        return;
      }

      const slot: WorkerSlot = {
        id,
        worker,
        ready: false,
        exited: false,
        respawning: false,
        inflight: new Map(),
        crashCount: 0,
      };

      // Forward worker stdout / stderr to the host's `onLog`. Plugin
      // code (and any runner-internal `console.*` calls inside the
      // worker) write into these streams via the unpatched native
      // console; we expose the bytes to the host with a stable shape.
      // `source: 'plugin'` matches what the prior in-process
      // postMessage path stamped, preserving downstream
      // categorisation (VS Code OutputChannel routing etc.).
      //
      // `StringDecoder` accumulates trailing incomplete UTF-8 sequences
      // across chunk boundaries. A bare `buf.toString('utf8')` per
      // chunk would replace any byte straddling a chunk split with
      // U+FFFD (e.g. a multi-byte CJK / emoji character split between
      // two reads), corrupting plugin log output. One decoder per
      // stream because stdout/stderr are independent pipes.
      const stdoutDecoder = new StringDecoder('utf8');
      const stderrDecoder = new StringDecoder('utf8');
      const forwardStream =
        (level: 'log' | 'error', decoder: StringDecoder) => (buf: Buffer) => {
          const text = decoder.write(buf);
          if (text.length === 0) return;
          this.opts.onLog?.({
            level,
            source: 'plugin',
            text,
          });
        };
      // Flush whatever the decoder is still holding when the stream
      // closes — typically empty, but a truncated tail byte would
      // otherwise be silently dropped instead of surfacing as U+FFFD.
      const flushOnEnd =
        (level: 'log' | 'error', decoder: StringDecoder) => () => {
          const tail = decoder.end();
          if (tail.length === 0) return;
          this.opts.onLog?.({
            level,
            source: 'plugin',
            text: tail,
          });
        };
      worker.stdout?.on('data', forwardStream('log', stdoutDecoder));
      worker.stdout?.on('end', flushOnEnd('log', stdoutDecoder));
      worker.stderr?.on('data', forwardStream('error', stderrDecoder));
      worker.stderr?.on('end', flushOnEnd('error', stderrDecoder));
      // Swallow `'error'` on the stdout/stderr pipes. These are
      // best-effort log-forwarding pipes; when a worker is torn down via
      // `terminateWorker()` (init/task timeout, crash respawn, or the
      // shutdown grace fallback), Windows can destroy the pipes mid-flight
      // and emit `'error'` (EPIPE / ERR_STREAM_PREMATURE_CLOSE) where
      // macOS/Linux emit a clean `'end'`. A stream `'error'` with NO
      // listener is re-thrown by Node as an UNCAUGHT exception that would
      // kill the host, so we listen and no-op. NOTE: this handles only the
      // JS-level pipe error; it is NOT the fix for the separate
      // windows-latest *native* abort during terminate of an oxc-loaded
      // worker (see the note on `terminateWorker`).
      const ignorePipeError = (): void => {
        /* best-effort log pipe; teardown errors are expected */
      };
      worker.stdout?.on('error', ignorePipeError);
      worker.stderr?.on('error', ignorePipeError);

      const initTimer = setTimeout(() => {
        void terminateWorker(worker);
        reject(
          new Error(
            `worker ${id} init timed out after ${this.opts.workerInitTimeoutMs}ms`,
          ),
        );
      }, this.opts.workerInitTimeoutMs);

      // Spawn-phase `exit` listener. A worker that hard-exits during
      // init — plugin module-top-level `process.exit()`, OOM, segfault
      // — emits only `'exit'` (no `message`, and a clean exit doesn't
      // fire `'error'`). Without this listener the spawn promise would
      // neither resolve nor reject until the 60 s `initTimer` fired,
      // hanging `init()` and then mis-reporting the crash as an init
      // timeout. Reject immediately with the exit code instead. Removed
      // on `'ready'` so the permanent exit/respawn handler in
      // `attachOngoingHandlers` is the only one live afterward.
      const onExit = (code: number) => {
        clearTimeout(initTimer);
        reject(
          new Error(
            `worker ${id} exited during init (code=${code}) before sending ready`,
          ),
        );
      };
      const onMessage = (msg: { kind: string; [k: string]: unknown }) => {
        if (msg.kind === 'ready') {
          clearTimeout(initTimer);
          slot.ready = true;
          worker.off('message', onMessage);
          worker.off('exit', onExit);
          worker.off('error', onError);
          this.attachOngoingHandlers(slot);
          resolveOk(slot);
        } else if (msg.kind === 'init-error') {
          clearTimeout(initTimer);
          worker.off('message', onMessage);
          worker.off('exit', onExit);
          // Keep `onError` attached. An UNHANDLED 'error' on a Worker is
          // re-thrown by Node as an uncaught exception in the host, so a
          // teardown fault after init-error (e.g. a plugin import's
          // floating rejection firing once the worker keeps running) would
          // crash the host. onError firing late here is a harmless no-op —
          // we've already rejected — but it keeps the error safety net up.
          // (The 'ready' path can drop onError because
          // `attachOngoingHandlers` installs a permanent replacement; this
          // path has none, so we must keep it.)
          //
          // Do NOT force-terminate here. `sendInitError` (lint-worker.ts)
          // has already set the worker's `process.exitCode = 1` and
          // dropped its inbound listener, so the thread is winding itself
          // down. Calling `terminate()` on a worker that is concurrently
          // tearing down its own native (oxc) bindings + piped stdio is a
          // prime trigger for the windows-latest native abort. Reject
          // immediately (init failed; message preserved), let the worker
          // exit on its own, and fall back to terminate() only if it
          // somehow doesn't exit within the grace window. `.unref()` the
          // fallback so it never holds the host's event loop open — a
          // stuck worker dies with the process anyway.
          const graceTimer = setTimeout(() => {
            void terminateWorker(worker);
          }, WORKER_EXIT_GRACE_MS);
          graceTimer.unref();
          worker.once('exit', () => {
            clearTimeout(graceTimer);
          });
          reject(new Error(`worker ${id} init failed: ${String(msg.message)}`));
        } else {
          // Unexpected pre-ready message — ignore.
        }
      };
      const onError = (err: Error) => {
        clearTimeout(initTimer);
        worker.off('exit', onExit);
        reject(new Error(`worker ${id} error during init: ${err.message}`));
      };
      worker.on('message', onMessage);
      worker.once('error', onError);
      worker.once('exit', onExit);
    });
  }

  /**
   * Wire post-init handlers: result routing, log-forwarding, and crash recovery.
   */
  private attachOngoingHandlers(slot: WorkerSlot): void {
    slot.worker.on('message', (msg: { kind: string; [k: string]: unknown }) => {
      // Note: the previous `kind: 'log'` branch is gone — plugin /
      // runner-internal `console.*` output now flows through the
      // worker's stdout/stderr streams (wired in `spawnWorker`), not
      // through postMessage.
      if (msg.kind === 'result') {
        const taskId = msg.taskId as number;
        const p = slot.inflight.get(taskId);
        // `if (!p) return` covers the late-result race: after a task
        // timeout the entry is deleted and the worker terminated, but
        // the worker can still deliver a pending result message before
        // its 'exit' event fires. `nextTaskId` is monotonic per-pool,
        // so the lookup is guaranteed to return either the original
        // (still-pending) PendingTask or undefined — never some other
        // task that happens to share the id.
        if (!p) return;
        slot.inflight.delete(taskId);
        clearTimeout(p.timer);
        this.cancelPool.release(p.cancelSlot);
        p.resolve(msg.result as LintFileResult);
        // Worker is now idle — pull the next queued task onto it.
        // Without this kick a 100-task batch on 8 workers would never
        // progress past the first 8 tasks: each worker dispatches once
        // (from lintBatch), completes, then sits idle while 92 tasks
        // wait forever in the queue.
        this.kickQueue();
      }
    });

    slot.worker.on('error', (err) => {
      this.opts.onLog?.({
        level: 'error',
        source: 'runner',
        text: `worker ${slot.id} error: ${err.message}`,
      });
    });

    slot.worker.on('exit', (code) => {
      // Reject all in-flight tasks on this worker.
      for (const [, p] of slot.inflight) {
        clearTimeout(p.timer);
        this.cancelPool.release(p.cancelSlot);
        p.reject(new Error(`worker ${slot.id} exited with code ${code}`));
      }
      slot.inflight.clear();
      slot.ready = false;
      slot.exited = true;

      // Respawn if we haven't exhausted the retry cap and we're still alive.
      if (!this.closed && slot.crashCount < this.opts.retryCap) {
        slot.crashCount++;
        slot.respawning = true;
        this.opts.onLog?.({
          level: 'warn',
          source: 'runner',
          text: `worker ${slot.id} exited unexpectedly (code=${code}); respawning (try ${slot.crashCount}/${this.opts.retryCap})`,
        });
        // Replace this slot with a fresh worker. We re-use init's logic.
        //
        // Race window: this.closed can flip to true BETWEEN the check
        // above and the .then callback below (shutdown() called while
        // spawnWorker is in flight). The .then callback must re-check
        // and tear down the freshly-built worker if we've already
        // shut down — otherwise the new worker_threads leak past
        // shutdown() returning, keeping the Node process alive. The
        // resulting promise is registered in `this.respawns` so
        // shutdown() can AWAIT this teardown rather than returning
        // mid-boot.
        const respawnP = this.spawnWorker(slot.id).then(
          (newSlot): void | Promise<void> => {
            slot.respawning = false;
            if (this.closed) {
              // Shutdown raced ahead. Terminate the new worker and
              // RETURN the terminate promise so an awaiting shutdown()
              // (via `this.respawns`) actually waits for the thread to
              // die instead of resolving while it's still booting.
              return terminateWorker(newSlot.worker).then(
                () => {
                  /* terminated */
                },
                () => {
                  /* best-effort */
                },
              );
            }
            // Preserve crash count across replacement so the cap actually caps.
            newSlot.crashCount = slot.crashCount;
            // Replace in the workers list at the same index (id position).
            const idx = this.workers.findIndex((w) => w.id === slot.id);
            if (idx >= 0) {
              this.workers[idx] = newSlot;
            } else {
              // The original slot is no longer in `this.workers` —
              // either init() failed before assigning (workers === [])
              // or shutdown moved on. Defense in depth alongside the
              // `closed = true` flip in init's failure path: even if
              // some future caller forgets to flip `closed`, this
              // branch makes sure the freshly spawned worker isn't
              // stranded as an orphan thread.
              terminateWorker(newSlot.worker).catch(() => {
                /* best-effort */
              });
              return;
            }
            // Replacement is ready and idle — let it pick up queued
            // tasks that piled up while the original was respawning.
            this.kickQueue();
          },
          (err) => {
            slot.respawning = false;
            this.opts.onLog?.({
              level: 'error',
              source: 'runner',
              text: `worker ${slot.id} respawn failed: ${(err as Error).message}`,
            });
            // Respawn rejected: slot stays `ready=false` from the exit
            // handler above and there is no further event that could
            // resurrect it. If this leaves the pool with zero ready
            // AND zero respawning slots AND `pendingQueue` non-empty
            // (sibling tasks that accumulated while this slot's
            // hang/crash held the worker), drain them with a
            // `pool_degraded` sentinel so `await pool.lintBatch()`
            // returns instead of hanging forever — same idempotent
            // drain used by the retry-cap-exhausted branch below.
            this.drainQueueIfAllSlotsDegraded();
          },
        );
        // Track the in-flight respawn so shutdown() can await it (and
        // the closed-branch teardown above) instead of returning while
        // the replacement thread is still booting.
        this.respawns.add(respawnP);
        void respawnP.finally(() => this.respawns.delete(respawnP));
      } else if (slot.crashCount >= this.opts.retryCap) {
        this.opts.onLog?.({
          level: 'error',
          source: 'runner',
          text: `worker ${slot.id} respawn cap reached (${this.opts.retryCap}); pool degraded`,
        });
        // If THIS exit was the last one keeping the pool servicing
        // tasks, drain `pendingQueue` so the caller's
        // `await pool.lintBatch(tasks)` actually returns instead of
        // hanging forever. `kickQueue` skips slots with `ready=false`;
        // when every slot is dead AND the cap is exhausted on this
        // one, no future event can resurrect the queue, so the only
        // safe move is to surface a result-shaped failure to every
        // pending task.
        this.drainQueueIfAllSlotsDegraded();
      }
    });
  }

  /**
   * Last-chance drain for the "every worker exhausted its respawn
   * cap" terminal state. Iterates `pendingQueue`, releases each
   * cancel-slot, and resolves with `parseError: 'pool_degraded'` so
   * the host distinguishes this case from a normal `shutdown` /
   * `worker_crashed` per-task failure. Idempotent: if `pendingQueue`
   * is already empty, the loop is a no-op.
   */
  private drainQueueIfAllSlotsDegraded(): void {
    if (this.closed) return;
    // A slot is "potentially recoverable" if it's either currently
    // serving (`ready`) or in the middle of a respawn that might
    // succeed (`respawning`). Only when EVERY slot is terminally dead
    // is the pool unable to make progress and the queue safe to drain
    // with `pool_degraded`. Pre-fix the check was `some(s => s.ready)`,
    // which treated a mid-respawn slot identically to a permanently
    // dead one — so a multi-worker pool where slot A's respawn
    // rejected while slot B was still spawning would mis-drain the
    // queue even though B was about to come back.
    if (this.workers.some((s) => s.ready || s.respawning)) return;
    if (this.pendingQueue.length === 0) return;
    for (const q of this.pendingQueue) {
      this.cancelPool.release(q.cancelSlot);
      q.resolve({
        filePath: q.task.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError: 'pool_degraded',
      });
    }
    this.pendingQueue = [];
  }

  /**
   * Post one queued task to a specific (idle) worker slot. Caller is
   * `kickQueue`, which has already verified the slot is ready +
   * idle. taskId / cancelSlot were allocated at `lintBatch` enqueue
   * time, so cancellation works against queued AND inflight tasks.
   *
   * The per-task timeout starts HERE — when the worker actually
   * takes the task off the queue — not at enqueue time. That's the
   * whole point of the queue model: the previous design ran every
   * task's timer from `lintBatch` time, so the last task on each
   * worker (sitting in the worker's postMessage backlog) raced a
   * 30 s deadline before the worker even reached it.
   */
  private dispatchToWorker(slot: WorkerSlot, q: QueuedTask): void {
    const { task, taskId, cancelSlot, resolve: resolveOk } = q;

    const timer = setTimeout(() => {
      // Per-task timeout fired: terminate the worker (its exit
      // handler triggers respawn) and fail just this task. Under the
      // queue model each worker has concurrency cap = 1, so there
      // are no sibling inflight tasks on this worker to reject.
      slot.inflight.delete(taskId);
      this.cancelPool.release(cancelSlot);
      // Mark the slot not-ready SYNCHRONOUSLY before yielding. The
      // exit handler (which sets ready=false during real cleanup)
      // runs async on the worker's 'exit' event — until that fires,
      // kickQueue would see this still-alive-looking slot, hand it
      // the next queued task, and the postMessage would then fail
      // because the worker is mid-terminate. The cascade would
      // mis-tag every task post-terminate as `postMessage_failed`
      // instead of `worker_crashed`. Setting ready=false here
      // closes the race window.
      slot.ready = false;
      // Also mark the slot respawning SYNCHRONOUSLY when the exit
      // handler will actually respawn it. Between this `terminate()`
      // and the async `'exit'` handler (which is what flips
      // `respawning = true`), the slot is {ready:false,
      // respawning:false}; if a sibling slot's respawn rejects in that
      // window, `drainQueueIfAllSlotsDegraded` (`some(s => s.ready ||
      // s.respawning)`) would mis-classify this slot as terminally dead
      // and wrongly drain the pendingQueue with `pool_degraded`. Gate
      // IDENTICALLY to the exit handler's respawn decision: that handler
      // respawns iff `crashCount < retryCap` (reading crashCount BEFORE
      // it increments), so we read the same pre-increment value here.
      // An unconditional `true` would leave `respawning` stuck true once
      // the cap is hit (the exit handler then takes the no-respawn
      // branch and never clears it) → the queue would never drain →
      // hang. Only set it when a respawn is actually coming.
      if (slot.crashCount < this.opts.retryCap) {
        slot.respawning = true;
      }
      this.opts.onLog?.({
        level: 'warn',
        source: 'runner',
        text: `task ${taskId} on worker ${slot.id} timed out after ${this.opts.taskTimeoutMs}ms; terminating worker`,
      });
      // Reject this task with a "plugin-lint-failed"-shaped result.
      resolveOk({
        filePath: task.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError: 'task_timeout',
      });
      // Force termination — exit handler respawns.
      void terminateWorker(slot.worker);
    }, this.opts.taskTimeoutMs);

    const pending: PendingTask = {
      taskId,
      cancelSlot,
      resolve: resolveOk,
      reject: (err, kind = 'crash') => {
        // Unify rejection shape with timeout: prefer a result-shaped
        // failure over a thrown error so the caller (internal/linter)
        // marks this file as plugin-lint-failed instead of letting one
        // file's crash abort the whole batch.
        //
        // `parseError` prefix encodes the kind so consumers can
        // distinguish: a 'shutdown' is not a crash — counting it
        // toward strict-runner thresholds would be wrong.
        const parseError =
          kind === 'shutdown' ? 'shutdown' : `worker_crashed: ${err.message}`;
        resolveOk({
          filePath: task.filePath,
          diagnostics: [],
          fixes: [],
          suggestionsCount: 0,
          cancelled: false,
          parseError,
        });
      },
      timer,
    };
    slot.inflight.set(taskId, pending);

    // postMessage can throw before the message is dispatched:
    //   - structuredClone failure (unserializable field in `task`)
    //   - worker died between kickQueue selection and now (race)
    //   - allocator pressure / Node bug
    //
    // Without a catch, the throw escapes the synchronous caller (the
    // result is a thrown sync error rather than a resolved Promise),
    // but the bookkeeping above (timer scheduled, cancelSlot acquired,
    // inflight.set) is never undone. The timer would later fire, find
    // a taskId still on this slot, and attempt to
    // `slot.worker.terminate()` on a possibly-already-replaced slot —
    // at best wasted work, at worst spurious termination of an
    // innocent successor.
    try {
      slot.worker.postMessage({
        kind: 'task',
        taskId,
        cancelSlot,
        request: task,
      });
    } catch (err) {
      clearTimeout(timer);
      slot.inflight.delete(taskId);
      this.cancelPool.release(cancelSlot);
      // postMessage throws in TWO distinct scenarios that look the
      // same at this catch site but are very different operationally:
      //
      //   1. `DataCloneError` — the SPECIFIC `task` payload contains
      //      something structured-clone can't transfer (a function
      //      on rule.meta, a circular ref, etc.). The worker thread
      //      itself is alive and healthy; only THIS task is poison.
      //      The next queued task is a normal LintTask and will
      //      postMessage fine. Keep `ready = true` so the slot keeps
      //      draining the queue.
      //
      //   2. `Worker is not running` (and any other catch reason) —
      //      the worker really did die between pickup and postMessage
      //      (rare race). Set `ready = false` so kickQueue stops
      //      dispatching to this dead slot synchronously. Without
      //      this, the synchronous `kickQueue()` below would
      //      immediately re-dispatch the next pending task to the
      //      same dead slot, postMessage would throw again, the
      //      catch would recurse — every queued task on this slot's
      //      next round-robin pass would be mis-tagged
      //      `postMessage_failed` instead of `worker_crashed` (which
      //      is what the eventual 'exit' handler stamps).
      const errName = (err as Error).name;
      if (errName !== 'DataCloneError') {
        slot.ready = false;
      }
      resolveOk({
        filePath: task.filePath,
        diagnostics: [],
        fixes: [],
        suggestionsCount: 0,
        cancelled: false,
        parseError: `postMessage_failed: ${(err as Error).message}`,
      });
      // Re-drain the queue, DEFERRED via `queueMicrotask` rather than a
      // direct call (#2). For DataCloneError the slot stays ready and
      // kickQueue re-dispatches the next queued task here; if that task is
      // ALSO non-cloneable a synchronous call recurses
      // (kickQueue → dispatch → catch → kickQueue → …) and overflows the
      // stack on a batch of many poison tasks — turning per-file
      // degradation into a batch-wide `RangeError`. The microtask hop gives
      // each re-dispatch a fresh stack frame, so every poison file degrades
      // to `postMessage_failed` independently. For the dead-worker case
      // ready=false makes kickQueue skip this slot until 'exit' + respawn.
      queueMicrotask(() => {
        this.kickQueue();
      });
    }
  }
}

// require for cpus() — in ESM context we use createRequire.
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
