import { window, workspace, type TextDocument } from 'vscode';
import { type ResolvedCoreRuntime } from './CoreResolver';
import type {
  DocumentRoutingRuntime,
  WorkspaceDocumentRouter,
} from './WorkspaceDocumentRouter';
import { isSupportedWorkspaceDocument } from './WorkspaceDocumentRouter';

export interface ManagedRslintRuntime extends DocumentRoutingRuntime {
  start(signal: AbortSignal): Promise<void>;
  close(): Promise<void>;
}

export type ManagedRslintRuntimeFactory = (
  resolved: ResolvedCoreRuntime,
) => ManagedRslintRuntime;

export interface RuntimeCoreResolver {
  clear(): void;
  resolve(
    document: TextDocument,
    workspaceFolder: NonNullable<
      ReturnType<typeof workspace.getWorkspaceFolder>
    >,
    configuredPath?: string,
  ): Promise<ResolvedCoreRuntime>;
}

export interface RuntimeManagerLogger {
  debug(message: string, ...args: unknown[]): void;
  info(message: string, ...args: unknown[]): void;
  error(message: string, error?: unknown, ...args: unknown[]): void;
}

interface RuntimeEntry {
  readonly resolved: ResolvedCoreRuntime;
  readonly runtime: ManagedRslintRuntime;
  readonly abortController: AbortController;
  readonly users: Set<string>;
  startPromise: Promise<void>;
  active: boolean;
  closePromise?: Promise<void>;
}

function documentKey(document: TextDocument): string {
  return document.uri.toString();
}

function cancellationError(key: string): Error {
  const error = new Error(`Rslint runtime ${JSON.stringify(key)} was released`);
  error.name = 'AbortError';
  return error;
}

/**
 * Resolves the core for each open document and shares one runtime for every
 * physical core installation used inside the same VS Code workspace folder.
 * A replacement is started before its document leaves the last-good runtime.
 */
export class RuntimeManager {
  private readonly entries = new Map<string, RuntimeEntry>();
  private readonly closingRuntimes = new Map<string, Promise<void>>();
  private readonly bindings = new Map<string, RuntimeEntry>();
  private readonly documentEpochs = new Map<string, number>();
  private readonly documentTails = new Map<string, Promise<void>>();
  private readonly reportedFailures = new Set<string>();
  private closePromise: Promise<void> | undefined;
  private closing = false;

  public constructor(
    private readonly router: WorkspaceDocumentRouter,
    private readonly resolver: RuntimeCoreResolver,
    private readonly runtimeFactory: ManagedRslintRuntimeFactory,
    private readonly logger: RuntimeManagerLogger,
  ) {}

  public initialize(documents: readonly TextDocument[]): void {
    for (const document of documents) {
      void this.reconcile(document).catch((error: unknown) => {
        this.logger.error(`Failed to initialize ${document.uri}`, error);
      });
    }
  }

  public clearResolutionCache(): void {
    this.resolver.clear();
  }

  public async reconcileOpenDocuments(): Promise<void> {
    await Promise.allSettled(
      workspace.textDocuments.map(async (document) => this.reconcile(document)),
    );
  }

  public async reconcile(document: TextDocument): Promise<void> {
    if (this.closing) return;
    const key = documentKey(document);
    const epoch = this.nextDocumentEpoch(key);
    this.releasePendingDocumentUses(key);
    await this.enqueueDocument(key, async () => {
      if (!this.isCurrentDocument(document, epoch)) return;
      await this.reconcileCurrentDocument(document, epoch);
    });
  }

  /** Defer cleanup until LanguageClient's didClose middleware has run. */
  public documentClosed(document: TextDocument): void {
    const key = documentKey(document);
    const epoch = this.nextDocumentEpoch(key);
    this.releasePendingDocumentUses(key);
    setTimeout(() => {
      void this.enqueueDocument(key, async () => {
        if (this.documentEpochs.get(key) !== epoch) return;
        await this.detachDocument(document);
        this.documentEpochs.delete(key);
      }).catch((error: unknown) => {
        this.logger.error(`Failed to release ${document.uri}`, error);
      });
    }, 0);
  }

  public async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private async reconcileCurrentDocument(
    document: TextDocument,
    epoch: number,
  ): Promise<void> {
    const key = documentKey(document);
    const existing = this.bindings.get(key);
    const workspaceFolder = workspace.getWorkspaceFolder(document.uri);
    if (!isSupportedWorkspaceDocument(document) || !workspaceFolder) {
      await this.detachDocument(document);
      return;
    }
    const configuration = workspace.getConfiguration('rslint', document.uri);

    let resolved: ResolvedCoreRuntime;
    try {
      resolved = await this.resolver.resolve(
        document,
        workspaceFolder,
        configuration.get<string>('corePath'),
      );
    } catch (error) {
      if (this.isCurrentDocument(document, epoch)) {
        this.reportFailure(document, error, existing);
      }
      return;
    }
    if (!this.isCurrentDocument(document, epoch)) return;
    if (existing?.resolved.key === resolved.key) {
      return;
    }

    let replacement: RuntimeEntry | undefined;
    let switched = false;
    try {
      replacement = this.acquireRuntime(resolved, key);
      await replacement.startPromise;
      if (!this.isCurrentDocument(document, epoch)) {
        await this.releaseRuntimeAfterFailure(replacement, key);
        return;
      }
      await this.router.assign(document, resolved.key);
      this.bindings.set(key, replacement);
      switched = true;
      this.logger.debug(
        `Using ${resolved.installation.packageDirectory} for ${document.uri}`,
      );
    } catch (error) {
      if (replacement && !switched) {
        await this.releaseRuntimeAfterFailure(replacement, key);
      }
      if (this.isCurrentDocument(document, epoch)) {
        this.reportFailure(document, error, existing);
      }
      return;
    }

    // The handoff is committed once didOpen reaches the replacement. Cleanup
    // of the old process is forward-only: a shutdown failure must not withdraw
    // the now-owning runtime or leave the binding pointing at a closed process.
    if (existing) {
      void this.releaseRuntime(existing, key).catch((error: unknown) => {
        this.logger.error(
          `Failed to close superseded Rslint ${existing.resolved.installation.version}`,
          error,
        );
      });
    }
  }

  private acquireRuntime(
    resolved: ResolvedCoreRuntime,
    documentUri: string,
  ): RuntimeEntry {
    let entry = this.entries.get(resolved.key);
    if (!entry) {
      const runtime = this.runtimeFactory(resolved);
      const abortController = new AbortController();
      entry = {
        resolved,
        runtime,
        abortController,
        users: new Set(),
        active: false,
        startPromise: Promise.resolve(),
      };
      const startPromise = this.startRuntime(entry);
      entry.startPromise = startPromise;
      this.entries.set(resolved.key, entry);
      void startPromise.catch(() => undefined);
    }
    entry.users.add(documentUri);
    return entry;
  }

  private async startRuntime(
    entry: Omit<RuntimeEntry, 'startPromise'>,
  ): Promise<void> {
    await this.closingRuntimes.get(entry.resolved.key);
    await entry.runtime.start(entry.abortController.signal);
    if (this.closing || entry.abortController.signal.aborted) {
      throw cancellationError(entry.resolved.key);
    }
    await this.router.activate(entry.runtime);
    entry.active = true;
    this.logger.info(
      `Rslint ${entry.resolved.installation.version} ready for ${entry.resolved.workspaceFolder.name}`,
    );
  }

  private async releaseRuntime(
    entry: RuntimeEntry,
    documentUri: string,
  ): Promise<void> {
    entry.users.delete(documentUri);
    if (entry.users.size > 0) return;
    await this.closeRuntime(entry);
  }

  private async closeRuntime(entry: RuntimeEntry): Promise<void> {
    if (!entry.closePromise) {
      if (this.entries.get(entry.resolved.key) === entry) {
        this.entries.delete(entry.resolved.key);
      }
      const closePromise = (async () => {
        entry.abortController.abort(cancellationError(entry.resolved.key));
        await entry.startPromise.catch(() => undefined);
        const errors: unknown[] = [];
        if (entry.active) {
          try {
            await this.router.deactivate(entry.resolved.key);
          } catch (error) {
            errors.push(error);
          }
          entry.active = false;
        }
        try {
          await entry.runtime.close();
        } catch (error) {
          errors.push(error);
        }
        if (errors.length === 1) throw errors[0];
        if (errors.length > 1) {
          throw new AggregateError(errors, 'failed to close Rslint runtime');
        }
      })();
      entry.closePromise = closePromise;
      const previousBarrier = this.closingRuntimes.get(entry.resolved.key);
      const barrier = previousBarrier
        ? Promise.all([previousBarrier, closePromise]).then(() => undefined)
        : closePromise;
      this.closingRuntimes.set(entry.resolved.key, barrier);
      void barrier.then(
        () => {
          if (this.closingRuntimes.get(entry.resolved.key) === barrier) {
            this.closingRuntimes.delete(entry.resolved.key);
          }
        },
        () => undefined,
      );
    }
    await entry.closePromise;
  }

  private async releaseRuntimeAfterFailure(
    entry: RuntimeEntry,
    documentUri: string,
  ): Promise<void> {
    try {
      await this.releaseRuntime(entry, documentUri);
    } catch (error) {
      this.logger.error(
        `Failed to clean up unusable Rslint ${entry.resolved.installation.version}`,
        error,
      );
    }
  }

  private releasePendingDocumentUses(documentUri: string): void {
    const bound = this.bindings.get(documentUri);
    for (const entry of [...this.entries.values()]) {
      if (entry === bound || !entry.users.has(documentUri)) continue;
      void this.releaseRuntime(entry, documentUri).catch((error: unknown) => {
        this.logger.error(
          `Failed to cancel pending Rslint ${entry.resolved.installation.version}`,
          error,
        );
      });
    }
  }

  private async detachDocument(document: TextDocument): Promise<void> {
    const key = documentKey(document);
    const existing = this.bindings.get(key);
    await this.router.assign(document, undefined);
    this.bindings.delete(key);
    if (existing) await this.releaseRuntime(existing, key);
  }

  private reportFailure(
    document: TextDocument,
    error: unknown,
    existing: RuntimeEntry | undefined,
  ): void {
    const message = error instanceof Error ? error.message : String(error);
    this.logger.error(
      `Could not select Rslint core for ${document.uri}`,
      error,
    );
    const workspaceKey = workspace
      .getWorkspaceFolder(document.uri)
      ?.uri.toString();
    const category =
      error instanceof Error && error.name === 'CoreNotFoundError'
        ? 'core-not-found'
        : message;
    const failureKey = `${workspaceKey ?? 'loose-file'}\0${category}`;
    if (this.reportedFailures.has(failureKey)) return;
    this.reportedFailures.add(failureKey);
    const suffix = existing
      ? ` Keeping Rslint ${existing.resolved.installation.version} active.`
      : '';
    void window.showWarningMessage(`Rslint: ${message}${suffix}`);
  }

  private isCurrentDocument(document: TextDocument, epoch: number): boolean {
    return (
      !this.closing &&
      this.documentEpochs.get(documentKey(document)) === epoch &&
      workspace.textDocuments.includes(document)
    );
  }

  private nextDocumentEpoch(key: string): number {
    const epoch = (this.documentEpochs.get(key) ?? 0) + 1;
    this.documentEpochs.set(key, epoch);
    return epoch;
  }

  private async enqueueDocument(
    key: string,
    operation: () => Promise<void>,
  ): Promise<void> {
    const previous = this.documentTails.get(key) ?? Promise.resolve();
    const run = previous.then(operation, operation);
    const tail = run.catch(() => undefined);
    this.documentTails.set(key, tail);
    try {
      await run;
    } finally {
      if (this.documentTails.get(key) === tail) {
        this.documentTails.delete(key);
      }
    }
  }

  private async closeImpl(): Promise<void> {
    if (this.closing) return;
    this.closing = true;
    for (const key of this.documentEpochs.keys()) this.nextDocumentEpoch(key);
    for (const entry of this.entries.values()) {
      entry.abortController.abort(cancellationError(entry.resolved.key));
    }

    const errors: unknown[] = [];
    try {
      await this.router.closeAll();
    } catch (error) {
      errors.push(error);
    }
    const closingBeforeShutdown = [...this.closingRuntimes.values()];
    const results = await Promise.allSettled([
      ...closingBeforeShutdown,
      ...[...this.entries.values()].map(async (entry) =>
        this.closeRuntime(entry),
      ),
    ]);
    for (const result of results) {
      if (result.status === 'rejected') errors.push(result.reason);
    }
    this.entries.clear();
    this.closingRuntimes.clear();
    this.bindings.clear();
    this.documentEpochs.clear();
    this.reportedFailures.clear();
    if (errors.length > 0) {
      throw new AggregateError(errors, 'failed to close runtime manager');
    }
  }
}
