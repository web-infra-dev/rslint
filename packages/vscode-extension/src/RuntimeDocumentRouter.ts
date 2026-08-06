import {
  RelativePattern,
  workspace,
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

export interface RuntimeRoutingTarget {
  readonly runtimeKey: string;
  readonly workspaceFolder: WorkspaceFolder;
  isRunning(): boolean;
  sendDocumentOpen(document: TextDocument): Promise<void>;
  sendDocumentClose(document: TextDocument): Promise<void>;
  clearDocumentDiagnostics(uri: Uri): void;
}

interface OpenSession {
  readonly runtime: RuntimeRoutingTarget;
  readonly document: TextDocument;
}

type SyncKind = 'open' | 'close';

interface SyncPermits {
  readonly open: Set<TextDocument>;
  readonly close: Set<TextDocument>;
}

function documentKey(document: TextDocument): string {
  return document.uri.toString();
}

export function isSupportedDocument(
  document: Pick<TextDocument, 'languageId' | 'uri'>,
): boolean {
  return (
    document.uri.scheme === 'file' &&
    SUPPORTED_LANGUAGE_IDS.has(document.languageId)
  );
}

export function createRuntimeDocumentSelector(
  folder: WorkspaceFolder,
): DocumentFilter[] {
  const pattern = new RelativePattern(folder, '**/*');
  return [...SUPPORTED_LANGUAGE_IDS].map((language) => ({
    scheme: 'file',
    language,
    pattern,
  }));
}

/** Routes every open document to exactly one pooled core runtime. */
export class RuntimeDocumentRouter {
  private readonly runtimes = new Map<string, RuntimeRoutingTarget>();
  private readonly assignments = new Map<string, string>();
  private readonly openSessions = new Map<string, OpenSession>();
  private readonly epochs = new WeakMap<TextDocument, number>();
  private readonly permits = new Map<RuntimeRoutingTarget, SyncPermits>();
  private tail: Promise<void> = Promise.resolve();

  register(runtime: RuntimeRoutingTarget): void {
    const existing = this.runtimes.get(runtime.runtimeKey);
    if (existing && existing !== runtime) {
      throw new Error(
        `runtime key is already registered: ${runtime.runtimeKey}`,
      );
    }
    this.runtimes.set(runtime.runtimeKey, runtime);
  }

  async assign(
    document: TextDocument,
    runtime: RuntimeRoutingTarget | undefined,
  ): Promise<void> {
    await this.enqueue(async () => {
      const key = documentKey(document);
      const previous = this.owner(document);
      if (runtime && this.runtimes.get(runtime.runtimeKey) !== runtime) {
        throw new Error('cannot assign a document to an unregistered runtime');
      }
      const open = this.openSessions.get(key);
      if (
        previous === runtime &&
        (!open || (open.runtime === runtime && open.document === document))
      ) {
        if (runtime?.isRunning() && !this.openSessions.has(key)) {
          await this.open(runtime, document);
        }
        return;
      }

      const errors: unknown[] = [];
      if (open) {
        try {
          await this.close(open.runtime, open.document);
        } catch (error) {
          errors.push(error);
        }
        this.release(open.runtime, open.document, errors);
      }

      if (runtime) this.assignments.set(key, runtime.runtimeKey);
      else this.assignments.delete(key);
      this.bump(document);

      if (runtime?.isRunning()) {
        try {
          await this.open(runtime, document);
        } catch (error) {
          errors.push(error);
          // Readiness is not enough to commit a handoff: the transport can
          // still fail while sending didOpen. Restore the previous live
          // session so a speculative replacement cannot strand the document.
          if (previous && previous !== runtime) {
            this.assignments.set(key, previous.runtimeKey);
            this.bump(document);
            if (open && previous.isRunning()) {
              try {
                // `open.document` may be the stale object from before VS Code
                // reopened this URI. Roll ownership back to the current
                // document instance that the failed handoff was trying to
                // transfer.
                await this.open(previous, document);
              } catch (rollbackError) {
                errors.push(rollbackError);
              }
            }
          } else if (!previous) {
            this.assignments.delete(key);
            this.bump(document);
          }
        }
      }
      if (errors.length === 1) throw errors[0];
      if (errors.length > 1) {
        throw new AggregateError(errors, 'failed to transfer routed document');
      }
    });
  }

  isAssignedTo(
    document: TextDocument,
    runtime: RuntimeRoutingTarget | undefined,
  ): boolean {
    return this.owner(document) === runtime;
  }

  async runtimeBecameReady(runtime: RuntimeRoutingTarget): Promise<void> {
    await this.enqueue(async () => {
      if (this.runtimes.get(runtime.runtimeKey) !== runtime) return;
      const errors: unknown[] = [];
      for (const document of workspace.textDocuments) {
        if (this.owner(document) !== runtime) continue;
        if (this.openSessions.has(documentKey(document))) continue;
        try {
          await this.open(runtime, document);
        } catch (error) {
          // One malformed/stale document must not prevent every later assigned
          // document from being restored after a runtime generation starts.
          errors.push(error);
        }
      }
      if (errors.length > 0) {
        throw new AggregateError(
          errors,
          'failed to open one or more routed documents',
        );
      }
    });
  }

  async unregister(runtime: RuntimeRoutingTarget): Promise<void> {
    await this.enqueue(async () => {
      if (this.runtimes.get(runtime.runtimeKey) !== runtime) return;
      const errors: unknown[] = [];
      for (const session of [...this.openSessions.values()]) {
        if (session.runtime !== runtime) continue;
        try {
          await this.close(runtime, session.document);
        } catch (error) {
          errors.push(error);
        }
        this.release(runtime, session.document, errors);
      }
      this.runtimes.delete(runtime.runtimeKey);
      for (const [uri, key] of this.assignments) {
        if (key === runtime.runtimeKey) this.assignments.delete(uri);
      }
      if (errors.length > 0) {
        throw new AggregateError(errors, 'failed to unregister runtime');
      }
    });
  }

  async resetServerSession(runtime: RuntimeRoutingTarget): Promise<void> {
    await this.enqueue(() => {
      const errors: unknown[] = [];
      for (const session of [...this.openSessions.values()]) {
        if (session.runtime === runtime) {
          this.release(runtime, session.document, errors);
        }
      }
      if (errors.length > 0) {
        throw new AggregateError(errors, 'failed to reset runtime session');
      }
    });
  }

  async closeAll(): Promise<void> {
    await this.enqueue(async () => {
      const errors: unknown[] = [];
      for (const session of [...this.openSessions.values()]) {
        try {
          await this.close(session.runtime, session.document);
        } catch (error) {
          errors.push(error);
        }
        this.release(session.runtime, session.document, errors);
      }
      this.openSessions.clear();
      this.assignments.clear();
      this.runtimes.clear();
      if (errors.length > 0) {
        throw new AggregateError(errors, 'failed to close routed documents');
      }
    });
  }

  createMiddleware(runtime: RuntimeRoutingTarget): Middleware {
    return {
      didOpen: async (document, next) => {
        if (this.hasPermit('open', runtime, document)) return next(document);
        return this.enqueue(async () => {
          const key = documentKey(document);
          if (
            this.owner(document) !== runtime ||
            !runtime.isRunning() ||
            this.openSessions.has(key)
          ) {
            return;
          }
          await next(document);
          this.openSessions.set(key, { runtime, document });
          this.bump(document);
        });
      },
      didChange: async (event, next) =>
        this.enqueue(async () => {
          if (!this.isOpenOwner(runtime, event.document)) return;
          await next(event);
        }),
      didSave: async (document, next) =>
        this.enqueue(async () => {
          if (!this.isOpenOwner(runtime, document)) return;
          await next(document);
        }),
      didClose: async (document, next) => {
        if (this.hasPermit('close', runtime, document)) return next(document);
        return this.enqueue(async () => {
          if (!this.isOpenOwner(runtime, document)) return;
          const errors: unknown[] = [];
          try {
            await next(document);
          } catch (error) {
            errors.push(error);
          }
          this.release(runtime, document, errors);
          if (errors.length > 0) throw errors[0];
        });
      },
      provideCodeActions: async (
        document,
        range,
        context,
        token,
        next,
      ): Promise<(Command | CodeAction)[] | null | undefined> => {
        if (!this.isOpenOwner(runtime, document)) return undefined;
        const epoch = this.epoch(document);
        const result = await Promise.resolve(
          next(document, range, context, token),
        );
        return epoch === this.epoch(document) &&
          this.isOpenOwner(runtime, document)
          ? result
          : undefined;
      },
      handleDiagnostics: (uri, diagnostics, next) => {
        const document = workspace.textDocuments.find(
          (candidate) => candidate.uri.toString() === uri.toString(),
        );
        if (document && this.isOpenOwner(runtime, document)) {
          next(uri, diagnostics);
        }
      },
    };
  }

  private owner(document: TextDocument): RuntimeRoutingTarget | undefined {
    const key = this.assignments.get(documentKey(document));
    return key ? this.runtimes.get(key) : undefined;
  }

  private isOpenOwner(
    runtime: RuntimeRoutingTarget,
    document: TextDocument,
  ): boolean {
    const session = this.openSessions.get(documentKey(document));
    return (
      this.owner(document) === runtime &&
      session?.runtime === runtime &&
      session.document === document
    );
  }

  private async open(
    runtime: RuntimeRoutingTarget,
    document: TextDocument,
  ): Promise<void> {
    const permits = this.permitsFor(runtime);
    permits.open.add(document);
    try {
      await runtime.sendDocumentOpen(document);
      this.openSessions.set(documentKey(document), { runtime, document });
      this.bump(document);
    } finally {
      permits.open.delete(document);
      this.releasePermits(runtime, permits);
    }
  }

  private async close(
    runtime: RuntimeRoutingTarget,
    document: TextDocument,
  ): Promise<void> {
    const permits = this.permitsFor(runtime);
    permits.close.add(document);
    try {
      await runtime.sendDocumentClose(document);
    } finally {
      permits.close.delete(document);
      this.releasePermits(runtime, permits);
    }
  }

  private release(
    runtime: RuntimeRoutingTarget,
    document: TextDocument,
    errors: unknown[],
  ): void {
    try {
      runtime.clearDocumentDiagnostics(document.uri);
    } catch (error) {
      errors.push(error);
    }
    const key = documentKey(document);
    if (this.openSessions.get(key)?.runtime === runtime) {
      this.openSessions.delete(key);
    }
    this.bump(document);
  }

  private hasPermit(
    kind: SyncKind,
    runtime: RuntimeRoutingTarget,
    document: TextDocument,
  ): boolean {
    return this.permits.get(runtime)?.[kind].has(document) ?? false;
  }

  private permitsFor(runtime: RuntimeRoutingTarget): SyncPermits {
    const existing = this.permits.get(runtime);
    if (existing) return existing;
    const created: SyncPermits = { open: new Set(), close: new Set() };
    this.permits.set(runtime, created);
    return created;
  }

  private releasePermits(
    runtime: RuntimeRoutingTarget,
    permits: SyncPermits,
  ): void {
    if (permits.open.size === 0 && permits.close.size === 0) {
      this.permits.delete(runtime);
    }
  }

  private epoch(document: TextDocument): number {
    return this.epochs.get(document) ?? 0;
  }

  private bump(document: TextDocument): void {
    this.epochs.set(document, this.epoch(document) + 1);
  }

  private async enqueue<T>(operation: () => Promise<T> | T): Promise<T> {
    const run = this.tail.then(operation, operation);
    this.tail = run.then(
      () => undefined,
      () => undefined,
    );
    return run;
  }
}
