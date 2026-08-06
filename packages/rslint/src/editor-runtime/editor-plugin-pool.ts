import type { ConfigModuleActivationPlan } from '../config/config-loader.js';

interface PluginLintRequest {
  generation?: string;
  [key: string]: unknown;
}

export interface EditorPluginLintHost {
  lint(request: PluginLintRequest, signal?: AbortSignal): Promise<unknown>;
  shutdown(): Promise<void>;
}

export interface EditorPluginModule {
  createPluginLintHost(
    configs: ConfigModuleActivationPlan['pluginConfigs'],
    onLog?: (record: { level: string; source: string; text: string }) => void,
    singleThreaded?: boolean,
  ): Promise<EditorPluginLintHost>;
}

interface HostState {
  readonly fingerprint: string;
  readonly host?: EditorPluginLintHost;
  readonly ready: boolean;
  leases: number;
  retiring: boolean;
  shutdown?: Promise<void>;
}

export interface EditorPluginPoolOptions {
  readonly loadPluginModule?: () => Promise<EditorPluginModule>;
  readonly graceGenerationMs?: number;
  readonly maxGraceGenerations?: number;
  readonly log?: (message: string) => void;
}

const DEFAULT_GRACE_GENERATION_MS = 30_000;
const DEFAULT_MAX_GRACE_GENERATIONS = 1;

let pluginModulePromise: Promise<EditorPluginModule> | undefined;

async function loadDefaultPluginModule(): Promise<EditorPluginModule> {
  // The library entry is flattened to dist/editor-runtime.js; the worker host
  // is emitted separately at dist/eslint-plugin/index.js.
  const entry: string = './eslint-plugin/index.js';
  pluginModulePromise ??= import(
    /* webpackIgnore: true */ entry
  ) as Promise<EditorPluginModule>;
  try {
    return await pluginModulePromise;
  } catch (error) {
    pluginModulePromise = undefined;
    throw error;
  }
}

function activationFingerprint(
  plan: ConfigModuleActivationPlan,
  dependencyRevision: number,
): string {
  const sourceByPath = new Map(
    plan.configs.map((config) => [config.configPath, config.sourceFingerprint]),
  );
  const plugins = [...plan.pluginConfigs]
    .map((config) => {
      const source = sourceByPath.get(config.configPath);
      if (source === undefined) {
        throw new Error(
          `missing source fingerprint for plugin config ${config.configPath}`,
        );
      }
      return `${config.configPath}\0${source}`;
    })
    .sort();
  return `${dependencyRevision}\0${plugins.join('\0')}`;
}

/**
 * Bounded generation router for one physical core installation. The active
 * generation and at most one grace generation own workers; exact active input
 * fingerprints reuse the same host.
 */
export class EditorPluginPool {
  private readonly generations = new Map<string, HostState>();
  private readonly staged = new Set<string>();
  private readonly retirementTimers = new Map<
    string,
    ReturnType<typeof setTimeout>
  >();
  private readonly liveStates = new Set<HostState>();
  private readonly loadPluginModule: () => Promise<EditorPluginModule>;
  private readonly graceGenerationMs: number;
  private readonly maxGraceGenerations: number;
  private readonly log: (message: string) => void;
  private activeGeneration: string | undefined;
  private activeState: HostState | undefined;
  private operation: Promise<void> = Promise.resolve();
  private disposed = false;

  constructor(options: EditorPluginPoolOptions = {}) {
    this.loadPluginModule = options.loadPluginModule ?? loadDefaultPluginModule;
    this.graceGenerationMs =
      options.graceGenerationMs ?? DEFAULT_GRACE_GENERATION_MS;
    this.maxGraceGenerations =
      options.maxGraceGenerations ?? DEFAULT_MAX_GRACE_GENERATIONS;
    this.log = options.log ?? ((message) => process.stderr.write(message));
    if (
      !Number.isSafeInteger(this.maxGraceGenerations) ||
      this.maxGraceGenerations < 0
    ) {
      throw new Error('maxGraceGenerations must be a non-negative integer');
    }
    if (
      !Number.isFinite(this.graceGenerationMs) ||
      this.graceGenerationMs < 0
    ) {
      throw new Error('graceGenerationMs must be non-negative');
    }
  }

  async prepare(
    plan: ConfigModuleActivationPlan,
    dependencyRevision: number,
  ): Promise<boolean> {
    let ready = false;
    await this.enqueue(async () => {
      this.assertOpen();
      const generation = plan.transactionId;
      const fingerprint = activationFingerprint(plan, dependencyRevision);
      const existing = this.generations.get(generation);
      if (existing) {
        ready = existing.ready;
        return;
      }

      if (
        this.activeState &&
        this.activeState.ready &&
        this.activeState.fingerprint === fingerprint &&
        !this.activeState.retiring
      ) {
        this.generations.set(generation, this.activeState);
        this.staged.add(generation);
        ready = true;
        return;
      }

      if (plan.pluginConfigs.length === 0) {
        const state: HostState = {
          fingerprint,
          ready: true,
          leases: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        this.staged.add(generation);
        ready = true;
        return;
      }

      try {
        const module = await this.loadPluginModule();
        const host = await module.createPluginLintHost(
          plan.pluginConfigs,
          (record) => {
            this.report(`[rslint:plugin:${record.level}] ${record.text}\n`);
          },
        );
        if (this.disposed) {
          await Promise.resolve()
            .then(async () => host.shutdown())
            .catch(() => undefined);
          return;
        }
        const state: HostState = {
          fingerprint,
          host,
          ready: true,
          leases: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        this.staged.add(generation);
        ready = true;
      } catch (error) {
        this.report(
          `rslint: failed to initialize ESLint-plugin host: ${String(error)}\n`,
        );
        // A first-start catalog is allowed to commit as native-only when the
        // optional plugin host is broken. Do not reuse this unavailable state:
        // the next transaction must retry initialization.
        const state: HostState = {
          fingerprint,
          ready: false,
          leases: 0,
          retiring: false,
        };
        this.liveStates.add(state);
        this.generations.set(generation, state);
        this.staged.add(generation);
      }
    });
    return ready;
  }

  async commit(generation: string): Promise<boolean> {
    let committed = false;
    await this.enqueue(() => {
      this.assertOpen();
      if (!this.staged.has(generation)) return;
      const next = this.generations.get(generation);
      if (!next) return;
      const previousGeneration = this.activeGeneration;
      const previousState = this.activeState;
      this.staged.delete(generation);
      this.activeGeneration = generation;
      this.activeState = next;
      committed = true;
      if (previousGeneration && previousGeneration !== generation) {
        this.scheduleRetirement(previousGeneration, previousState);
      }
    });
    return committed;
  }

  async abort(generation: string): Promise<void> {
    await this.enqueue(() => {
      if (!this.staged.delete(generation)) return;
      const state = this.generations.get(generation);
      this.generations.delete(generation);
      if (state && state !== this.activeState && !this.isReferenced(state)) {
        this.retire(state);
      }
    });
  }

  async lint(
    request: PluginLintRequest,
    signal?: AbortSignal,
  ): Promise<unknown> {
    if (this.disposed) return { results: [] };
    const generation = request.generation ?? this.activeGeneration;
    const state = generation
      ? this.generations.get(generation)
      : this.activeState;
    if (!state) {
      throw new Error(
        `unknown ESLint-plugin generation ${JSON.stringify(generation)}`,
      );
    }
    if (!state.host) {
      if (signal?.aborted) return { results: [] };
      throw new Error(
        `ESLint-plugin generation ${JSON.stringify(generation)} has no host`,
      );
    }
    state.leases++;
    try {
      return await state.host.lint(request, signal);
    } finally {
      state.leases--;
      if (state.retiring && state.leases === 0) this.startShutdown(state);
    }
  }

  async dispose(): Promise<void> {
    this.disposed = true;
    await this.enqueue(() => {
      for (const timer of this.retirementTimers.values()) clearTimeout(timer);
      this.retirementTimers.clear();
      this.generations.clear();
      this.staged.clear();
      this.activeGeneration = undefined;
      this.activeState = undefined;
      // Graceful generation retirement preserves in-flight lint leases, but a
      // whole sidecar shutdown has no future caller that can consume them.
      // Force every host to terminate so a wedged lease cannot keep worker
      // threads (and therefore the editor-runtime process) alive forever.
      for (const state of [...this.liveStates]) {
        state.retiring = true;
        this.startShutdown(state);
      }
    });
    await Promise.all(
      [...this.liveStates]
        .map((state) => state.shutdown)
        .filter(
          (shutdown): shutdown is Promise<void> => shutdown !== undefined,
        ),
    );
  }

  private scheduleRetirement(
    generation: string,
    state: HostState | undefined,
  ): void {
    const existing = this.retirementTimers.get(generation);
    if (existing) clearTimeout(existing);
    this.retirementTimers.set(
      generation,
      setTimeout(
        () => this.completeRetirement(generation, state),
        this.graceGenerationMs,
      ),
    );
    while (this.retirementTimers.size > this.maxGraceGenerations) {
      const oldest = this.retirementTimers.keys().next().value;
      if (oldest === undefined) break;
      this.completeRetirement(oldest, this.generations.get(oldest));
    }
  }

  private completeRetirement(
    generation: string,
    state: HostState | undefined,
  ): void {
    const timer = this.retirementTimers.get(generation);
    if (!timer) return;
    clearTimeout(timer);
    this.retirementTimers.delete(generation);
    if (generation === this.activeGeneration) return;
    if (this.generations.get(generation) !== state) return;
    this.generations.delete(generation);
    if (state && state !== this.activeState && !this.isReferenced(state)) {
      this.retire(state);
    }
  }

  private isReferenced(state: HostState): boolean {
    for (const candidate of this.generations.values()) {
      if (candidate === state) return true;
    }
    return false;
  }

  private retire(state: HostState): void {
    state.retiring = true;
    if (state.leases === 0) this.startShutdown(state);
  }

  private startShutdown(state: HostState): void {
    if (state.shutdown) return;
    state.shutdown = state.host
      ? Promise.resolve()
          .then(async () => state.host?.shutdown())
          .catch((error: unknown) => {
            this.report(
              `rslint: failed to shut down plugin host: ${String(error)}\n`,
            );
          })
      : Promise.resolve();
    const forgetState = (): void => {
      this.liveStates.delete(state);
    };
    void state.shutdown.then(forgetState, forgetState);
  }

  private async enqueue(operation: () => Promise<void> | void): Promise<void> {
    const run = this.operation.then(operation, operation);
    this.operation = run.catch(() => undefined);
    await run;
  }

  private assertOpen(): void {
    if (this.disposed) throw new Error('editor plugin pool is disposed');
  }

  private report(message: string): void {
    try {
      this.log(message);
    } catch {
      // Logging is observational and must never corrupt generation lifecycle.
    }
  }
}
