import {
  languages,
  workspace,
  RelativePattern,
  type CodeAction,
  type Command,
  type DocumentFilter,
  type TextDocument,
  type Uri,
  type WorkspaceFolder,
} from 'vscode';
import type { Middleware } from 'vscode-languageclient/node';

const SUPPORTED_LANGUAGE_IDS = new Set([
  'typescript',
  'typescriptreact',
  'javascript',
  'javascriptreact',
]);

export interface DocumentRoutingRuntime {
  readonly rootKey: string;
  readonly workspaceFolder: WorkspaceFolder;
  sendDocumentOpen(document: TextDocument): Promise<void>;
  sendDocumentClose(document: TextDocument): Promise<void>;
  clearDocumentDiagnostics(uri: Uri): void;
}

interface ActiveRuntime {
  readonly runtime: DocumentRoutingRuntime;
  readonly selector: DocumentFilter[];
}

interface ServerOpenDocumentSession {
  readonly runtime: DocumentRoutingRuntime;
  readonly document: TextDocument;
}

type TextSyncKind = 'open' | 'close';

function documentKey(document: TextDocument): string {
  return document.uri.toString();
}

function permitKey(kind: TextSyncKind, document: TextDocument): string {
  return `${kind}\0${documentKey(document)}`;
}

function errorList(error: unknown): unknown[] {
  if (!(error instanceof AggregateError)) return [error];
  const errors: readonly unknown[] = error.errors;
  return [...errors];
}

function throwCollectedErrors(errors: unknown[], message: string): void {
  if (errors.length === 0) return;
  if (errors.length === 1) throw errors[0];
  throw new AggregateError(errors, message);
}

export function createWorkspaceDocumentSelector(
  workspaceFolder: WorkspaceFolder,
): DocumentFilter[] {
  const workspacePattern = new RelativePattern(workspaceFolder, '**/*');
  return [...SUPPORTED_LANGUAGE_IDS].map((language) => ({
    scheme: 'file',
    language,
    pattern: workspacePattern,
  }));
}

export function isSupportedWorkspaceDocument(
  document: Pick<TextDocument, 'languageId' | 'uri'>,
): boolean {
  return (
    document.uri.scheme === 'file' &&
    SUPPORTED_LANGUAGE_IDS.has(document.languageId)
  );
}

/**
 * Owns document-to-root routing independently from root process lifecycle.
 *
 * Root start/config evaluation never runs on this queue. The queue contains
 * only bounded document notification handoffs and middleware gates, so a
 * non-settling user config cannot block unrelated workspace lifecycle work.
 */
export class WorkspaceDocumentRouter {
  private activeRoots = new Map<string, ActiveRuntime>();
  private readonly serverOpenDocuments = new Map<
    string,
    ServerOpenDocumentSession
  >();
  private readonly documentEpochs = new WeakMap<TextDocument, number>();
  private readonly transferPermits = new Map<
    DocumentRoutingRuntime,
    Set<string>
  >();
  private operationTail: Promise<void> = Promise.resolve();

  public createMiddleware(runtime: DocumentRoutingRuntime): Middleware {
    return {
      didOpen: async (document, next) => {
        if (this.hasTransferPermit('open', runtime, document)) {
          return next(document);
        }
        return this.enqueue(async () => {
          const uri = documentKey(document);
          if (
            this.ownerForDocument(document) !== runtime ||
            this.serverOpenDocuments.has(uri)
          ) {
            return;
          }
          await next(document);
          this.serverOpenDocuments.set(uri, { runtime, document });
          this.bumpDocumentEpoch(document);
        });
      },
      didChange: async (event, next) =>
        this.enqueue(async () => {
          if (!this.isServerOpenOwner(runtime, event.document)) return;
          await next(event);
        }),
      didSave: async (document, next) =>
        this.enqueue(async () => {
          if (!this.isServerOpenOwner(runtime, document)) return;
          await next(document);
        }),
      didClose: async (document, next) => {
        if (this.hasTransferPermit('close', runtime, document)) {
          return next(document);
        }
        return this.enqueue(async () => {
          const uri = documentKey(document);
          if (this.serverOpenDocuments.get(uri)?.runtime !== runtime) return;
          const errors: unknown[] = [];
          try {
            await next(document);
          } catch (error) {
            errors.push(...errorList(error));
          }
          this.releaseDocumentSession(runtime, document, errors);
          throwCollectedErrors(errors, 'failed to close routed document');
        });
      },
      provideCodeActions: async (
        document,
        range,
        context,
        token,
        next,
      ): Promise<(Command | CodeAction)[] | null | undefined> => {
        if (!this.isServerOpenOwner(runtime, document)) return undefined;
        const epoch = this.documentEpoch(document);
        const result = await Promise.resolve(
          next(document, range, context, token),
        );
        if (
          epoch !== this.documentEpoch(document) ||
          !this.isServerOpenOwner(runtime, document)
        ) {
          return undefined;
        }
        return result;
      },
      handleDiagnostics: (uri, diagnostics, next) => {
        const document = workspace.textDocuments.find(
          (candidate) => candidate.uri.toString() === uri.toString(),
        );
        if (!document || !this.isServerOpenOwner(runtime, document)) return;
        next(uri, diagnostics);
      },
    };
  }

  public async activate(runtime: DocumentRoutingRuntime): Promise<void> {
    return this.enqueue(async () => {
      const existing = this.activeRoots.get(runtime.rootKey);
      if (existing?.runtime === runtime) return;
      if (existing) {
        throw new Error(
          `workspace root ${JSON.stringify(runtime.rootKey)} is already active`,
        );
      }

      const before = this.activeRoots;
      const after = new Map(before);
      after.set(runtime.rootKey, {
        runtime,
        selector: createWorkspaceDocumentSelector(runtime.workspaceFolder),
      });
      await this.transferDocuments(before, after, true);
    });
  }

  public async deactivate(rootKey: string): Promise<void> {
    return this.enqueue(async () => {
      const removed = this.activeRoots.get(rootKey);
      if (!removed) return;
      const before = this.activeRoots;
      const after = new Map(before);
      after.delete(rootKey);
      const errors: unknown[] = [];
      try {
        await this.transferDocuments(before, after, false);
      } catch (error) {
        errors.push(...errorList(error));
      }

      // Removal is forward-only. Drain sessions that no longer appear in
      // workspace.textDocuments (for example, a close during a restart
      // listener gap) by exact runtime identity before forgetting the root.
      this.activeRoots = after;
      for (const session of [...this.serverOpenDocuments.values()]) {
        if (session.runtime !== removed.runtime) continue;
        this.releaseDocumentSession(removed.runtime, session.document, errors);
      }
      throwCollectedErrors(errors, 'failed to deactivate document owner');
    });
  }

  /**
   * Invalidates the document session owned by a native process that exited.
   * The LanguageClient re-registers didOpen after its replacement reaches
   * Running; those callbacks share this queue and therefore run after reset.
   */
  public async resetServerSession(
    runtime: DocumentRoutingRuntime,
  ): Promise<void> {
    return this.enqueue(() => {
      if (this.activeRoots.get(runtime.rootKey)?.runtime !== runtime) return;
      const errors: unknown[] = [];
      for (const session of [...this.serverOpenDocuments.values()]) {
        if (session.runtime !== runtime) continue;
        this.releaseDocumentSession(runtime, session.document, errors);
      }
      throwCollectedErrors(errors, 'failed to reset routed server session');
    });
  }

  public async closeAll(): Promise<void> {
    return this.enqueue(async () => {
      const errors: unknown[] = [];
      for (const session of [...this.serverOpenDocuments.values()]) {
        try {
          await this.sendClose(session.runtime, session.document);
        } catch (error) {
          errors.push(...errorList(error));
        }
        this.releaseDocumentSession(session.runtime, session.document, errors);
      }
      this.serverOpenDocuments.clear();
      this.activeRoots = new Map();
      throwCollectedErrors(errors, 'failed to close routed documents');
    });
  }

  public getServerOpenOwner(document: TextDocument): string | undefined {
    return this.serverOpenDocuments.get(documentKey(document))?.runtime.rootKey;
  }

  public ownerKeyForDocument(document: TextDocument): string | undefined {
    return this.ownerKeyForDocumentIn(this.activeRoots, document);
  }

  private async transferDocuments(
    before: Map<string, ActiveRuntime>,
    after: Map<string, ActiveRuntime>,
    rollbackOnFailure: boolean,
  ): Promise<void> {
    const transfers = workspace.textDocuments
      .filter(isSupportedWorkspaceDocument)
      .map((document) => ({
        document,
        oldOwnerKey: this.ownerKeyForDocumentIn(before, document),
        newOwnerKey: this.ownerKeyForDocumentIn(after, document),
      }))
      .filter(({ oldOwnerKey, newOwnerKey }) => oldOwnerKey !== newOwnerKey);

    const closedOld: typeof transfers = [];
    const openedNew: typeof transfers = [];
    const errors: unknown[] = [];

    for (const transfer of transfers) {
      if (!transfer.oldOwnerKey) continue;
      const oldOwner = before.get(transfer.oldOwnerKey);
      if (!oldOwner) continue;
      const uri = documentKey(transfer.document);
      if (this.serverOpenDocuments.get(uri)?.runtime !== oldOwner.runtime) {
        continue;
      }
      // A rejected notification promise does not prove that no bytes reached
      // the server. Treat every attempted close as needing compensation.
      closedOld.push(transfer);
      try {
        await this.sendClose(oldOwner.runtime, transfer.document);
      } catch (error) {
        errors.push(...errorList(error));
      }
      this.releaseDocumentSession(oldOwner.runtime, transfer.document, errors);
    }

    if (errors.length > 0 && rollbackOnFailure) {
      await this.restoreOldOwners(before, closedOld, errors);
      throw new AggregateError(
        errors,
        'failed to close previous document owners',
      );
    }

    this.activeRoots = after;

    for (const transfer of transfers) {
      if (!transfer.newOwnerKey) continue;
      const newOwner = after.get(transfer.newOwnerKey);
      if (!newOwner) continue;
      try {
        await this.sendOpen(newOwner.runtime, transfer.document);
        this.serverOpenDocuments.set(documentKey(transfer.document), {
          runtime: newOwner.runtime,
          document: transfer.document,
        });
        this.bumpDocumentEpoch(transfer.document);
        openedNew.push(transfer);
      } catch (error) {
        errors.push(...errorList(error));
      }
    }

    if (errors.length === 0) return;
    if (!rollbackOnFailure) {
      throw new AggregateError(
        errors,
        'failed to open fallback document owners',
      );
    }

    for (const transfer of openedNew.reverse()) {
      if (!transfer.newOwnerKey) continue;
      const newOwner = after.get(transfer.newOwnerKey);
      if (!newOwner) continue;
      try {
        await this.sendClose(newOwner.runtime, transfer.document);
      } catch (error) {
        errors.push(...errorList(error));
      }
      this.releaseDocumentSession(newOwner.runtime, transfer.document, errors);
    }
    this.activeRoots = before;
    await this.restoreOldOwners(before, closedOld, errors);
    throw new AggregateError(errors, 'failed to activate document owner');
  }

  private async restoreOldOwners(
    before: Map<string, ActiveRuntime>,
    transfers: ReadonlyArray<{
      document: TextDocument;
      oldOwnerKey: string | undefined;
    }>,
    errors: unknown[],
  ): Promise<void> {
    this.activeRoots = before;
    for (const transfer of transfers) {
      if (!transfer.oldOwnerKey) continue;
      const oldOwner = before.get(transfer.oldOwnerKey);
      if (!oldOwner) continue;
      try {
        await this.sendOpen(oldOwner.runtime, transfer.document);
        this.serverOpenDocuments.set(documentKey(transfer.document), {
          runtime: oldOwner.runtime,
          document: transfer.document,
        });
        this.bumpDocumentEpoch(transfer.document);
      } catch (error) {
        errors.push(...errorList(error));
      }
    }
  }

  private ownerKeyForDocumentIn(
    roots: ReadonlyMap<string, ActiveRuntime>,
    document: TextDocument,
  ): string | undefined {
    if (!isSupportedWorkspaceDocument(document)) return undefined;
    let best: { key: string; depth: number } | undefined;
    for (const [key, entry] of roots) {
      if (languages.match(entry.selector, document) <= 0) continue;
      const depth = entry.runtime.workspaceFolder.uri.path
        .split('/')
        .filter(Boolean).length;
      if (
        !best ||
        depth > best.depth ||
        (depth === best.depth && key.localeCompare(best.key) < 0)
      ) {
        best = { key, depth };
      }
    }
    return best?.key;
  }

  private ownerForDocument(
    document: TextDocument,
  ): DocumentRoutingRuntime | undefined {
    const ownerKey = this.ownerKeyForDocument(document);
    return ownerKey ? this.activeRoots.get(ownerKey)?.runtime : undefined;
  }

  private isServerOpenOwner(
    runtime: DocumentRoutingRuntime,
    document: TextDocument,
  ): boolean {
    const session = this.serverOpenDocuments.get(documentKey(document));
    return (
      this.ownerForDocument(document) === runtime &&
      session?.runtime === runtime &&
      session.document === document
    );
  }

  private releaseDocumentSession(
    runtime: DocumentRoutingRuntime,
    document: TextDocument,
    errors: unknown[],
  ): void {
    const uri = documentKey(document);
    try {
      runtime.clearDocumentDiagnostics(document.uri);
    } catch (error) {
      errors.push(...errorList(error));
    } finally {
      if (this.serverOpenDocuments.get(uri)?.runtime === runtime) {
        this.serverOpenDocuments.delete(uri);
      }
      this.bumpDocumentEpoch(document);
    }
  }

  private async sendOpen(
    runtime: DocumentRoutingRuntime,
    document: TextDocument,
  ): Promise<void> {
    const key = permitKey('open', document);
    const permits = this.transferPermits.get(runtime) ?? new Set<string>();
    permits.add(key);
    this.transferPermits.set(runtime, permits);
    let send: Promise<void>;
    try {
      // getProvider().send() enters middleware synchronously before returning
      // its promise, so the narrowly-scoped permit cannot leak across awaits.
      send = runtime.sendDocumentOpen(document);
    } finally {
      permits.delete(key);
      if (permits.size === 0) this.transferPermits.delete(runtime);
    }
    await send;
  }

  private async sendClose(
    runtime: DocumentRoutingRuntime,
    document: TextDocument,
  ): Promise<void> {
    const key = permitKey('close', document);
    const permits = this.transferPermits.get(runtime) ?? new Set<string>();
    permits.add(key);
    this.transferPermits.set(runtime, permits);
    let send: Promise<void>;
    try {
      send = runtime.sendDocumentClose(document);
    } finally {
      permits.delete(key);
      if (permits.size === 0) this.transferPermits.delete(runtime);
    }
    await send;
  }

  private hasTransferPermit(
    kind: TextSyncKind,
    runtime: DocumentRoutingRuntime,
    document: TextDocument,
  ): boolean {
    return (
      this.transferPermits.get(runtime)?.has(permitKey(kind, document)) ?? false
    );
  }

  private documentEpoch(document: TextDocument): number {
    return this.documentEpochs.get(document) ?? 0;
  }

  private bumpDocumentEpoch(document: TextDocument): void {
    this.documentEpochs.set(document, this.documentEpoch(document) + 1);
  }

  private async enqueue<T>(operation: () => Promise<T> | T): Promise<T> {
    const run = this.operationTail.then(operation, operation);
    this.operationTail = run.then(
      () => undefined,
      () => undefined,
    );
    return run;
  }
}
