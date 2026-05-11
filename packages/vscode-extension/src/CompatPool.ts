/**
 * CompatPool — extension-side WorkerPool wrapper for ESLint-plugin rule
 * execution.
 *
 * Replaces the previous Go-spawns-Node sidecar architecture. The Go LSP
 * server now sends each ESLint-plugin batch as a `rslint/lintCompatBatch`
 * LSP custom request; the extension handles it by dispatching to the
 * WorkerPool owned by this class.
 *
 * Responsibilities:
 *
 *   - Lifecycle (three-tier, mirrors the previous syncSidecarToConfig
 *     policy that lived in Go):
 *       * unchanged plugin set       → noop
 *       * changed plugin set         → drain old pool, spawn new pool
 *       * cleared (no plugin rules)  → drain, leave pool null
 *   - Lazy init: the pool isn't actually spawned on `reconfigure`; we
 *     only record the desired state. The first lintBatch triggers
 *     `pool.init()`. Rationale: a workspace whose config declares no
 *     `eslintPlugins` should pay zero worker_thread cost; same for one
 *     that declares plugins but never gets linted (idle editor).
 *   - Per-file plugin routing: every file in a batch carries a
 *     `configKey` (the config directory path). The worker uses that
 *     key to pick the right `LoadedPlugins` from its per-config map.
 *     This is the monorepo multi-version dispatch contract: file
 *     under /pkg-a sees /pkg-a/node_modules's plugin, file under
 *     /pkg-b sees /pkg-b's, because the worker imported each config
 *     file separately and Node's ESM resolver anchored at the config's
 *     own location. Unmatched configKey → worker emits an
 *     internal-error parseError (an invariant violation: every
 *     configKey on the wire was previously declared in
 *     WorkerPoolOptions.configs[]).
 *   - Cancellation: the LSP custom-request handler receives a
 *     `CancellationToken`; when fired, we propagate to every taskId
 *     dispatched for the current batch via `pool.cancelTask`. Workers
 *     observe the per-task SharedArrayBuffer flag and bail at the next
 *     per-node visit.
 *   - Dynamic import: the runner is ESM, the extension is CJS; we use
 *     `await import('@rslint/eslint-plugin-runner')` inside async
 *     methods. The runner is marked `external` in the extension's
 *     esbuild config so its `dist/` ships alongside the extension.
 */

import { Disposable, CancellationToken } from 'vscode';
import { Logger } from './logger';
import type {
  WorkerPool as WorkerPoolType,
  WorkerPoolOptions as WorkerPoolOptionsType,
  ConfigDescriptor,
  CompatBatchInput,
  CompatBatchResult,
} from '@rslint/eslint-plugin-runner';
import {
  extractConfigDescriptors,
  fingerprintConfigs,
  type NormalizedConfig,
} from './compat-pool-helpers';

// Re-export NormalizedConfig so callers don't need to know about the
// helper module split.
export type { NormalizedConfig };

// The two pure result builders from the runner package are loaded by
// DYNAMIC import only — the package is ESM-only and the extension bundles
// to CJS with it marked external, so a static `import` compiles to a
// top-level `require("@rslint/eslint-plugin-runner")` that throws
// ERR_REQUIRE_ESM on VS Code hosts whose Node can't require ESM
// (Node < 22 / VS Code <= 1.100). WorkerPool is loaded the same way; the
// builders are pure so — unlike WorkerPool — they need no test hook.
type RunnerBuilders = Pick<
  typeof import('@rslint/eslint-plugin-runner'),
  | 'buildCompatTasksByConfigKey'
  | 'buildCompatBatchResult'
  | 'isWorkerClosedError'
>;
let cachedRunnerBuilders: RunnerBuilders | undefined;
async function loadRunnerBuilders(): Promise<RunnerBuilders> {
  if (!cachedRunnerBuilders) {
    const mod = await import('@rslint/eslint-plugin-runner');
    cachedRunnerBuilders = {
      buildCompatTasksByConfigKey: mod.buildCompatTasksByConfigKey,
      buildCompatBatchResult: mod.buildCompatBatchResult,
      isWorkerClosedError: mod.isWorkerClosedError,
    };
  }
  return cachedRunnerBuilders;
}

/**
 * Wire schema for `rslint/lintCompatBatch` — matches Go's
 * `internal/lsp.LintCompatBatchParams` (an alias of linter.CompatBatch).
 *
 * Defined here as an alias to the shared `CompatBatchInput` so the
 * CLI host (engine.ts) and the LSP host (this file) decode the same
 * wire shape — any future field added on the Go side lands in one
 * runner-package type and both consumers get it automatically.
 */
export type LintCompatBatchParams = CompatBatchInput;

/**
 * Wire response — matches Go's `LintCompatBatchResult`. Alias of the
 * shared `CompatBatchResult` for the same reason as
 * {@link LintCompatBatchParams}.
 */
export type LintCompatBatchResult = CompatBatchResult;

export class CompatPool implements Disposable {
  private readonly logger: Logger;

  /**
   * Currently running pool. Null when:
   *  - never spawned (no reconfigure call yet, or all configs had no
   *    eslintPlugins),
   *  - drained for a tier-3 transition,
   *  - drained for a tier-2 restart (briefly nil until lazy re-init).
   */
  private pool: WorkerPoolType | null = null;

  /**
   * Identity of the entry set that produced (or will produce) the
   * current pool. Null when there's nothing to lint. Used by
   * `reconfigure` to skip work when the user's config save didn't
   * change anything load-bearing.
   */
  private currentFingerprint: string | null = null;

  /**
   * Descriptors for every config the pool is responsible for. Each
   * descriptor's `configPath` is what the worker imports; its
   * `configDirectory` is the filesystem path Go writes into per-file
   * `configKey` during compat dispatch.
   *
   * Only configs that declare at least one plugin go here — the worker
   * pool pays one config-import per descriptor at init.
   */
  private currentConfigs: ConfigDescriptor[] = [];

  /**
   * Set of `configDirectory` strings the pool knows about. Used purely
   * to surface "unknown configKey on the wire" warnings before the
   * worker's terser internal-error parseError lands. Must byte-equal
   * what Go writes (filesystem path, forward-slashed).
   */
  private configDirSet: Set<string> = new Set();

  /**
   * Init-failure circuit-breaker. Re-attempting init on every batch after
   * a failure (e.g. a structurally broken plugin import) re-thrashes worker
   * spawn / ESM import — multiple seconds per attempt in big monorepos. So
   * we record the failing fingerprint + timestamp and short-circuit
   * subsequent batches against the SAME configs to empty results.
   *
   * Recovery is TIME-BASED: after RETRY_BACKOFF_MS we allow exactly one
   * retry (re-arming the timer if it fails again). This un-sticks the pool
   * once the environment is fixed — e.g. the user installs the missing
   * plugin — WITHOUT watching node_modules/lockfiles: a dependency install
   * changes neither the config nor its fingerprint, so a fingerprint-only
   * reset would never fire. A reconfigure to different configs still clears
   * it immediately (the fingerprint no longer matches in the gate check).
   */
  private static readonly RETRY_BACKOFF_MS = 30_000;
  private failedFingerprint: string | null = null;
  private failedAt = 0;

  /**
   * Serialization chain for pool-state-mutating work.
   *
   * Two operations mutate pool state: `reconfigure` (drains old, swaps
   * entries) and the lazy-init phase of `lintBatch` (spawns a new pool
   * the first time it's needed). Both must observe consistent state and
   * must not interleave.
   *
   * Without a chain we hit two real races:
   *
   *   B1 — Concurrent lintBatch calls during cold init: both pass the
   *        `if (!this.pool)` check, both `await import()`, both assign
   *        `this.pool = new WorkerPool(...)`. The second assignment
   *        overwrites the first, orphaning its worker_threads.
   *
   *   B2 — reconfigure-vs-lintBatch: reconfigure sets `this.pool = null`
   *        then `await stale.shutdown()`. During that await,
   *        `currentConfigs` is the new value but `this.pool` is
   *        also null — a concurrent lintBatch sees null and spawns a
   *        NEW pool. BUT: if state assignment came AFTER the await,
   *        the concurrent lintBatch would spawn a pool against the
   *        OLD entries (silently running outdated plugins).
   *
   * Both go away if we serialize on a Promise chain. `reconfigure` and
   * the init phase of `lintBatch` both pass through `.chain(...)`. The
   * lintBatch's actual `pool.lintBatch(tasks)` work runs OUTSIDE the
   * chain so unrelated batches don't block on each other — only the
   * init / drain serializes.
   */
  private opChain: Promise<unknown> = Promise.resolve();

  /**
   * Queue `work` after any in-flight chained operation. Errors in
   * prior chain entries do NOT block subsequent ones (we
   * `.catch(() => {})` between chain links); they're surfaced to the
   * original caller of that prior operation.
   */
  private async chain<T>(work: () => Promise<T>): Promise<T> {
    await this.opChain.catch(() => undefined);
    const result = work();
    this.opChain = result;
    return result;
  }

  /**
   * Optional injection point for tests: when set, lazy init uses this
   * instead of `await import('@rslint/eslint-plugin-runner')`. Production
   * callers leave this unset so the runner module loads lazily; tests
   * pass a stub WorkerPool to exercise CompatPool's lifecycle logic
   * without spawning real worker_threads.
   *
   * Only the `WorkerPool` named export is read — see lazy init in
   * `lintBatch`.
   */
  private readonly runnerModule:
    | { WorkerPool: new (opts: WorkerPoolOptionsType) => WorkerPoolType }
    | undefined;

  /**
   * Clock for the init-failure retry backoff. Injectable so tests advance
   * time deterministically; production uses Date.now.
   */
  private readonly now: () => number;

  constructor(
    logger: Logger,
    opts?: {
      runnerModule?: {
        WorkerPool: new (opts: WorkerPoolOptionsType) => WorkerPoolType;
      };
      now?: () => number;
    },
  ) {
    this.logger = logger;
    this.runnerModule = opts?.runnerModule;
    this.now = opts?.now ?? Date.now;
  }

  /**
   * Apply a new normalized-config set. Async because tier-2 / tier-3
   * transitions need to drain (await pool.shutdown()) before returning;
   * a caller that immediately follows up with lintBatch must see a
   * fully torn-down prior pool, not a half-shutdown one.
   *
   * Callers should await this; firing parallel reconfigures from
   * back-to-back config-on-disk changes is harmless but the second one
   * waits on the first.
   */
  async reconfigure(
    configs: NormalizedConfig[],
    _workspaceFallbackUrl: string,
  ): Promise<void> {
    return this.chain(async () => {
      // 1. Descriptors for the WorkerPool (one config-import per descriptor).
      // 2. `configDirSet` for the unknown-configKey warning sink.
      // 3. Identity fingerprint for tier-1 short-circuit.
      const descriptors = extractConfigDescriptors(configs);
      const dirSet = new Set(descriptors.map((d) => d.configDirectory));
      const fp = fingerprintConfigs(configs);

      // ── Three-tier transition ────────────────────────────────────
      if (fp === this.currentFingerprint) {
        // tier 1: unchanged. Refresh the dir-set in case nothing-load-
        // bearing-but-still-relevant shifted (descriptor ordering).
        this.configDirSet = dirSet;
        return;
      }

      // tier 2 or tier 3 — drain old pool. CRITICAL: update state
      // BEFORE awaiting shutdown. With state already pointing at the
      // new descriptors, any next chained operation (e.g. an enqueued
      // lintBatch) that runs after the await sees the new state and
      // spawns a fresh pool against the NEW descriptors. If state were
      // updated AFTER the await, that next operation would spawn a
      // pool against stale descriptors and the user's config change
      // would silently fail.
      //
      // The chain itself prevents concurrent ops from running during
      // this method, so reads of currentFingerprint / currentConfigs
      // by a queued caller are serialized after the assignment block
      // below.
      const stale = this.pool;
      this.pool = null;
      this.currentFingerprint = fp;
      if (descriptors.length === 0) {
        // tier 3: nothing to lint.
        this.currentConfigs = [];
        this.configDirSet = new Set();
      } else {
        // tier 2: descriptors changed. Pool re-spawn happens lazily in
        // the next lintBatch chained call.
        this.currentConfigs = descriptors;
        this.configDirSet = dirSet;
      }

      if (stale) {
        try {
          await stale.shutdown();
        } catch (err) {
          this.logger.error('CompatPool: error shutting down stale pool', err);
        }
      }

      if (descriptors.length === 0) {
        this.logger.debug(
          'CompatPool: reconfigured to zero plugin-bearing configs; pool stays unspawned',
        );
      } else {
        this.logger.debug(
          `CompatPool: reconfigured with ${descriptors.length} config(s); pool will spawn on first lintBatch`,
        );
      }
    });
  }

  /**
   * Handle one `rslint/lintCompatBatch` request. Returns the wire-shape
   * the Go server expects.
   *
   * Cancellation: caller passes the LSP request's CancellationToken;
   * when the token cancels, we issue cancelTask for every dispatched
   * taskId. Workers observe the per-task SAB flag and bail at the
   * next per-node visit. The pool's `lintBatch` itself resolves
   * normally (each cancelled task returns a result with
   * `cancelled: true`); we surface that to the Go side which folds it
   * into the standard diagnostic stream.
   *
   * Errors: a thrown error here turns into a JSON-RPC error response
   * to the Go side. The Go dispatcher logs it and marks the batch's
   * files as compat-skipped. Per-rule failures DON'T throw — they
   * land in result[i].ruleErrors and are surfaced as user-visible
   * stderr by the linter.
   */
  async lintBatch(
    params: LintCompatBatchParams,
    token: CancellationToken,
  ): Promise<LintCompatBatchResult> {
    // Lazy init + state read go through the chain so we (a) never start
    // a second pool while the first is still being initialized and
    // (b) never observe a half-applied reconfigure.
    //
    // The chained block resolves with either the live pool (when there
    // are plugin entries) or null (tier-3 empty state). The actual
    // `pool.lintBatch(tasks)` work runs OUTSIDE the chain so unrelated
    // batches can proceed in parallel — only init / drain serializes.
    const pool = await this.chain(async () => {
      if (this.currentConfigs.length === 0) {
        return null;
      }
      if (this.pool) {
        return this.pool;
      }
      // Circuit-breaker: if init failed against the EXACT same configs
      // before, short-circuit subsequent batches to empty rather than
      // re-thrashing worker spawn / ESM import (dominant cost in big
      // monorepos — multiple seconds per lint trigger). Recovery is
      // time-based: once RETRY_BACKOFF_MS has elapsed we fall through and
      // allow ONE retry (re-armed below if it fails again), so installing
      // the missing plugin un-sticks the pool without watching
      // node_modules. A reconfigure to different configs clears it
      // immediately (fingerprint no longer matches).
      if (
        this.failedFingerprint === this.currentFingerprint &&
        this.now() - this.failedAt < CompatPool.RETRY_BACKOFF_MS
      ) {
        return null;
      }
      const { WorkerPool } =
        this.runnerModule ?? (await import('@rslint/eslint-plugin-runner'));
      const fresh = new WorkerPool({
        configs: this.currentConfigs,
        onLog: (rec) => {
          // Forward plugin / runner diagnostics to the extension's
          // logger so users can see them in the "Rslint" output channel
          // exactly like other extension logs.
          const text = `[runner:${rec.source}] ${rec.text}`;
          if (rec.level === 'error') this.logger.error(text);
          else if (rec.level === 'warn') this.logger.warn(text);
          else this.logger.debug(text);
        },
      });
      try {
        await fresh.init();
      } catch (err) {
        // Mark this exact fingerprint as failed, WITH the time, so
        // subsequent batches short-circuit (above) instead of re-thrashing
        // — until RETRY_BACKOFF_MS elapses (one retry) or a reconfigure
        // changes currentFingerprint.
        this.failedFingerprint = this.currentFingerprint;
        this.failedAt = this.now();
        this.logger.error(
          `CompatPool: init failed; subsequent batches return empty results for up to ${CompatPool.RETRY_BACKOFF_MS}ms or until config changes`,
          err,
        );
        throw err;
      }
      this.pool = fresh;
      // Successful init clears any prior failure marker (configs may
      // have been edited to a previously-broken-now-fixed state).
      this.failedFingerprint = null;
      return fresh;
    });

    // Fast path: no plugin entries currently configured. Returning a
    // result-per-file array (with empty diagnostics each) preserves the
    // Go-side cardinality invariant in lintcompat_dispatcher.go.
    if (pool === null) {
      return {
        results: params.files.map((f) => ({
          filePath: f.path,
          diagnostics: [],
          cancelled: false,
        })),
      };
    }

    // Result builders come from a DYNAMIC import (never static): the
    // runner package is ESM-only and the extension bundles to CJS, so a
    // static import compiles to a top-level `require()` of ESM and throws
    // ERR_REQUIRE_ESM on VS Code hosts whose Node can't require ESM
    // (Node < 22). WorkerPool above is loaded the same way.
    const {
      buildCompatTasksByConfigKey,
      buildCompatBatchResult,
      isWorkerClosedError,
    } = await loadRunnerBuilders();

    // Per-file routing + LintTask building is shared with the CLI
    // host (packages/rslint/src/engine.ts) via
    // `buildCompatTasksByConfigKey`. The builder pass-throughs each
    // file's `configKey`; the worker picks the right `LoadedPlugins`
    // from its per-config map.
    const tasks = buildCompatTasksByConfigKey(params, {
      configDirSet: this.configDirSet,
      onUnknownConfigKey: (filePath, configKey) => {
        this.logger.warn(
          `CompatPool: file ${filePath} carries unknown configKey ${JSON.stringify(
            configKey,
          )}; eslint-plugin rules will not run on it`,
        );
      },
    });

    // Wire up cancellation. Subscribed BEFORE dispatch so a cancel
    // that fires while pool.lintBatch is mid-dispatch finds the
    // taskId list already partially populated.
    //
    // Three timing windows the design has to handle without ever
    // "leaking" a task past cancel:
    //
    //   A. **Token already cancelled when we register**: per the VS Code
    //      contract, `onCancellationRequested` invokes the listener
    //      immediately. At that instant `dispatchedIds` is still empty,
    //      so the for-loop here is a no-op. The `cancelled` flag below
    //      catches every task as it gets dispatched.
    //   B. **Cancel during dispatch loop**: cancel listener runs in the
    //      middle of pool.lintBatch's per-task dispatch. Already-
    //      dispatched ids in `dispatchedIds` get cancelTask'd by the
    //      listener; subsequently dispatched ones get cancelTask'd by
    //      the `if (cancelled)` branch in onTaskDispatched.
    //   C. **Cancel after dispatch loop**: every id is in
    //      `dispatchedIds`; the listener iterates them all.
    //
    // We target the LOCAL `pool` variable, not `this.pool`. A concurrent
    // reconfigure may null out `this.pool` and start a new one mid-batch;
    // we still want cancel to reach the workers that are actually
    // running OUR tasks (which is `pool`, the one we held throughout
    // this method). cancelTask is a no-op on a shut-down pool, so this
    // stays safe even if reconfigure already drained `pool`.
    let cancelled = false;
    const dispatchedIds: number[] = [];
    const cancelSub = token.onCancellationRequested(() => {
      cancelled = true;
      for (const id of dispatchedIds) pool.cancelTask(id);
    });

    try {
      const results = await pool.lintBatch(tasks, (taskId) => {
        dispatchedIds.push(taskId);
        // Window-B catch-up: if the cancel listener already fired
        // (either case A — registered an already-cancelled token — or
        // case B — fired mid-dispatch loop, before this task's slot
        // existed), propagate the cancel to this freshly dispatched
        // task now. WorkerPool's `onTaskDispatched` now fires after
        // `inflight.set`, so cancelTask reaches the SAB flag before
        // the worker starts polling.
        if (cancelled) pool.cancelTask(taskId);
      });
      // Shared result projection so both host paths hand Go an
      // identical, byte-stable wire shape.
      return buildCompatBatchResult(results);
    } catch (err) {
      // A concurrent reconfigure/dispose can close `pool` between our
      // chained acquisition and this (deliberately un-chained) lintBatch
      // call — WorkerPool / IpcClient then reject with a WorkerClosedError.
      // Treat that narrow window as "no compat diagnostics for this batch"
      // rather than failing the whole batch; the superseding reconfigure's
      // next batch lints against the fresh pool. The runner's structural
      // guard keys on the stable error code (not message text), so it works
      // across the ESM-runner / CJS-extension boundary.
      if (isWorkerClosedError(err)) {
        return {
          results: params.files.map((f) => ({
            filePath: f.path,
            diagnostics: [],
            cancelled: false,
          })),
        };
      }
      throw err;
    } finally {
      cancelSub.dispose();
    }
  }

  /**
   * Drain workers and release resources. Called from extension
   * disposal (workspace folder unloaded / VS Code shutting down).
   * Safe to call multiple times. Runs through the same serialization
   * chain as reconfigure / lintBatch so dispose can't interleave with
   * an in-flight init or reconfigure.
   */
  async dispose(): Promise<void> {
    return this.chain(async () => {
      const p = this.pool;
      this.pool = null;
      this.currentFingerprint = null;
      this.currentConfigs = [];
      this.configDirSet = new Set();
      if (p) {
        try {
          await p.shutdown();
        } catch (err) {
          this.logger.error('CompatPool: error during shutdown', err);
        }
      }
    });
  }
}
