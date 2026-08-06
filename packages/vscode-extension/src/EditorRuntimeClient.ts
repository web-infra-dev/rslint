import type { TextDocument, Uri, OutputChannel } from 'vscode';
import { workspace } from 'vscode';
import {
  CloseAction,
  DidCloseTextDocumentNotification,
  DidOpenTextDocumentNotification,
  ErrorAction,
  LanguageClient,
  State,
  Trace,
  type ErrorHandler,
  type LanguageClientOptions,
  type ServerOptions,
} from 'vscode-languageclient/node';

import { LanguageServerProcessOwner } from './LanguageServerProcessOwner';
import type { Logger } from './logger';
import type { ResolvedRuntime } from './RuntimeResolver';
import {
  createRuntimeDocumentSelector,
  type RuntimeDocumentRouter,
  type RuntimeRoutingTarget,
} from './RuntimeDocumentRouter';

let clientSequence = 0;
const EDITOR_RUNTIME_READY_PROTOCOL_VERSION = 1;

function editorRuntimeEnvironment(
  pnpPath: string | undefined,
): NodeJS.ProcessEnv {
  const env: NodeJS.ProcessEnv = {
    ...process.env,
    ELECTRON_RUN_AS_NODE: '1',
  };
  // These names are an authenticated internal control plane, not user-facing
  // configuration. Never let stale shell values impersonate this generation.
  delete env.RSLINT_RUNTIME_IPC_ADDRESS;
  delete env.RSLINT_RUNTIME_IPC_TOKEN;
  delete env.RSLINT_RUNTIME_PNP_PATH;
  if (pnpPath) env.RSLINT_RUNTIME_PNP_PATH = pnpPath;
  return env;
}

/**
 * vscode-languageclient 9 calls `void this.stop()` from its initialization
 * failure path while the public client state is still Starting/StartFailed.
 * Its default stop implementation rejects in those states, which turns an
 * otherwise handled startup cancellation into an unhandled rejection. Closing
 * a speculative runtime while package files are changing is a normal event for
 * this extension, so make stop idempotent outside the one state in which the
 * protocol shutdown exchange is actually available.
 */
class LifecycleSafeLanguageClient extends LanguageClient {
  override stop(timeout = 2_000): Promise<void> {
    if (this.state !== State.Running) return Promise.resolve();
    return super.stop(timeout);
  }
}

export interface EditorRuntimeClientOptions {
  readonly descriptor: ResolvedRuntime;
  readonly router: RuntimeDocumentRouter;
  readonly outputChannel: OutputChannel;
  readonly traceOutputChannel: OutputChannel;
  readonly logger: Logger;
  readonly onTerminalFailure: () => void;
}

/** One LanguageClient + @rslint/core editor sidecar for a pooled install. */
export class EditorRuntimeClient implements RuntimeRoutingTarget {
  readonly runtimeKey: string;
  readonly workspaceFolder;
  private readonly processOwner: LanguageServerProcessOwner;
  private readonly client: LifecycleSafeLanguageClient;
  private readonly stateSubscription;
  private readonly runtimeReadySubscription;
  private readonly firstRuntimeReady: Promise<Error | undefined>;
  private resolveFirstRuntimeReady!: (error?: Error) => void;
  private firstRuntimeReadySettled = false;
  private generationReady = false;
  private startPromise: Promise<void> | undefined;
  private closePromise: Promise<void> | undefined;
  private closing = false;
  private terminalFailureReported = false;

  constructor(private readonly options: EditorRuntimeClientOptions) {
    const { descriptor, router, outputChannel, traceOutputChannel, logger } =
      options;
    this.runtimeKey = descriptor.key;
    this.workspaceFolder = descriptor.workspaceFolder;
    this.processOwner = new LanguageServerProcessOwner(
      process.execPath,
      [...descriptor.nodeArgs, descriptor.entryPath],
      descriptor.workingDirectory,
      editorRuntimeEnvironment(descriptor.pnpPath),
    );
    const serverOptions: ServerOptions = async () => this.processOwner.start();
    const selector = createRuntimeDocumentSelector(descriptor.workspaceFolder);
    const clientOptions: LanguageClientOptions = {
      workspaceFolder: descriptor.workspaceFolder,
      documentSelector:
        selector as unknown as LanguageClientOptions['documentSelector'],
      outputChannel,
      middleware: router.createMiddleware(this),
    };

    const trace = workspace
      .getConfiguration('rslint', descriptor.workspaceFolder.uri)
      .get<string>('trace.server', 'off');
    if (trace !== 'off') clientOptions.traceOutputChannel = traceOutputChannel;

    const errors: { current?: ErrorHandler } = {};
    clientOptions.errorHandler = {
      error: async (error, message, count) =>
        Promise.resolve(
          errors.current?.error(error, message, count) ?? {
            action: ErrorAction.Shutdown,
          },
        ),
      closed: async () => {
        if (this.closing) {
          return { action: CloseAction.DoNotRestart, handled: true };
        }
        const result = await Promise.resolve(
          errors.current?.closed() ?? {
            action: CloseAction.Restart,
            handled: true,
          },
        );
        if (
          result.action === CloseAction.DoNotRestart &&
          !this.terminalFailureReported
        ) {
          this.reportTerminalFailure();
        }
        return result;
      },
    };

    const suffix = ++clientSequence;
    this.client = new LifecycleSafeLanguageClient(
      `rslint-${suffix}`,
      `Rslint ${descriptor.version ?? 'local'} (${descriptor.workspaceFolder.name})`,
      serverOptions,
      clientOptions,
    );
    errors.current = this.client.createDefaultErrorHandler();
    this.firstRuntimeReady = new Promise<Error | undefined>((resolve) => {
      this.resolveFirstRuntimeReady = resolve;
    });
    this.runtimeReadySubscription = this.client.onNotification(
      'rslint/runtimeReady',
      async (params: unknown) => this.handleRuntimeReady(params),
    );
    this.stateSubscription = this.client.onDidChangeState((event) => {
      logger.debug(
        `Runtime ${descriptor.entryPath} state ${event.oldState} -> ${event.newState}`,
      );
      if (
        event.oldState === State.Running &&
        event.newState !== State.Running
      ) {
        this.generationReady = false;
        void router.resetServerSession(this).catch((error: unknown) => {
          logger.error(
            'Failed to reset routed documents after runtime exit',
            error,
          );
        });
      }
    });
  }

  async start(): Promise<void> {
    if (this.closing) throw new Error('editor runtime is closing');
    await (this.startPromise ??= this.startImpl());
  }

  private async startImpl(): Promise<void> {
    await this.client.start();
    const trace = workspace
      .getConfiguration('rslint', this.workspaceFolder.uri)
      .get<string>('trace.server', 'off');
    if (trace !== 'off') {
      await this.client.setTrace(
        trace === 'verbose' ? Trace.Verbose : Trace.Messages,
      );
    }
    // LanguageClient Running only covers the public LSP initialize response.
    // The core sends runtimeReady after its initial config/plugin generation
    // has committed, which is the actual safe ownership handoff point.
    const readinessError = await this.firstRuntimeReady;
    if (readinessError) throw readinessError;
  }

  isRunning(): boolean {
    return this.client.state === State.Running && this.generationReady;
  }

  private async handleRuntimeReady(params: unknown): Promise<void> {
    if (
      params === null ||
      typeof params !== 'object' ||
      (params as { protocolVersion?: unknown }).protocolVersion !==
        EDITOR_RUNTIME_READY_PROTOCOL_VERSION
    ) {
      const error = new Error(
        `unsupported @rslint/core editor readiness protocol: ${JSON.stringify(params)}`,
      );
      const initialHandshakePending = !this.firstRuntimeReadySettled;
      this.generationReady = false;
      this.settleFirstRuntimeReady(error);
      // Initial startup propagates the error through start(). After an
      // automatic restart that promise is already settled, so explicitly
      // retire the unusable generation instead of leaving a Running client
      // that can never own or reopen documents.
      if (!initialHandshakePending) this.reportTerminalFailure();
      return;
    }
    if (this.closing || this.client.state !== State.Running) return;
    this.generationReady = true;
    try {
      await this.options.router.runtimeBecameReady(this);
    } catch (error) {
      // The runtime generation itself is committed. A failed didOpen is logged
      // and can be repaired by the next idempotent document reconciliation.
      this.options.logger.error(
        'Failed to reopen routed documents after runtime commit',
        error,
      );
    }
    this.settleFirstRuntimeReady();
  }

  private settleFirstRuntimeReady(error?: Error): void {
    if (this.firstRuntimeReadySettled) return;
    this.firstRuntimeReadySettled = true;
    this.resolveFirstRuntimeReady(error);
  }

  private reportTerminalFailure(): void {
    if (this.closing || this.terminalFailureReported) return;
    this.terminalFailureReported = true;
    // Let LanguageClient finish the notification/close transition before the
    // manager disposes this client generation.
    queueMicrotask(() => {
      if (!this.closing) this.options.onTerminalFailure();
    });
  }

  async sendDocumentOpen(document: TextDocument): Promise<void> {
    const provider = this.client
      .getFeature(DidOpenTextDocumentNotification.method)
      .getProvider(document);
    if (!provider) {
      throw new Error(`didOpen provider is unavailable for ${document.uri}`);
    }
    await provider.send(document);
  }

  async sendDocumentClose(document: TextDocument): Promise<void> {
    const provider = this.client
      .getFeature(DidCloseTextDocumentNotification.method)
      .getProvider(document);
    if (!provider) {
      throw new Error(`didClose provider is unavailable for ${document.uri}`);
    }
    await provider.send(document);
  }

  clearDocumentDiagnostics(uri: Uri): void {
    this.client.diagnostics?.delete(uri);
  }

  async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private async closeImpl(): Promise<void> {
    this.closing = true;
    this.generationReady = false;
    this.settleFirstRuntimeReady(
      new Error('editor runtime closed before ready'),
    );
    this.stateSubscription.dispose();
    this.runtimeReadySubscription.dispose();
    this.processOwner.beginClose();
    const errors: unknown[] = [];

    // Let a healthy LSP generation exchange shutdown/exit before closing its
    // transport. If public initialize or private config readiness is stuck,
    // stop is either unavailable or times out; the process owner below remains
    // the bounded EOF/signal/tree-kill fallback. beginClose prevents an
    // automatic restart from creating another child during either path.
    if (this.client.state === State.Running) {
      try {
        await this.client.stop();
      } catch (error) {
        if (this.client.state === State.Running) errors.push(error);
      }
    }
    try {
      await this.processOwner.close();
    } catch (error) {
      errors.push(error);
    }
    try {
      await this.client.dispose();
    } catch (error) {
      if (this.client.state === State.Running) errors.push(error);
    }
    try {
      this.client.diagnostics?.dispose();
    } catch (error) {
      errors.push(error);
    }
    if (errors.length === 1) throw errors[0];
    if (errors.length > 1) {
      throw new AggregateError(errors, 'failed to close editor runtime');
    }
  }
}
