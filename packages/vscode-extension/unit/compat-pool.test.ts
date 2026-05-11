/**
 * Race-condition regression tests for CompatPool's pool lifecycle.
 *
 * Two specific races we guard against:
 *
 *   B1 — Concurrent lintBatch during cold init: two callers see
 *        `pool === null`, both try to spawn, the second overwrites
 *        the first's reference. The lost first pool leaks
 *        worker_threads. Fix: serialize init through opChain so the
 *        second caller waits for the first's init to publish to
 *        `this.pool`.
 *
 *   B2 — reconfigure-vs-lintBatch: while reconfigure awaits the old
 *        pool's shutdown, currentFlatEntries must already point at
 *        the new entries so a concurrent lintBatch (which re-enters
 *        through opChain after the reconfigure's chained block
 *        finishes) spawns the new pool against the new entries.
 *
 * The tests use an injected stub WorkerPool — CompatPool's
 * `runnerModule` constructor option — so the lifecycle can be
 * exercised deterministically without spawning real worker_threads.
 */

import { describe, test, expect } from '@rstest/core';

import os from 'node:os';
import path from 'node:path';
import { existsSync, mkdirSync, writeFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';

import type {
  LintTask,
  LintFileResult,
  WorkerPool as WorkerPoolType,
  WorkerPoolOptions,
} from '@rslint/eslint-plugin-runner';
import { WorkerClosedError } from '@rslint/eslint-plugin-runner';
import type { CancellationToken } from 'vscode';

import { CompatPool, type LintCompatBatchParams } from '../src/CompatPool';

// ── Stub infrastructure ───────────────────────────────────────────

interface StubPoolRecord {
  opts: WorkerPoolOptions;
  initCalls: number;
  initBlocker: { resolve: () => void; promise: Promise<void> };
  lintCalls: number;
  shutdownCalls: number;
  cancelled: number[];
}

function makePending<T>(): { resolve: (v: T) => void; promise: Promise<T> } {
  let resolve!: (v: T) => void;
  const promise = new Promise<T>((r) => {
    resolve = r;
  });
  return { resolve, promise };
}

function makeStubRunner(): {
  module: { WorkerPool: new (opts: WorkerPoolOptions) => WorkerPoolType };
  pools: StubPoolRecord[];
} {
  const pools: StubPoolRecord[] = [];

  class StubWorkerPool {
    private record: StubPoolRecord;
    constructor(opts: WorkerPoolOptions) {
      this.record = {
        opts,
        initCalls: 0,
        initBlocker: makePending<void>(),
        lintCalls: 0,
        shutdownCalls: 0,
        cancelled: [],
      };
      pools.push(this.record);
    }
    async init(): Promise<void> {
      this.record.initCalls++;
      // Block until the test releases this pool's init. Lets us pin
      // CompatPool inside the lazy-init `await` step deterministically.
      await this.record.initBlocker.promise;
    }
    async lintBatch(
      tasks: LintTask[],
      onTaskDispatched?: (taskId: number) => void,
    ): Promise<LintFileResult[]> {
      this.record.lintCalls++;
      // Synthesize a result per task to satisfy the cardinality
      // invariant CompatPool relies on downstream.
      return tasks.map((t, i) => {
        onTaskDispatched?.(1000 + i);
        return {
          filePath: t.filePath,
          diagnostics: [],
          fixes: [],
          suggestionsCount: 0,
          cancelled: false,
        };
      });
    }
    cancelTask(taskId: number): boolean {
      this.record.cancelled.push(taskId);
      return true;
    }
    async shutdown(): Promise<void> {
      this.record.shutdownCalls++;
    }
  }

  return {
    module: {
      WorkerPool: StubWorkerPool as unknown as new (
        opts: WorkerPoolOptions,
      ) => WorkerPoolType,
    },
    pools,
  };
}

const noopToken = {
  isCancellationRequested: false,
  onCancellationRequested: () => ({ dispose: () => {} }),
} as unknown as CancellationToken;

const stubLogger = {
  setLogLevel: () => {},
  useDefaultLogLevel: () => stubLogger,
  trace: () => {},
  debug: () => {},
  info: () => {},
  warn: () => {},
  error: () => {},
} as unknown as ConstructorParameters<typeof CompatPool>[0];

// Fixture: a real on-disk rslint.config.mjs per test, so
// `extractConfigDescriptors` keeps it (configPath has to be a real
// path) and `fingerprintConfigs`'s statSync succeeds. We rotate paths
// so each test gets a fresh fingerprint without depending on
// fs.statSync mtime granularity.
const FIXTURE_ROOT = path.join(
  os.tmpdir(),
  `rslint-compatpool-test-${process.pid}`,
);

function ensureFixtureDir(suffix: string): {
  dir: string;
  configPath: string;
} {
  const dir = path.join(FIXTURE_ROOT, suffix);
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  const configPath = path.join(dir, 'rslint.config.mjs');
  // Content is irrelevant — CompatPool.reconfigure only reads
  // configPath / configDirectory / entries metadata. We just need a
  // file to exist for statSync.
  if (!existsSync(configPath)) {
    writeFileSync(configPath, 'export default [];\n');
  }
  return { dir, configPath };
}

interface CfgInput {
  dirSuffix: string;
  pluginPrefixes?: string[];
}

function cfg(input: CfgInput): {
  configDirectory: string;
  configPath: string;
  entries: Array<{
    eslintPlugins?: Array<{ prefix: string; ruleNames: string[] }>;
  }>;
} {
  const { dir, configPath } = ensureFixtureDir(input.dirSuffix);
  return {
    configDirectory: pathToFileURL(dir).toString(),
    configPath,
    entries: input.pluginPrefixes
      ? [
          {
            eslintPlugins: input.pluginPrefixes.map((p) => ({
              prefix: p,
              ruleNames: [],
            })),
          },
        ]
      : [],
  };
}

function batch(filePath: string): LintCompatBatchParams {
  return {
    files: [{ path: filePath, text: '' }],
    rules: {},
  };
}

// ── Tests ─────────────────────────────────────────────────────────

describe('CompatPool race conditions', () => {
  test('concurrent lintBatch during cold init spawns exactly one pool (B1)', async () => {
    const { module, pools } = makeStubRunner();
    const cp = new CompatPool(stubLogger, { runnerModule: module });

    await cp.reconfigure(
      [cfg({ dirSuffix: 'b1', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );

    // Fire two lintBatch calls in parallel before init finishes.
    const lint1 = cp.lintBatch(batch('/a.ts'), noopToken);
    const lint2 = cp.lintBatch(batch('/b.ts'), noopToken);

    // Give microtasks a chance to schedule both init paths.
    await new Promise((r) => setTimeout(r, 0));

    // CRITICAL: only one pool should have been constructed. The
    // opChain serialization ensures the second lintBatch waits for
    // the first's init (via the chain) and then re-uses `this.pool`.
    expect(pools).toHaveLength(1);
    expect(pools[0].initCalls).toBe(1);

    // Release the (single) pool's init.
    pools[0].initBlocker.resolve();
    const [r1, r2] = await Promise.all([lint1, lint2]);
    expect(r1.results).toHaveLength(1);
    expect(r2.results).toHaveLength(1);
    expect(pools[0].lintCalls).toBe(2);
  });

  test('reconfigure during in-flight lintBatch routes next batch to the new entries (B2)', async () => {
    const { module, pools } = makeStubRunner();
    const cp = new CompatPool(stubLogger, { runnerModule: module });

    // Start with plugin X (config under b2-x/).
    await cp.reconfigure(
      [cfg({ dirSuffix: 'b2-x', pluginPrefixes: ['x'] })],
      'file:///proj',
    );

    // Kick off first lintBatch — pool 0 spawns and gets pinned in init.
    const lint1 = cp.lintBatch(batch('/a.ts'), noopToken);
    await new Promise((r) => setTimeout(r, 0));
    expect(pools).toHaveLength(1);

    // CRITICAL race setup: reconfigure runs WHILE lint1 is still
    // blocked in init. The previous version awaited lint1 before
    // reconfigure, which removed the race and made the B2 invariant
    // (state-update-before-await-shutdown) untestable. With the
    // current ordering, lint1 is mid-init in opChain when
    // reconfigure enqueues, so the chain serializes them and the
    // state-mutation ordering inside reconfigure (CompatPool.ts:236-256)
    // actually matters.
    const yCfg = cfg({ dirSuffix: 'b2-y', pluginPrefixes: ['y'] });
    const reconfigP = cp.reconfigure([yCfg], 'file:///proj');

    // Allow microtasks to advance — reconfigure is now ENQUEUED behind
    // lint1's init in opChain. It has NOT mutated state yet.
    await new Promise((r) => setTimeout(r, 0));

    // Release lint1's init. lint1 proceeds to dispatch its tasks on
    // pool 0 (whose configs still describe X). Then opChain drains to
    // reconfigure, which swaps state to yCfg and awaits pool 0's
    // shutdown.
    pools[0].initBlocker.resolve();
    await lint1;
    await reconfigP;

    // pool 0 was drained.
    expect(pools[0].shutdownCalls).toBe(1);

    // Issue a new lintBatch. It MUST spawn pool 1 with the NEW
    // configs — verifying that reconfigure committed the new state
    // before relinquishing the chain.
    const lint2 = cp.lintBatch(batch('/b.ts'), noopToken);
    await new Promise((r) => setTimeout(r, 0));
    expect(pools).toHaveLength(2);
    expect(pools[1].opts.configs).toHaveLength(1);
    // The discriminator: pool 1 must describe yCfg, not xCfg. If a
    // regression reordered reconfigure to write state AFTER awaiting
    // shutdown, lint2 (enqueued while shutdown was pending) could
    // still see the old configs.
    expect(pools[1].opts.configs![0].configPath).toBe(yCfg.configPath);

    pools[1].initBlocker.resolve();
    await lint2;
  });

  test('reconfigure to tier-3 (empty) drains pool and short-circuits subsequent lintBatch', async () => {
    const { module, pools } = makeStubRunner();
    const cp = new CompatPool(stubLogger, { runnerModule: module });

    await cp.reconfigure(
      [cfg({ dirSuffix: 'tier3', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );
    const lint1 = cp.lintBatch(batch('/a.ts'), noopToken);
    await new Promise((r) => setTimeout(r, 0));
    pools[0].initBlocker.resolve();
    await lint1;

    // Reconfigure to empty (user removed eslintPlugins from config).
    await cp.reconfigure([], 'file:///proj');
    expect(pools[0].shutdownCalls).toBe(1);

    // Subsequent lintBatch should short-circuit to empty results
    // WITHOUT spawning a new pool.
    const result = await cp.lintBatch(batch('/b.ts'), noopToken);
    expect(pools).toHaveLength(1); // still just pool 0, drained
    expect(result.results).toHaveLength(1);
    expect(result.results[0].diagnostics).toHaveLength(0);
  });

  test('lintBatch with already-cancelled token cancels every dispatched task', async () => {
    // Regression: VS Code's `onCancellationRequested` invokes the
    // listener IMMEDIATELY if the token is already cancelled when the
    // handler is registered. Pre-fix, the listener iterated an empty
    // `dispatchedIds`, fired no cancelTask, and tasks dispatched after
    // it (by pool.lintBatch's per-task dispatch loop) were never
    // cancelled — the user's cancel intent leaked through.
    //
    // Fixed by setting `cancelled = true` in the listener and also
    // calling cancelTask from inside the `onTaskDispatched` callback
    // when `cancelled` is already set. This test exercises that path
    // by feeding lintBatch a multi-file batch with a pre-cancelled
    // token; every task's taskId must show up in the stub's
    // `cancelled` array.
    const { module, pools } = makeStubRunner();
    const cp = new CompatPool(stubLogger, { runnerModule: module });

    await cp.reconfigure(
      [cfg({ dirSuffix: 'cancel', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );

    // Pre-cancelled token: state=true, listener fires synchronously
    // on register.
    let registeredListener: (() => void) | undefined;
    const preCancelledToken = {
      isCancellationRequested: true,
      onCancellationRequested: (cb: () => void) => {
        // VS Code semantics: invoke immediately when already cancelled,
        // and also return a disposable subscription handle.
        cb();
        registeredListener = cb;
        return { dispose: () => {} };
      },
    } as unknown as CancellationToken;

    const multiFileBatch: LintCompatBatchParams = {
      files: [
        { path: '/a.ts', text: '' },
        { path: '/b.ts', text: '' },
        { path: '/c.ts', text: '' },
      ],
      rules: {},
    };

    const lintP = cp.lintBatch(multiFileBatch, preCancelledToken);
    await new Promise((r) => setTimeout(r, 0));
    pools[0].initBlocker.resolve();
    await lintP;

    // The stub's lintBatch calls onTaskDispatched once per file with
    // ids 1000, 1001, 1002. All three MUST appear in the cancelled
    // list — proving the in-callback `if (cancelled) cancelTask()`
    // path catches every task dispatched after the listener fired.
    expect(pools[0].cancelled.sort((a, b) => a - b)).toEqual([
      1000, 1001, 1002,
    ]);
    // The listener was registered (sanity — confirms our token stub
    // mirrors the real VS Code shape).
    expect(registeredListener).toBeDefined();
  });

  test('dispose drains the live pool', async () => {
    const { module, pools } = makeStubRunner();
    const cp = new CompatPool(stubLogger, { runnerModule: module });

    await cp.reconfigure(
      [cfg({ dirSuffix: 'dispose', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );
    const lint1 = cp.lintBatch(batch('/a.ts'), noopToken);
    await new Promise((r) => setTimeout(r, 0));
    pools[0].initBlocker.resolve();
    await lint1;

    await cp.dispose();
    expect(pools[0].shutdownCalls).toBe(1);

    // dispose is idempotent — second call is a no-op (no live pool).
    await cp.dispose();
    expect(pools[0].shutdownCalls).toBe(1);
  });

  // #11: lintBatch runs OUTSIDE the serialization chain, so a concurrent
  // reconfigure/dispose can close the pool between acquisition and the
  // lintBatch call — WorkerPool/IpcClient then reject with the stable code
  // ERR_RSLINT_WORKER_CLOSED. That narrow race must degrade to empty
  // results (detected by CODE, not message text), not fail the whole batch.
  test('lintBatch tolerates a concurrently-closed pool (#11)', async () => {
    class ClosedOnLintPool {
      async init(): Promise<void> {}
      async lintBatch(): Promise<never> {
        // Message intentionally does NOT contain "WorkerPool: closed" —
        // detection relies on the WorkerClosedError type (its stable code),
        // not the message text. (If CompatPool regressed to
        // message-matching, this test would fail.)
        throw new WorkerClosedError('pool gone');
      }
      cancelTask(): boolean {
        return true;
      }
      async shutdown(): Promise<void> {}
    }
    const cp = new CompatPool(stubLogger, {
      runnerModule: {
        WorkerPool: ClosedOnLintPool as unknown as new (
          opts: WorkerPoolOptions,
        ) => WorkerPoolType,
      },
    });
    await cp.reconfigure(
      [cfg({ dirSuffix: 'closed-race', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );

    const result = await cp.lintBatch(batch('/a.ts'), noopToken);

    // Graceful: one empty result per file, NOT a rejected promise.
    expect(result.results).toHaveLength(1);
    expect(result.results[0].filePath).toBe('/a.ts');
    expect(result.results[0].diagnostics).toEqual([]);
  });

  // M12: a structurally broken plugin import causes WorkerPool.init to
  // reject. Without a circuit breaker, every subsequent lintBatch would
  // re-thrash the spawn + ESM import cycle. With the gate, repeated
  // batches against the SAME fingerprint short-circuit to empty
  // results. A new reconfigure to different configs clears the gate.
  test('init failure short-circuits subsequent lintBatch until reconfigure changes', async () => {
    const initErr = new Error('boom: plugin import failed');
    // Use a custom stub that fails init the first time.
    let initCalls = 0;
    let shutdownCalls = 0;
    const module = {
      WorkerPool: class {
        async init() {
          initCalls++;
          throw initErr;
        }
        async lintBatch() {
          // Should never reach here in this test — init failure
          // means the pool never publishes to this.pool.
          throw new Error('lintBatch invoked after failed init');
        }
        cancelTask() {
          return false;
        }
        async shutdown() {
          shutdownCalls++;
        }
      } as unknown as new (opts: WorkerPoolOptions) => WorkerPoolType,
    };

    const cp = new CompatPool(stubLogger, { runnerModule: module });
    await cp.reconfigure(
      [cfg({ dirSuffix: 'm12-bad', pluginPrefixes: ['bad'] })],
      'file:///proj',
    );

    // First lintBatch — init fires and throws. CompatPool surfaces the
    // throw to the caller; check that initCalls bumped.
    let first: unknown;
    try {
      await cp.lintBatch(batch('/a.ts'), noopToken);
    } catch (err) {
      first = err;
    }
    expect(first).toBeDefined();
    expect((first as Error).message).toMatch(/boom: plugin import failed/);
    expect(initCalls).toBe(1);

    // Second lintBatch — SAME configs → circuit-breaker short-circuits.
    // No new init attempt; result is the empty-pool fast-path.
    const second = await cp.lintBatch(batch('/b.ts'), noopToken);
    expect(initCalls).toBe(1); // unchanged — gate prevented retry
    expect(second.results).toHaveLength(1);
    expect(second.results[0].diagnostics).toHaveLength(0);
    expect(second.results[0].filePath).toBe('/b.ts');

    // Reconfigure to a DIFFERENT fingerprint clears the gate; next
    // lintBatch tries init again (still fails, but the retry IS made).
    await cp.reconfigure(
      [cfg({ dirSuffix: 'm12-other', pluginPrefixes: ['other'] })],
      'file:///proj',
    );
    let third: unknown;
    try {
      await cp.lintBatch(batch('/c.ts'), noopToken);
    } catch (err) {
      third = err;
    }
    expect(third).toBeDefined();
    expect(initCalls).toBe(2); // retry was attempted

    // No pool was ever published, so dispose has nothing to shut down.
    await cp.dispose();
    expect(shutdownCalls).toBe(0);
  });

  // #8: the breaker resets on a config-fingerprint change, but a dependency
  // install (npm install of the missing plugin) fixes a broken import
  // WITHOUT touching the config — so a fingerprint-only reset never fires
  // and the pool would stay stuck forever. Recovery is TIME-BASED: after
  // RETRY_BACKOFF_MS the breaker allows one retry. We drive the injected
  // clock to assert short-circuit within the window, retry after it.
  test('init-failure breaker retries after the backoff window elapses (#8)', async () => {
    let clock = 0;
    let initCalls = 0;
    const module = {
      WorkerPool: class {
        async init() {
          initCalls++;
          if (initCalls === 1) throw new Error('boom: plugin import failed');
          // 2nd init (after the gate reset) succeeds.
        }
        async lintBatch(tasks: LintTask[]) {
          return tasks.map((t) => ({
            filePath: t.filePath,
            diagnostics: [],
            fixes: [],
            suggestionsCount: 0,
            cancelled: false,
          }));
        }
        cancelTask() {
          return false;
        }
        async shutdown() {}
      } as unknown as new (opts: WorkerPoolOptions) => WorkerPoolType,
    };
    const cp = new CompatPool(stubLogger, {
      runnerModule: module,
      now: () => clock,
    });
    await cp.reconfigure(
      [cfg({ dirSuffix: 'gate-reset', pluginPrefixes: ['uc'] })],
      'file:///proj',
    );

    // First batch: init throws → gate trips → caller sees the throw.
    let threw = false;
    try {
      await cp.lintBatch(batch('/a.ts'), noopToken);
    } catch {
      threw = true;
    }
    expect(threw).toBe(true);
    expect(initCalls).toBe(1);

    // Same config, still inside the backoff window → short-circuit, NO
    // retry — even as the clock advances to just before the threshold.
    const blocked = await cp.lintBatch(batch('/b.ts'), noopToken);
    expect(initCalls).toBe(1);
    expect(blocked.results[0].diagnostics).toHaveLength(0);
    clock += 29_999; // still < RETRY_BACKOFF_MS (30_000)
    await cp.lintBatch(batch('/b2.ts'), noopToken);
    expect(initCalls).toBe(1); // window not elapsed → no retry

    // Cross the backoff threshold → next batch retries init (now succeeds).
    clock += 2; // now 30_001 ≥ RETRY_BACKOFF_MS
    const ok = await cp.lintBatch(batch('/c.ts'), noopToken);
    expect(initCalls).toBe(2); // retry happened after the window elapsed
    expect(ok.results).toHaveLength(1);
  });
});
