/**
 * VS Code-side host for ESLint-plugin lint requests.
 *
 * The Go LSP server lints natively, but rules mounted via a config's
 * object-form `plugins` run in JS. So when Go encounters such a rule it sends a
 * server→client `rslint/pluginLint` request back to this extension;
 * we answer it from an in-process WorkerPool owned by `@rslint/core`'s
 * `createPluginLintHost`. This file is a thin lifecycle wrapper over that
 * host — the request→tasks→result boundary itself lives in `@rslint/core`
 * (`buildPluginLintTasks` / `buildPluginLintResult`), shared with the CLI
 * engine so the two paths never drift.
 *
 * The owning runtime injects the factory from the exact project-local
 * `@rslint/core` installation that also supplied the Go binary. This wrapper
 * therefore contains no bundled worker or package-resolution policy.
 */

import { window } from 'vscode';
import type { CancellationToken } from 'vscode';
import type { Logger } from './logger';
import type {
  ConfigDescriptor,
  PluginLintHost,
  EslintPluginLintRequest,
  EslintPluginLintResult,
} from '@rslint/core/eslint-plugin';

type PluginHostFactory = (
  configs: ConfigDescriptor[],
  onLog: (rec: { level: string; source: string; text: string }) => void,
) => Promise<PluginLintHost>;

// One predecessor is retained without a timer so an active commit can be
// rolled back if its JSON-RPC response is lost and Go subsequently aborts.
// Keep one additional grace generation for already-dispatched requests: the
// bound remains two old pools plus the active pool. Hosts with an acquired
// lint lease may temporarily exceed this bound until requests drain.
const MAX_GRACE_GENERATIONS = 1;

export class PluginLintPool {
  private readonly logger: Logger;
  private readonly generations = new Map<string, HostGeneration>();
  private readonly generationRetirementTimers = new Map<
    string,
    ReturnType<typeof setTimeout>
  >();
  private activeGeneration: string | undefined;
  private activeState: HostGeneration | undefined;
  /**
   * The active generation's compensating rollback record. JSON-RPC has no
   * response acknowledgement, so commit cannot discard this predecessor: Go
   * may keep last-good and send abort when the commit response is lost or
   * invalid. A later successful commit proves Go accepted this generation and
   * moves its predecessor into the ordinary grace-retirement queue.
   */
  private activeCommitRollback: ActiveCommitRollback | undefined;
  private readonly liveStates = new Set<HostGeneration>();
  private readonly shutdowns = new Set<Promise<void>>();
  /**
   * Serializes every lifecycle op (prepare/commit/abort/dispose). Each op
   * chains onto the previous one's settlement, so concurrent config reloads
   * cannot race host installation or map mutation. Lint requests for an
   * installed generation take a lease immediately; only a generation that is
   * not installed yet waits for this chain and checks again.
   */
  private opChain: Promise<void> = Promise.resolve();
  private disposed = false;
  private readonly createHost: PluginHostFactory;
  private readonly retirementDelayMs: number;
  private warnedHostFailure = false;

  constructor(
    logger: Logger,
    createHost: PluginHostFactory,
    retirementDelayMs = 30_000,
  ) {
    this.logger = logger;
    this.createHost = createHost;
    this.retirementDelayMs = retirementDelayMs;
  }

  /** Append `op` to the serialized lifecycle chain and await its turn. */
  private async enqueue(op: () => Promise<void>): Promise<void> {
    const run = this.opChain.then(op, op);
    // Keep the chain alive even if `op` throws — swallow on the chain copy so a
    // single failed op doesn't poison every subsequent one. Callers awaiting
    // the returned promise still observe the rejection.
    this.opChain = run.catch(() => undefined);
    return run;
  }

  /**
   * Prepare a generation without making it the active fallback for requests
   * without a key. The transport commits it at the matching config
   * transaction's commit point;
   * an abort after commit can still compensate for a lost response and return
   * to the prior Go last-good generation.
   *
   * Returns whether the requested host state is active. Rebuilds are
   * transactional: a failed replacement leaves the previous host available so
   * the caller can preserve the matching last-good config payload.
   *
   * Empty `descriptors` needs no host, avoiding a module load and worker-pool
   * allocation when no object-form community plugins are configured. The
   * matching activation publishes no plugin metadata, so Go must never issue
   * a plugin-lint request for that generation without a host.
   */
  async prepare(
    descriptors: ConfigDescriptor[],
    fingerprint: string,
    generation: string,
  ): Promise<boolean> {
    let ready = false;
    await this.enqueue(async () => {
      if (this.disposed || generation === '') return;

      const existing = this.generations.get(generation);
      if (existing) {
        ready = existing.ready;
        return;
      }

      // Config-only changes can reuse the same plugin host. The new generation
      // is still staged separately and is not routable as active until commit.
      if (
        this.activeState?.ready &&
        this.activeState.fingerprint === fingerprint
      ) {
        this.generations.set(generation, this.activeState);
        ready = true;
        return;
      }

      if (descriptors.length === 0) {
        const state: HostGeneration = {
          fingerprint,
          ready: true,
          activeLints: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        ready = true;
        return;
      }

      try {
        const replacement = await this.createHost(descriptors, (rec) => {
          const text = `[rslint:plugin] ${rec.text}`;
          if (rec.level === 'error') this.logger.error(text);
          else this.logger.debug(text);
        });
        if (this.disposed) {
          // Disposed while initializing — shut the fresh pool back down.
          await replacement.shutdown().catch(() => undefined);
          return;
        }
        const state: HostGeneration = {
          fingerprint,
          host: replacement,
          ready: true,
          activeLints: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        ready = true;
      } catch (err: unknown) {
        // Init failed because the selected local core's host or a referenced
        // plugin could not load. Keep the previous active host intact. Record
        // an unavailable staged generation so the first valid config can still
        // be committed and serve native rules; later prepares remain retryable.
        const state: HostGeneration = {
          fingerprint,
          ready: false,
          activeLints: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        this.logger.error('Failed to initialize ESLint-plugin host', err);
        // Make the failure visible — but ONLY when a config actually mounted
        // plugins (an empty-descriptor host builds no worker and failing is
        // not a user-facing problem), and only once per session so a
        // persistent failure doesn't re-warn on every reload. Keep this latch
        // per runtime so one broken local core cannot hide another core's
        // independently actionable failure.
        if (descriptors.length > 0 && !this.warnedHostFailure) {
          this.warnedHostFailure = true;
          void window.showWarningMessage(
            'Rslint: failed to load the ESLint-plugin host; rules mounted via a config’s `plugins` will report no diagnostics. See the Rslint output channel for details.',
          );
        }
      }
    });
    return ready;
  }

  /** Commit a previously prepared generation after Go accepts its config. */
  async commit(generation: string): Promise<boolean> {
    let committed = false;
    await this.enqueue(async () => {
      if (this.disposed) return;
      const next = this.generations.get(generation);
      if (!next) return;
      if (generation === this.activeGeneration) {
        committed = true;
        return;
      }

      const previousGeneration = this.activeGeneration;
      const previous = this.activeState;
      this.finalizeActiveCommitRollback();
      if (previousGeneration) {
        this.cancelGenerationRetirement(previousGeneration);
      }
      this.activeGeneration = generation;
      this.activeState = next;
      this.activeCommitRollback = {
        generation,
        previousGeneration,
        previousState: previous,
      };
      committed = true;
    });
    return committed;
  }

  /** Discard a staged generation when source validation or Go commit fails. */
  async abort(generation: string): Promise<void> {
    await this.enqueue(async () => {
      if (generation === this.activeGeneration) {
        const rollback = this.activeCommitRollback;
        if (!rollback || rollback.generation !== generation) return;
        const aborted = this.activeState;
        this.activeGeneration = rollback.previousGeneration;
        this.activeState = rollback.previousState;
        this.activeCommitRollback = undefined;
        if (rollback.previousGeneration) {
          this.cancelGenerationRetirement(rollback.previousGeneration);
        }
        this.generations.delete(generation);
        if (
          aborted &&
          aborted !== this.activeState &&
          !this.hasGenerationReference(aborted)
        ) {
          this.retire(aborted);
        }
        return;
      }
      const state = this.generations.get(generation);
      if (!state) return;
      this.generations.delete(generation);
      if (state !== this.activeState && !this.hasGenerationReference(state)) {
        this.retire(state);
      }
    });
  }

  /** Answer one reverse `rslint/pluginLint` request. */
  async lint(
    req: EslintPluginLintRequest,
    token?: CancellationToken,
  ): Promise<EslintPluginLintResult> {
    if (this.disposed) return { results: [] };

    let state = req.generation
      ? this.generations.get(req.generation)
      : this.activeState;

    // A reverse request may arrive after Go accepts a config but just before
    // Node installs that generation. Wait only in that case. Existing
    // generations must remain routable while an unrelated prepare is slow.
    if (req.generation && !state) {
      if (!(await this.waitForLifecycle(token))) return { results: [] };
      if (this.disposed) return { results: [] };
      state = this.generations.get(req.generation);
    }
    if (req.generation && !state) {
      throw new Error(
        `unknown ESLint-plugin config generation: ${req.generation}`,
      );
    }
    if (!state) return { results: [] };
    const host = state.host;
    if (!host) {
      // Generations without a host are valid committed states for native-only or
      // degraded catalogs, but their activation exposes no plugin metadata.
      // Reaching this branch therefore means Go and the extension disagree on
      // the committed lifecycle. Do not turn that protocol failure into a
      // false-green empty diagnostic set. Cancellation remains benign.
      if (token?.isCancellationRequested) return { results: [] };
      const generation = req.generation ?? this.activeGeneration;
      throw new Error(
        `LSP pluginLint requested for config generation ${JSON.stringify(generation)} without an activated plugin host`,
      );
    }

    // Take the lease before yielding. Retirement removes future routing
    // references, but cannot shut this state down until the lease is released.
    state.activeLints++;
    // Bridge the LSP CancellationToken → AbortSignal for the core host, so a
    // superseding keystroke / close (Go sends $/cancelRequest) stops the worker
    // instead of letting it run to completion.
    let signal: AbortSignal | undefined;
    let cancellationSubscription: { dispose(): unknown } | undefined;
    try {
      if (token) {
        const ac = new AbortController();
        if (token.isCancellationRequested) ac.abort();
        else
          cancellationSubscription = token.onCancellationRequested(() => {
            ac.abort();
          });
        signal = ac.signal;
      }
      return await host.lint(req, signal);
    } finally {
      cancellationSubscription?.dispose();
      state.activeLints--;
      if (state.retiring && state.activeLints === 0) {
        this.startShutdown(state);
      }
    }
  }

  /** Wait for the lifecycle snapshot that could be installing a generation. */
  private async waitForLifecycle(token?: CancellationToken): Promise<boolean> {
    const pending = this.opChain;
    if (!token) {
      await pending;
      return true;
    }
    if (token.isCancellationRequested) return false;

    let cancellationSubscription: { dispose(): unknown } | undefined;
    const cancelled = new Promise<false>((resolve) => {
      cancellationSubscription = token.onCancellationRequested(() => {
        resolve(false);
      });
    });
    try {
      return await Promise.race([pending.then(() => true as const), cancelled]);
    } finally {
      cancellationSubscription?.dispose();
    }
  }

  private hasGenerationReference(state: HostGeneration): boolean {
    for (const candidate of this.generations.values()) {
      if (candidate === state) return true;
    }
    return false;
  }

  private retire(state: HostGeneration): void {
    state.retiring = true;
    if (state.activeLints === 0) this.startShutdown(state);
  }

  private finalizeActiveCommitRollback(): void {
    const rollback = this.activeCommitRollback;
    if (!rollback) return;
    this.activeCommitRollback = undefined;
    if (rollback.previousGeneration) {
      this.scheduleGenerationRetirement(
        rollback.previousGeneration,
        rollback.previousState,
      );
    }
  }

  private cancelGenerationRetirement(generation: string): void {
    const timer = this.generationRetirementTimers.get(generation);
    if (!timer) return;
    clearTimeout(timer);
    this.generationRetirementTimers.delete(generation);
  }

  private scheduleGenerationRetirement(
    generation: string,
    state: HostGeneration | undefined,
  ): void {
    const existing = this.generationRetirementTimers.get(generation);
    if (existing) clearTimeout(existing);
    const timer = setTimeout(() => {
      this.completeGenerationRetirement(generation, state);
    }, this.retirementDelayMs);
    this.generationRetirementTimers.set(generation, timer);

    // A burst of config updates must not retain one complete WorkerPool per
    // generation for the full production grace period. Expire the oldest
    // routing generation immediately once the bounded history is full.
    while (this.generationRetirementTimers.size > MAX_GRACE_GENERATIONS) {
      const oldest = this.generationRetirementTimers.keys().next().value;
      if (oldest === undefined) break;
      this.completeGenerationRetirement(oldest, this.generations.get(oldest));
    }
  }

  private completeGenerationRetirement(
    generation: string,
    state: HostGeneration | undefined,
  ): void {
    const timer = this.generationRetirementTimers.get(generation);
    if (!timer) return;
    clearTimeout(timer);
    this.generationRetirementTimers.delete(generation);
    if (this.activeGeneration === generation) return;
    if (this.generations.get(generation) !== state) return;

    this.generations.delete(generation);
    if (
      state &&
      state !== this.activeState &&
      !this.hasGenerationReference(state)
    ) {
      this.retire(state);
    }
  }

  private startShutdown(state: HostGeneration): void {
    if (state.shutdown) return;
    for (const [generation, candidate] of this.generations) {
      if (candidate === state) {
        this.generations.delete(generation);
        const timer = this.generationRetirementTimers.get(generation);
        if (timer) clearTimeout(timer);
        this.generationRetirementTimers.delete(generation);
      }
    }
    const shutdown = state.host
      ? state.host.shutdown().catch((err: unknown) => {
          this.logger.error('Error shutting down previous plugin host', err);
        })
      : Promise.resolve();
    state.shutdown = shutdown;
    this.shutdowns.add(shutdown);
    void shutdown.finally(() => {
      this.shutdowns.delete(shutdown);
      this.liveStates.delete(state);
    });
  }

  /** Shut down the worker pool. Idempotent. */
  async dispose(): Promise<void> {
    this.disposed = true;
    await this.enqueue(async () => {
      const states = [...this.liveStates];
      this.generations.clear();
      for (const timer of this.generationRetirementTimers.values()) {
        clearTimeout(timer);
      }
      this.generationRetirementTimers.clear();
      this.activeGeneration = undefined;
      this.activeState = undefined;
      this.activeCommitRollback = undefined;
      for (const state of states) {
        // Terminal disposal intentionally forces shutdown even if a request is
        // still active; WorkerPool turns those tasks into benign cancellation.
        this.startShutdown(state);
      }
    });
    await Promise.all([...this.shutdowns]);
  }
}

interface HostGeneration {
  fingerprint: string;
  host?: PluginLintHost;
  ready: boolean;
  activeLints: number;
  retiring: boolean;
  shutdown?: Promise<void>;
}

interface ActiveCommitRollback {
  generation: string;
  previousGeneration: string | undefined;
  previousState: HostGeneration | undefined;
}
