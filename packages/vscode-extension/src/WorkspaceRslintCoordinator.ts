import type { WorkspaceFolder, WorkspaceFoldersChangeEvent } from 'vscode';
import type { DocumentRoutingRuntime } from './WorkspaceDocumentRouter';

export interface WorkspaceRuntime extends DocumentRoutingRuntime {
  start(signal: AbortSignal): Promise<void>;
  close(): Promise<void>;
}

export type WorkspaceRuntimeFactory = (
  folder: WorkspaceFolder,
  rootKey: string,
) => WorkspaceRuntime;

export interface WorkspaceRootRouter {
  activate(runtime: DocumentRoutingRuntime): Promise<void>;
  deactivate(rootKey: string): Promise<void>;
  closeAll(): Promise<void>;
}

export interface WorkspaceCoordinatorLogger {
  debug(message: string, ...args: unknown[]): void;
  info(message: string, ...args: unknown[]): void;
  warn(message: string, ...args: unknown[]): void;
  error(message: string, error?: unknown, ...args: unknown[]): void;
}

interface DesiredRoot {
  readonly folder: WorkspaceFolder;
  readonly generation: number;
  readonly readiness: Deferred<void>;
}

interface CurrentRuntime {
  readonly generation: number;
  readonly runtime: WorkspaceRuntime;
  readonly abortController: AbortController;
  phase: 'starting' | 'active' | 'closing' | 'close-failed';
  closeError?: unknown;
  closeFailureRecorded?: boolean;
}

interface RootSlot {
  readonly key: string;
  current?: CurrentRuntime;
  failedGeneration?: number;
  worker?: Promise<void>;
  rerun: boolean;
}

interface Deferred<T> {
  readonly promise: Promise<T>;
  resolve(value: T | PromiseLike<T>): void;
  reject(reason?: unknown): void;
}

function deferred<T>(): Deferred<T> {
  let resolvePromise!: (value: T | PromiseLike<T>) => void;
  let rejectPromise!: (reason?: unknown) => void;
  let settled = false;
  const promise = new Promise<T>((resolve, reject) => {
    resolvePromise = resolve;
    rejectPromise = reject;
  });
  // Root readiness is also observed by dynamic fire-and-forget reconciliation.
  // Keep a rejection handler attached even when no activation caller awaits it.
  void promise.catch(() => undefined);
  return {
    promise,
    resolve(value) {
      if (settled) return;
      settled = true;
      resolvePromise(value);
    },
    reject(reason) {
      if (settled) return;
      settled = true;
      rejectPromise(reason);
    },
  };
}

function cancellationError(rootKey: string): Error {
  const error = new Error(
    `workspace root ${JSON.stringify(rootKey)} was superseded`,
  );
  error.name = 'AbortError';
  return error;
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError';
}

function errorReasons(error: unknown): unknown[] {
  if (!(error instanceof AggregateError)) return [error];
  const reasons: readonly unknown[] = error.errors;
  return [...reasons];
}

function folderMetadataChanged(
  previous: WorkspaceFolder,
  next: WorkspaceFolder,
): boolean {
  // index is positional metadata and routinely changes when an unrelated root
  // is inserted before this one. It is neither identity nor a restart reason.
  return previous.name !== next.name;
}

export function workspaceRootKey(folder: WorkspaceFolder): string {
  return folder.uri.toString();
}

/**
 * Reconciles VS Code workspace-folder identity with independently-lived root
 * runtimes. It never serializes different roots behind config evaluation or
 * worker shutdown; only each URI slot is ordered.
 */
export class WorkspaceRslintCoordinator {
  private readonly desiredRoots = new Map<string, DesiredRoot>();
  private readonly slots = new Map<string, RootSlot>();
  private readonly generations = new Map<string, number>();
  private readonly terminalCloseErrors: unknown[] = [];
  private topologyChanged = deferred<void>();
  private closePromise: Promise<void> | undefined;
  private closing = false;

  public constructor(
    private readonly router: WorkspaceRootRouter,
    private readonly runtimeFactory: WorkspaceRuntimeFactory,
    private readonly logger: WorkspaceCoordinatorLogger,
  ) {}

  public async initialize(folders: readonly WorkspaceFolder[]): Promise<void> {
    this.reconcile(folders);
    await this.waitForAnyDesiredRoot();
  }

  public handleWorkspaceFoldersChanged(
    event: WorkspaceFoldersChangeEvent,
    folders: readonly WorkspaceFolder[],
  ): void {
    if (this.closing) return;
    const removedKeys = new Set<string>();
    for (const folder of event.removed) {
      removedKeys.add(workspaceRootKey(folder));
    }
    const forceReplace = new Set<string>();
    for (const folder of event.added) {
      const key = workspaceRootKey(folder);
      // A rename/remove+add replacement can preserve the URI. A plain added
      // event for a URI already captured by the activation snapshot is not a
      // replacement and must not abort that in-flight initial generation.
      if (removedKeys.has(key)) forceReplace.add(key);
    }
    this.reconcile(folders, forceReplace);
  }

  public async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private reconcile(
    folders: readonly WorkspaceFolder[],
    forceReplace: ReadonlySet<string> = new Set(),
  ): void {
    if (this.closing) return;
    const nextFolders = new Map(
      folders.map((folder) => [workspaceRootKey(folder), folder]),
    );
    const changedKeys = new Set<string>();

    for (const [key, desired] of this.desiredRoots) {
      const next = nextFolders.get(key);
      if (!next) {
        this.desiredRoots.delete(key);
        this.nextGeneration(key);
        desired.readiness.reject(cancellationError(key));
        changedKeys.add(key);
        this.slots
          .get(key)
          ?.current?.abortController.abort(cancellationError(key));
        continue;
      }
      if (
        forceReplace.has(key) ||
        folderMetadataChanged(desired.folder, next)
      ) {
        desired.readiness.reject(cancellationError(key));
        const replacement = this.createDesiredRoot(next);
        this.desiredRoots.set(key, replacement);
        changedKeys.add(key);
        this.slots
          .get(key)
          ?.current?.abortController.abort(cancellationError(key));
      }
      nextFolders.delete(key);
    }

    for (const [key, folder] of nextFolders) {
      this.desiredRoots.set(key, this.createDesiredRoot(folder));
      changedKeys.add(key);
    }

    for (const key of changedKeys) this.kick(key);
    if (changedKeys.size > 0) this.signalTopologyChanged();
  }

  private createDesiredRoot(folder: WorkspaceFolder): DesiredRoot {
    return {
      folder,
      generation: this.nextGeneration(workspaceRootKey(folder)),
      readiness: deferred<void>(),
    };
  }

  private async waitForAnyDesiredRoot(): Promise<void> {
    for (;;) {
      const snapshot = [...this.desiredRoots.entries()];
      if (snapshot.length === 0) return;
      // Resolve as soon as one independent root is usable. A pending root
      // cannot undo another root's successful activation.
      const readiness: Promise<void>[] = [];
      for (const [, desired] of snapshot) {
        readiness.push(desired.readiness.promise);
      }
      const result = await Promise.race([
        Promise.any(readiness).then(
          () => ({ kind: 'ready' as const }),
          (error: unknown) => ({ kind: 'failed' as const, error }),
        ),
        this.topologyChanged.promise.then(() => ({
          kind: 'topology' as const,
        })),
      ]);
      if (result.kind === 'topology') continue;
      if (result.kind === 'ready') {
        return;
      }
      if (this.closing) throw result.error;
      const topologyChanged =
        snapshot.length !== this.desiredRoots.size ||
        snapshot.some(
          ([key, desired]) =>
            this.desiredRoots.get(key)?.generation !== desired.generation,
        );
      // Folder events are installed before initialization. If that topology
      // superseded every promise in this snapshot, observe the new desired
      // generations instead of treating cancellation as an activation
      // failure.
      if (topologyChanged) continue;

      const reasons = errorReasons(result.error);
      throw new AggregateError(
        reasons,
        `All Rslint workspace roots failed: ${reasons
          .map((reason) =>
            reason instanceof Error ? reason.message : String(reason),
          )
          .join('; ')}`,
      );
    }
  }

  private signalTopologyChanged(): void {
    const previous = this.topologyChanged;
    this.topologyChanged = deferred<void>();
    previous.resolve(undefined);
  }

  private nextGeneration(rootKey: string): number {
    const generation = (this.generations.get(rootKey) ?? 0) + 1;
    this.generations.set(rootKey, generation);
    return generation;
  }

  private kick(rootKey: string): void {
    let slot = this.slots.get(rootKey);
    if (!slot) {
      slot = { key: rootKey, rerun: false };
      this.slots.set(rootKey, slot);
    }
    slot.rerun = true;
    if (slot.worker) return;
    slot.worker = this.runSlot(slot).finally(() => {
      slot.worker = undefined;
      if (slot.rerun && !this.closing) {
        this.kick(rootKey);
      } else if (!slot.current && !this.desiredRoots.has(rootKey)) {
        this.slots.delete(rootKey);
        this.generations.delete(rootKey);
      }
    });
  }

  private async runSlot(slot: RootSlot): Promise<void> {
    while (slot.rerun || this.slotNeedsReconcile(slot)) {
      slot.rerun = false;
      const desired = this.desiredRoots.get(slot.key);
      const current = slot.current;

      if (
        current &&
        (!desired || desired.generation !== current.generation || this.closing)
      ) {
        if (!(await this.closeCurrent(slot, current))) return;
        continue;
      }

      if (!current && desired && !this.closing) {
        if (slot.failedGeneration === desired.generation) return;
        if (!(await this.startDesired(slot, desired))) return;
        continue;
      }

      return;
    }
  }

  private slotNeedsReconcile(slot: RootSlot): boolean {
    const desired = this.desiredRoots.get(slot.key);
    const current = slot.current;
    if (this.closing) return current !== undefined;
    if (!current) {
      return !!desired && slot.failedGeneration !== desired.generation;
    }
    return !desired || desired.generation !== current.generation;
  }

  private async startDesired(
    slot: RootSlot,
    desired: DesiredRoot,
  ): Promise<boolean> {
    const abortController = new AbortController();
    let runtime: WorkspaceRuntime;
    try {
      runtime = this.runtimeFactory(desired.folder, slot.key);
    } catch (error) {
      slot.failedGeneration = desired.generation;
      desired.readiness.reject(error);
      this.logger.error(`Failed to create Rslint workspace ${slot.key}`, error);
      return true;
    }
    const current: CurrentRuntime = {
      generation: desired.generation,
      runtime,
      abortController,
      phase: 'starting',
    };
    slot.current = current;
    this.logger.debug(
      `Starting Rslint workspace ${slot.key} generation ${desired.generation}`,
    );

    try {
      await runtime.start(abortController.signal);
      if (
        this.closing ||
        abortController.signal.aborted ||
        this.desiredRoots.get(slot.key)?.generation !== desired.generation
      ) {
        throw cancellationError(slot.key);
      }
      await this.router.activate(runtime);
      if (
        this.closing ||
        abortController.signal.aborted ||
        this.desiredRoots.get(slot.key)?.generation !== desired.generation
      ) {
        await this.router.deactivate(slot.key).catch((error: unknown) => {
          this.logger.error(
            `Failed to withdraw stale workspace ${slot.key}`,
            error,
          );
        });
        throw cancellationError(slot.key);
      }
      current.phase = 'active';
      desired.readiness.resolve(undefined);
      this.logger.info(`Rslint workspace ready: ${slot.key}`);
      return true;
    } catch (error) {
      const stale =
        isAbortError(error) ||
        abortController.signal.aborted ||
        this.desiredRoots.get(slot.key)?.generation !== desired.generation ||
        this.closing;
      if (!stale) {
        slot.failedGeneration = desired.generation;
        desired.readiness.reject(error);
        this.logger.error(
          `Failed to start Rslint workspace ${slot.key}`,
          error,
        );
      } else {
        desired.readiness.reject(cancellationError(slot.key));
      }
      return this.closeCurrent(slot, current);
    }
  }

  private async closeCurrent(
    slot: RootSlot,
    current: CurrentRuntime,
  ): Promise<boolean> {
    if (slot.current !== current) return true;
    current.abortController.abort(cancellationError(slot.key));
    if (current.phase === 'active') {
      try {
        await this.router.deactivate(slot.key);
      } catch (error) {
        this.logger.error(
          `Failed to transfer documents away from ${slot.key}`,
          error,
        );
      }
    }
    current.phase = 'closing';
    try {
      await current.runtime.close();
    } catch (error) {
      current.phase = 'close-failed';
      current.closeError = error;
      this.logger.error(`Failed to close Rslint workspace ${slot.key}`, error);
      if (!current.closeFailureRecorded) {
        current.closeFailureRecorded = true;
        this.terminalCloseErrors.push(error);
      }

      // A runtime whose resources did not close remains the sole owner of its
      // URI slot. Starting a replacement here could overlap native processes,
      // workers, watchers, and diagnostics for one workspace. Quarantine it
      // until an explicit later reconciliation (or terminal close) retries.
      const replacement = this.desiredRoots.get(slot.key);
      if (replacement && replacement.generation !== current.generation) {
        slot.failedGeneration = replacement.generation;
        replacement.readiness.reject(
          new Error(
            `Could not replace Rslint workspace ${JSON.stringify(slot.key)} because the previous runtime failed to close`,
            { cause: error },
          ),
        );
      }
      return false;
    }
    if (slot.current === current) slot.current = undefined;
    this.logger.debug(`Closed Rslint workspace ${slot.key}`);
    return true;
  }

  private async closeImpl(): Promise<void> {
    if (this.closing) return;
    this.closing = true;
    for (const [key, desired] of this.desiredRoots) {
      desired.readiness.reject(cancellationError(key));
    }
    this.desiredRoots.clear();
    for (const slot of this.slots.values()) {
      slot.current?.abortController.abort(cancellationError(slot.key));
      slot.rerun = true;
    }

    const routerResult = await Promise.allSettled([
      Promise.resolve().then(async () => {
        await this.router.closeAll();
      }),
    ]);
    for (const slot of this.slots.values()) this.kick(slot.key);
    const slotPromises: Promise<void>[] = [];
    for (const slot of this.slots.values()) {
      slotPromises.push(slot.worker ?? Promise.resolve());
    }
    const slotResults = await Promise.allSettled(slotPromises);
    this.slots.clear();

    const errors = this.terminalCloseErrors.splice(0);
    for (const result of [...routerResult, ...slotResults]) {
      if (result.status === 'rejected') {
        const reason: unknown = result.reason;
        errors.push(reason);
      }
    }
    if (errors.length > 0) {
      throw new AggregateError(errors, 'failed to close workspace coordinator');
    }
  }
}
