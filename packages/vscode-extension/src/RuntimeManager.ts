import path from 'node:path';
import { statSync } from 'node:fs';

import {
  RelativePattern,
  Uri,
  window,
  workspace,
  type Disposable,
  type OutputChannel,
  type TextDocument,
} from 'vscode';

import { EditorRuntimeClient } from './EditorRuntimeClient';
import type { Logger } from './logger';
import {
  resolveRuntimeForDocument,
  type ResolvedRuntime,
} from './RuntimeResolver';
import {
  isSupportedDocument,
  RuntimeDocumentRouter,
} from './RuntimeDocumentRouter';

const TOPOLOGY_GLOB =
  '**/{package.json,package-lock.json,npm-shrinkwrap.json,pnpm-lock.yaml,pnpm-workspace.yaml,yarn.lock,bun.lock,bun.lockb,.pnp.cjs,.pnp.js,.pnp.loader.mjs,.pnp.data.json}';

interface RuntimeSlot {
  readonly descriptor: ResolvedRuntime;
  readonly client: EditorRuntimeClient;
  readonly ready: Promise<void>;
  readonly documents: Set<string>;
  readonly invalidationWatchers: Disposable[];
  reservations: number;
  closing?: Promise<void>;
}

// Core bounds config work at 30s and may spend up to 5s aborting a timed-out
// transaction. Leave enough room for that cleanup before abandoning a
// speculative replacement and retaining the current owner.
const RUNTIME_MIGRATION_READY_TIMEOUT_MS = 40_000;

export interface RuntimeManagerOptions {
  readonly outputChannel: OutputChannel;
  readonly traceOutputChannel: OutputChannel;
  readonly logger: Logger;
}

/** Resolves documents independently and pools them by the actual core entry. */
export class RuntimeManager {
  private readonly router = new RuntimeDocumentRouter();
  private readonly runtimes = new Map<string, RuntimeSlot>();
  private readonly closingRuntimes = new Map<string, Promise<void>>();
  private readonly documentRuntimes = new Map<string, string>();
  private readonly documentInstances = new Map<string, TextDocument>();
  private readonly documentGenerations = new Map<string, number>();
  private readonly disposables: Disposable[] = [];
  private readonly configuredPathWatchers: Disposable[] = [];
  private readonly warnedFolders = new Set<string>();
  private topologyTimer: ReturnType<typeof setTimeout> | undefined;
  private closePromise: Promise<void> | undefined;
  private closing = false;

  constructor(private readonly options: RuntimeManagerOptions) {}

  async start(): Promise<void> {
    this.refreshConfiguredPathWatchers();
    const watcher = workspace.createFileSystemWatcher(TOPOLOGY_GLOB);
    const topologyChanged = () => this.scheduleReconcile();
    this.disposables.push(watcher);
    this.disposables.push(watcher.onDidCreate(topologyChanged));
    this.disposables.push(watcher.onDidChange(topologyChanged));
    this.disposables.push(watcher.onDidDelete(topologyChanged));
    this.disposables.push(
      workspace.onDidOpenTextDocument((document) => {
        void this.reconcileDocumentSafely(document);
      }),
      workspace.onDidCloseTextDocument((document) => {
        void this.releaseDocumentSafely(document);
      }),
      workspace.onDidChangeWorkspaceFolders(() => {
        this.refreshConfiguredPathWatchers();
        this.scheduleReconcile();
      }),
      workspace.onDidChangeConfiguration((event) => {
        if (
          event.affectsConfiguration('rslint.runtime.path') ||
          event.affectsConfiguration('rslint.enable')
        ) {
          if (event.affectsConfiguration('rslint.runtime.path')) {
            this.refreshConfiguredPathWatchers();
          }
          this.scheduleReconcile();
        }
      }),
    );

    await Promise.all(
      workspace.textDocuments
        .filter(isSupportedDocument)
        .map(async (document) => this.reconcileDocumentSafely(document)),
    );
  }

  async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private scheduleReconcile(): void {
    if (this.closing) return;
    // A topology/settings change is the point at which a previously missing
    // installation may have appeared (or disappeared). Reset warning
    // suppression once per change, not once per successfully resolved file.
    this.warnedFolders.clear();
    clearTimeout(this.topologyTimer);
    this.topologyTimer = setTimeout(() => {
      this.topologyTimer = undefined;
      void this.reconcileAll().catch((error: unknown) => {
        this.options.logger.error('Failed to reconcile Rslint runtimes', error);
      });
    }, 300);
  }

  private async reconcileAll(): Promise<void> {
    if (this.closing) return;
    await Promise.all(
      workspace.textDocuments
        .filter(isSupportedDocument)
        .map(async (document) => this.reconcileDocumentSafely(document)),
    );
    const open = new Set(
      workspace.textDocuments.map((doc) => doc.uri.toString()),
    );
    for (const uri of [...this.documentRuntimes.keys()]) {
      if (!open.has(uri)) await this.releaseDocumentByUri(uri);
    }
  }

  private async reconcileDocument(document: TextDocument): Promise<void> {
    if (this.closing || !isSupportedDocument(document)) return;
    const uri = document.uri.toString();
    this.documentInstances.set(uri, document);
    const generation = (this.documentGenerations.get(uri) ?? 0) + 1;
    this.documentGenerations.set(uri, generation);

    const folder = workspace.getWorkspaceFolder(document.uri);
    const enabled =
      folder !== undefined &&
      workspace
        .getConfiguration('rslint', folder.uri)
        .get<boolean>('enable', true);
    let descriptor: ResolvedRuntime | undefined;
    try {
      if (enabled) descriptor = await resolveRuntimeForDocument(document);
    } catch (error) {
      this.options.logger.error(
        `Failed to resolve @rslint/core for ${document.uri.fsPath}`,
        error,
      );
      this.warnResolution(folder?.uri);
      // Package managers commonly replace manifests and entry files in
      // separate filesystem events. Keep a running last-good assignment on a
      // transient/incomplete generation; an actual missing install resolves
      // cleanly to undefined and follows the removal path below.
      return;
    }
    if (
      this.closing ||
      this.documentGenerations.get(uri) !== generation ||
      this.documentInstances.get(uri) !== document ||
      !workspace.textDocuments.includes(document)
    ) {
      return;
    }
    if (!descriptor) {
      if (enabled) this.warnResolution(folder?.uri);
      await this.moveDocument(document, undefined);
      return;
    }

    const previousKey = this.documentRuntimes.get(uri);
    const slot = await this.getOrCreateRuntime(descriptor);
    slot.reservations++;
    try {
      if (previousKey && previousKey !== descriptor.key) {
        await this.waitForMigrationReady(slot);
      }
      if (
        this.closing ||
        this.documentGenerations.get(uri) !== generation ||
        this.documentInstances.get(uri) !== document ||
        !workspace.textDocuments.includes(document) ||
        slot.closing !== undefined ||
        this.runtimes.get(descriptor.key) !== slot
      ) {
        return;
      }
      await this.moveDocument(document, slot);
    } finally {
      slot.reservations--;
      if (slot.reservations === 0 && slot.documents.size === 0) {
        void this.closeRuntime(slot);
      }
    }
  }

  private async reconcileDocumentSafely(document: TextDocument): Promise<void> {
    try {
      await this.reconcileDocument(document);
    } catch (error) {
      this.options.logger.error(
        `Failed to route ${document.uri.fsPath} to @rslint/core`,
        error,
      );
    }
  }

  private async releaseDocumentSafely(document: TextDocument): Promise<void> {
    try {
      await this.releaseDocument(document);
    } catch (error) {
      this.options.logger.error(
        `Failed to release Rslint runtime for ${document.uri.fsPath}`,
        error,
      );
    }
  }

  private async getOrCreateRuntime(
    descriptor: ResolvedRuntime,
  ): Promise<RuntimeSlot> {
    const closing = this.closingRuntimes.get(descriptor.key);
    if (closing) await closing;
    if (this.closing) throw new Error('runtime manager is closing');
    return this.runtimes.get(descriptor.key) ?? this.createRuntime(descriptor);
  }

  private createRuntime(descriptor: ResolvedRuntime): RuntimeSlot {
    let slot: RuntimeSlot | undefined;
    const client = new EditorRuntimeClient({
      descriptor,
      router: this.router,
      outputChannel: this.options.outputChannel,
      traceOutputChannel: this.options.traceOutputChannel,
      logger: this.options.logger,
      onTerminalFailure: () => {
        if (!slot) return;
        this.reportRuntimeFailure(slot, 'stopped after repeated failures');
      },
    });
    this.router.register(client);
    const start = client.start();
    // Publish before observing start: another document resolving in the same
    // tick must share this exact process even while initialize is pending.
    const createdSlot: RuntimeSlot = {
      descriptor,
      client,
      ready: start,
      documents: new Set(),
      invalidationWatchers: this.watchRuntimeFiles(descriptor),
      reservations: 0,
    };
    slot = createdSlot;
    this.runtimes.set(descriptor.key, createdSlot);
    void start.catch((error: unknown) => {
      // A zero-reference runtime may be intentionally closed while its initial
      // initialize request is still pending. That cancellation is not a user-
      // visible startup failure.
      if (
        this.closing ||
        createdSlot.closing !== undefined ||
        this.runtimes.get(descriptor.key) !== createdSlot
      ) {
        return;
      }
      this.reportRuntimeFailure(createdSlot, error);
    });
    this.options.logger.info(
      `Using @rslint/core${descriptor.version ? ` ${descriptor.version}` : ''} from ${descriptor.packagePath}`,
    );
    return createdSlot;
  }

  private async waitForMigrationReady(slot: RuntimeSlot): Promise<void> {
    let timer: ReturnType<typeof setTimeout> | undefined;
    let cancelled = false;
    const waitUntilRunning = async (): Promise<void> => {
      await slot.ready;
      while (!cancelled && !slot.client.isRunning()) {
        if (
          this.closing ||
          slot.closing !== undefined ||
          this.runtimes.get(slot.descriptor.key) !== slot
        ) {
          throw new Error(
            `replacement runtime stopped before becoming ready: ${slot.descriptor.entryPath}`,
          );
        }
        await new Promise((resolve) => setTimeout(resolve, 25));
      }
    };
    try {
      await Promise.race([
        waitUntilRunning(),
        new Promise<never>((_, reject) => {
          timer = setTimeout(() => {
            reject(
              new Error(
                `timed out waiting for replacement runtime ${slot.descriptor.entryPath}`,
              ),
            );
          }, RUNTIME_MIGRATION_READY_TIMEOUT_MS);
        }),
      ]);
    } finally {
      cancelled = true;
      clearTimeout(timer);
    }
  }

  private watchRuntimeFiles(descriptor: ResolvedRuntime): Disposable[] {
    const externalPaths = new Set(
      descriptor.watchPaths.map((filePath) => path.resolve(filePath)),
    );
    const targetsByDirectory = new Map<string, Set<string>>();
    for (const filePath of externalPaths) {
      const directory = path.dirname(filePath);
      const targets = targetsByDirectory.get(directory) ?? new Set<string>();
      targets.add(runtimeFileIdentity(filePath));
      targetsByDirectory.set(directory, targets);
    }

    const watchers: Disposable[] = [];
    for (const [directory, targets] of targetsByDirectory) {
      const watcher = workspace.createFileSystemWatcher(
        new RelativePattern(Uri.file(directory), '*'),
      );
      const changed = (uri: Uri): void => {
        if (targets.has(runtimeFileIdentity(uri.fsPath))) {
          this.scheduleReconcile();
        }
      };
      watcher.onDidCreate(changed);
      watcher.onDidChange(changed);
      watcher.onDidDelete(changed);
      watchers.push(watcher);
    }
    return watchers;
  }

  private refreshConfiguredPathWatchers(): void {
    for (const watcher of this.configuredPathWatchers.splice(0)) {
      watcher.dispose();
    }
    for (const folder of workspace.workspaceFolders ?? []) {
      if (folder.uri.scheme !== 'file') continue;
      const configuredPath = workspace
        .getConfiguration('rslint', folder.uri)
        .get<string>('runtime.path', '')
        .trim();
      if (!configuredPath) continue;
      const packageDirectory = path.resolve(folder.uri.fsPath, configuredPath);

      // This watcher exists before package resolution succeeds, so a custom
      // directory that is created or built in several writes can recover
      // without toggling the setting. Base it at the nearest existing ancestor:
      // some hosts cannot watch a missing parent, and package managers commonly
      // create several missing path components in one operation.
      const watchBase = nearestExistingDirectory(
        path.dirname(packageDirectory),
      );
      const firstMissingComponent = path
        .relative(watchBase, packageDirectory)
        .split(path.sep)
        .find(Boolean);
      const watchedChild = firstMissingComponent
        ? path.join(watchBase, firstMissingComponent)
        : packageDirectory;
      const parentWatcher = workspace.createFileSystemWatcher(
        new RelativePattern(Uri.file(watchBase), '*'),
      );
      const parentChanged = (uri: Uri): void => {
        if (
          runtimeFileIdentity(uri.fsPath) === runtimeFileIdentity(watchedChild)
        ) {
          this.refreshConfiguredPathWatchers();
          this.scheduleReconcile();
        }
      };
      parentWatcher.onDidCreate(parentChanged);
      parentWatcher.onDidChange(parentChanged);
      parentWatcher.onDidDelete(parentChanged);

      const contentWatcher = workspace.createFileSystemWatcher(
        new RelativePattern(Uri.file(packageDirectory), '**/*'),
      );
      const contentChanged = (): void => this.scheduleReconcile();
      contentWatcher.onDidCreate(contentChanged);
      contentWatcher.onDidChange(contentChanged);
      contentWatcher.onDidDelete(contentChanged);
      this.configuredPathWatchers.push(parentWatcher, contentWatcher);
    }
  }

  private reportRuntimeFailure(slot: RuntimeSlot, error: unknown): void {
    if (
      this.closing ||
      slot.closing !== undefined ||
      this.runtimes.get(slot.descriptor.key) !== slot
    ) {
      return;
    }
    this.options.logger.error(
      `Rslint runtime failed (${slot.descriptor.entryPath})`,
      error,
    );
    void window.showErrorMessage(
      `Rslint could not run @rslint/core${slot.descriptor.version ? ` ${slot.descriptor.version}` : ''}. See the Rslint output channel.`,
    );
    void this.failRuntime(slot);
  }

  private async moveDocument(
    document: TextDocument,
    next: RuntimeSlot | undefined,
  ): Promise<void> {
    const uri = document.uri.toString();
    const previousKey = this.documentRuntimes.get(uri);
    if (previousKey === next?.descriptor.key) {
      // Assignment identity can survive a failed didOpen or a server restart.
      // Re-entering the router is idempotent and repairs a missing open session.
      await this.router.assign(document, next?.client);
      return;
    }
    const previous = previousKey ? this.runtimes.get(previousKey) : undefined;

    if (next) {
      next.documents.add(uri);
      this.documentRuntimes.set(uri, next.descriptor.key);
    } else {
      this.documentRuntimes.delete(uri);
    }
    let committed = false;
    try {
      await this.router.assign(document, next?.client);
    } finally {
      committed = this.router.isAssignedTo(document, next?.client);
      if (!committed) {
        next?.documents.delete(uri);
        if (previous) this.documentRuntimes.set(uri, previous.descriptor.key);
        else this.documentRuntimes.delete(uri);
      } else if (previous) {
        previous.documents.delete(uri);
        if (previous.documents.size === 0) void this.closeRuntime(previous);
      }
    }
  }

  private async releaseDocument(document: TextDocument): Promise<void> {
    const uri = document.uri.toString();
    // A close callback for an old TextDocument can finish after VS Code has
    // reopened the same URI. Never let that stale callback detach the new
    // document's runtime assignment.
    if (this.documentInstances.get(uri) !== document) return;
    this.documentInstances.delete(uri);
    const generation = (this.documentGenerations.get(uri) ?? 0) + 1;
    this.documentGenerations.set(uri, generation);
    const key = this.documentRuntimes.get(uri);
    this.documentRuntimes.delete(uri);
    const slot = key ? this.runtimes.get(key) : undefined;
    slot?.documents.delete(uri);
    await this.router.assign(document, undefined).catch((error: unknown) => {
      this.options.logger.error('Failed to release routed document', error);
    });
    if (this.documentGenerations.get(uri) === generation) {
      this.documentGenerations.delete(uri);
    }
    if (slot?.documents.size === 0) await this.closeRuntime(slot);
  }

  private async releaseDocumentByUri(uri: string): Promise<void> {
    if (
      workspace.textDocuments.some(
        (document) => document.uri.toString() === uri,
      )
    ) {
      return;
    }
    const document = this.documentInstances.get(uri);
    this.documentInstances.delete(uri);
    const key = this.documentRuntimes.get(uri);
    this.documentRuntimes.delete(uri);
    this.documentGenerations.delete(uri);
    if (document) {
      await this.router.assign(document, undefined).catch((error: unknown) => {
        this.options.logger.error('Failed to release routed document', error);
      });
    }
    if (!key) return;
    const slot = this.runtimes.get(key);
    if (!slot) return;
    slot.documents.delete(uri);
    if (slot.documents.size === 0) await this.closeRuntime(slot);
  }

  private async failRuntime(slot: RuntimeSlot): Promise<void> {
    if (this.runtimes.get(slot.descriptor.key) !== slot) return;
    for (const uri of slot.documents) {
      if (this.documentRuntimes.get(uri) === slot.descriptor.key) {
        this.documentRuntimes.delete(uri);
      }
    }
    slot.documents.clear();
    await this.closeRuntime(slot);
  }

  private async closeRuntime(slot: RuntimeSlot): Promise<void> {
    if ((slot.documents.size > 0 || slot.reservations > 0) && !this.closing) {
      return;
    }
    if (slot.closing) return slot.closing;
    if (this.runtimes.get(slot.descriptor.key) === slot) {
      this.runtimes.delete(slot.descriptor.key);
    }
    for (const watcher of slot.invalidationWatchers.splice(0)) {
      watcher.dispose();
    }
    slot.closing = (async () => {
      const results = await Promise.allSettled([
        this.router.resetServerSession(slot.client),
        slot.client.close(),
      ]);
      for (const result of results) {
        if (result.status === 'rejected') {
          this.options.logger.error(
            'Failed to close Rslint runtime',
            result.reason,
          );
        }
      }
      try {
        await this.router.unregister(slot.client);
      } catch (error) {
        this.options.logger.error('Failed to unregister Rslint runtime', error);
      }
    })();
    this.closingRuntimes.set(slot.descriptor.key, slot.closing);
    const forgetClosingRuntime = (): void => {
      if (this.closingRuntimes.get(slot.descriptor.key) === slot.closing) {
        this.closingRuntimes.delete(slot.descriptor.key);
      }
    };
    void slot.closing.then(forgetClosingRuntime, forgetClosingRuntime);
    return slot.closing;
  }

  private warnResolution(folderUri: Uri | undefined): void {
    if (!folderUri) return;
    const key = folderUri.toString();
    if (this.warnedFolders.has(key)) return;
    this.warnedFolders.add(key);
    void window.showWarningMessage(
      'Rslint: no local @rslint/core with an editor runtime was found for this file. Install @rslint/core in the project or set rslint.runtime.path.',
    );
  }

  private async closeImpl(): Promise<void> {
    this.closing = true;
    clearTimeout(this.topologyTimer);
    for (const disposable of this.disposables.splice(0).reverse()) {
      disposable.dispose();
    }
    for (const watcher of this.configuredPathWatchers.splice(0)) {
      watcher.dispose();
    }
    for (const uri of this.documentGenerations.keys()) {
      this.documentGenerations.set(
        uri,
        (this.documentGenerations.get(uri) ?? 0) + 1,
      );
    }
    this.documentRuntimes.clear();
    this.documentInstances.clear();
    const slots = [...this.runtimes.values()];
    const alreadyClosing = [...this.closingRuntimes.values()];
    for (const slot of slots) slot.documents.clear();
    await Promise.allSettled([
      ...alreadyClosing,
      ...slots.map(async (slot) => this.closeRuntime(slot)),
    ]);
    await this.router.closeAll();
  }
}

function runtimeFileIdentity(filePath: string): string {
  const normalized = path.normalize(filePath);
  return process.platform === 'win32' ? normalized.toLowerCase() : normalized;
}

function nearestExistingDirectory(startPath: string): string {
  let current = path.resolve(startPath);
  for (;;) {
    try {
      if (statSync(current).isDirectory()) return current;
    } catch {
      // Walk upward until a watcher can attach to a real directory.
    }
    const parent = path.dirname(current);
    if (parent === current) return current;
    current = parent;
  }
}
