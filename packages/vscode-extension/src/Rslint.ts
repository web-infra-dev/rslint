import {
  workspace,
  Uri,
  Disposable,
  FileSystemWatcher,
  RelativePattern,
  WorkspaceFolder,
  OutputChannel,
  TextDocument,
  type CancellationToken,
} from 'vscode';
import {
  CloseAction,
  DidCloseTextDocumentNotification,
  DidOpenTextDocumentNotification,
  ErrorAction,
  LanguageClient,
  LanguageClientOptions,
  type ErrorHandler,
  type Middleware,
  ServerOptions,
  State,
  Trace,
} from 'vscode-languageclient/node';
import { Logger } from './logger';
import path from 'node:path';
import fs from 'node:fs';
import {
  type ActivateConfigsRequest,
  type ConfigModuleActivationPlan,
  type LoadConfigsRequest,
} from '@rslint/core/config-loader';
import { PluginLintPool } from './PluginLintPool';
import type { EslintPluginLintRequest } from '@rslint/core/eslint-plugin';
import {
  LspConfigTransactionAdapter,
  type ConfigTransactionControlRequest,
} from './ConfigTransactionAdapter';
import {
  createWorkspaceDocumentSelector,
  type WorkspaceDocumentRouter,
} from './WorkspaceDocumentRouter';
import { LanguageServerProcessOwner } from './LanguageServerProcessOwner';
import type { CoreInstallation } from './CoreResolver';

/**
 * Workspace-relative lockfiles whose individual metadata feeds the
 * plugin-host fingerprint. A dependency install can swap a plugin's
 * implementation without touching the config file, so the host must rebuild.
 */
const LOCKFILE_NAMES = [
  'package-lock.json',
  'pnpm-lock.yaml',
  'yarn.lock',
] as const;

export const CONFIG_REFRESH_WATCH_GLOB = `**/{rslint.config.js,rslint.config.mjs,rslint.config.ts,rslint.config.mts,rslint.json,rslint.jsonc,${LOCKFILE_NAMES.join(',')}}`;

export type ConfigRefreshReason =
  | 'initial'
  | 'config-change'
  | 'dependency-change';

interface ConfigRefreshRequest {
  protocolVersion: number;
  reason: ConfigRefreshReason;
  /** Absolute JS/TS config module path; this client omits it for automatic discovery. */
  configPath?: string;
}

export type ConfigRefreshRequester = (
  reason: ConfigRefreshReason,
  beforeRequest?: (adapter: LspConfigTransactionAdapter) => Promise<void>,
) => Promise<void>;

/**
 * Recover the extension-side transaction host when LanguageClient restarts its
 * native server. The listener using this helper is installed only after the
 * initial Running transition, so a later Running state unambiguously means the
 * replacement process needs a new initial catalog.
 */
export function recoverConfigDiscoveryOnServerState(
  newState: State,
  requestConfigRefresh: ConfigRefreshRequester,
): Promise<void> | undefined {
  if (newState !== State.Running) return undefined;
  return requestConfigRefresh('initial', async (adapter) =>
    adapter.resetForServerRestart(),
  );
}

export function shouldResetDocumentSessionOnServerState(
  oldState: State,
  newState: State,
): boolean {
  return oldState === State.Running && newState !== State.Running;
}

/** Bind each language client to the workspace whose Go process owns discovery. */
export function createLanguageClientOptions(
  workspaceFolder: WorkspaceFolder,
  outputChannel: OutputChannel | undefined,
  middleware?: Middleware,
): LanguageClientOptions {
  const documentSelector = createWorkspaceDocumentSelector(workspaceFolder);
  return {
    workspaceFolder,
    // languageclient v9 types this client-only selector as the LSP shape,
    // whose pattern is string-only. Its converter forwards the pattern to
    // VS Code's DocumentFilter, which supports RelativePattern and preserves
    // an unambiguous workspace base even when the path contains glob syntax.
    documentSelector:
      // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
      documentSelector as unknown as LanguageClientOptions['documentSelector'],
    outputChannel,
    middleware,
  };
}

export function configRefreshReasonForPath(
  filePath: string,
): Exclude<ConfigRefreshReason, 'initial'> {
  const basename = path.basename(filePath);
  if ((LOCKFILE_NAMES as readonly string[]).includes(basename)) {
    return 'dependency-change';
  }
  return 'config-change';
}

export function isConfigSourceChangeDuringTransaction(error: unknown): boolean {
  if (!isRecord(error)) return false;
  return (
    error.code === 'CONFIG_CHANGED_DURING_LOAD' ||
    (typeof error.message === 'string' &&
      error.message.includes('config changed while'))
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

export async function retryConfigRefreshOnSourceChange(
  initial: () => Promise<void>,
  retry: () => Promise<void>,
): Promise<boolean> {
  try {
    await initial();
    return false;
  } catch (error) {
    if (!isConfigSourceChangeDuringTransaction(error)) throw error;
    await retry();
    return true;
  }
}

async function withCancellationSignal<T>(
  token: CancellationToken,
  operation: (signal: AbortSignal) => Promise<T>,
): Promise<T> {
  const controller = new AbortController();
  if (token.isCancellationRequested) controller.abort();
  const subscription = token.onCancellationRequested(() => {
    controller.abort();
  });
  try {
    return await operation(controller.signal);
  } finally {
    subscription.dispose();
  }
}

function abortError(signal: AbortSignal): Error {
  if (signal.reason instanceof Error) return signal.reason;
  const error = new Error('Rslint workspace start was cancelled');
  error.name = 'AbortError';
  return error;
}

function throwIfAborted(signal: AbortSignal): void {
  if (signal.aborted) throw abortError(signal);
}

async function raceWithAbort<T>(
  operation: Promise<T>,
  signal: AbortSignal,
): Promise<T> {
  throwIfAborted(signal);
  return new Promise<T>((resolve, reject) => {
    const onAbort = () => {
      reject(abortError(signal));
    };
    signal.addEventListener('abort', onAbort, { once: true });
    operation.then(
      (value) => {
        signal.removeEventListener('abort', onAbort);
        resolve(value);
      },
      (error: unknown) => {
        signal.removeEventListener('abort', onAbort);
        reject(error);
      },
    );
  });
}

export interface LanguageClientCloseTarget {
  readonly state: State;
  readonly diagnostics: Disposable | undefined;
  dispose(): Promise<void>;
}

/**
 * vscode-languageclient calls stop() without observing its promise when an
 * initialize request fails. Its base stop rejects for non-Running states; the
 * process owner handles those states, so normalize only that inactive case and
 * preserve actionable failures from a Running shutdown.
 */
export class ManagedLanguageClient extends LanguageClient {
  public override async start(): Promise<void> {
    const outerStart = super.start();
    if (this.state !== State.Starting) return outerStart;

    // languageclient v9 keeps a shared start promise behind its public async
    // start() call. A transport close during initialize clears the private
    // reference before the outer call adopts it, leaving its rejection
    // unobserved. A second, idempotent call adopts that shared promise now.
    const sharedStart = super.start();
    void outerStart.catch(() => undefined);
    return sharedStart;
  }

  public override async stop(timeout?: number): Promise<void> {
    const stateBeforeStop = this.state;
    try {
      await super.stop(timeout);
    } catch (error) {
      if (stateBeforeStop === State.Running) throw error;
    }
  }
}

/**
 * Disposes a LanguageClient without waiting for a possibly hung initialize
 * request. Its outer LanguageServerProcessOwner blocks restarts and terminates
 * the callback-owned child after an inactive-state rejection; failures from a
 * Running client remain independently actionable.
 */
export async function disposeLanguageClient(
  client: LanguageClientCloseTarget,
): Promise<void> {
  const diagnostics = client.diagnostics;
  const reportDisposeFailure = client.state === State.Running;
  const errors: unknown[] = [];
  try {
    await client.dispose();
  } catch (error) {
    if (reportDisposeFailure) errors.push(error);
  }
  try {
    diagnostics?.dispose();
  } catch (error) {
    errors.push(error);
  }
  if (errors.length === 1) throw errors[0];
  if (errors.length > 1) {
    throw new AggregateError(errors, 'failed to dispose language client');
  }
}

export async function waitForPromiseSettlement(
  promise: Promise<unknown>,
  timeoutMs: number,
  description: string,
): Promise<void> {
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    const settled = await Promise.race([
      promise.then(
        () => true,
        () => true,
      ),
      new Promise<false>((resolve) => {
        timer = setTimeout(() => {
          resolve(false);
        }, timeoutMs);
      }),
    ]);
    if (!settled) {
      throw new Error(`${description} did not settle within ${timeoutMs}ms`);
    }
  } finally {
    clearTimeout(timer);
  }
}

interface ClientStoppedObservation {
  readonly promise: Promise<void>;
  dispose(): void;
}

function observeClientStopped(
  client: LanguageClient,
): ClientStoppedObservation {
  if (client.state === State.Stopped) {
    return { promise: Promise.resolve(), dispose: () => undefined };
  }
  let subscription: Disposable | undefined;
  const promise = new Promise<void>((resolve) => {
    subscription = client.onDidChangeState((event) => {
      if (event.newState === State.Stopped) resolve();
    });
  });
  return {
    promise,
    dispose() {
      subscription?.dispose();
      subscription = undefined;
    },
  };
}

export interface RslintOptions {
  readonly rootKey: string;
  readonly workspaceFolder: WorkspaceFolder;
  readonly installation: CoreInstallation;
  readonly outputChannel: OutputChannel;
  readonly lspOutputChannel: OutputChannel;
  readonly router: WorkspaceDocumentRouter;
}

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  public readonly rootKey: string;
  public readonly workspaceFolder: WorkspaceFolder;
  private readonly installation: CoreInstallation;
  private readonly router: WorkspaceDocumentRouter;
  private readonly lspOutputChannel: OutputChannel | undefined;
  private readonly outputChannel: OutputChannel | undefined;
  private configWatcher: FileSystemWatcher | undefined;
  private configReloadTimer: ReturnType<typeof setTimeout> | undefined;
  private configReloadChain: Promise<void> = Promise.resolve();
  private serverRestartWatcher: Disposable | undefined;
  private serverProcessOwner: LanguageServerProcessOwner | undefined;
  private stateWatcher: Disposable | undefined;
  private readonly requestHandlers: Disposable[] = [];
  private lifecycleEpoch = 0;
  private pluginDependencyRevision = 0;
  private pluginLintPoolDisposed = false;
  private configTransactionAdapter: LspConfigTransactionAdapter | undefined;
  private startPromise: Promise<void> | undefined;
  private startOperation: Promise<void> | undefined;
  private clientStartPromise: Promise<void> | undefined;
  private closePromise: Promise<void> | undefined;
  private closing = false;
  /**
   * Hosts the in-process WorkerPool that answers Go's reverse
   * `rslint/pluginLint` requests for rules mounted via a config's
   * object-form `plugins`. It stays uninitialized until a config actually
   * mounts plugins.
   */
  private readonly pluginLintPool: PluginLintPool;

  constructor(options: RslintOptions) {
    this.rootKey = options.rootKey;
    this.workspaceFolder = options.workspaceFolder;
    this.installation = options.installation;
    this.router = options.router;
    const logger = new Logger(
      `Rslint ${options.installation.version} (${options.workspaceFolder.name})`,
    ).useDefaultLogLevel();
    this.logger = logger;
    this.lspOutputChannel = options.lspOutputChannel;
    this.outputChannel = options.outputChannel;
    try {
      this.pluginLintPool = new PluginLintPool(
        logger,
        options.installation.createPluginLintHost,
      );
    } catch (error) {
      logger.dispose();
      throw error;
    }
  }

  public async start(signal: AbortSignal): Promise<void> {
    if (this.startPromise) {
      await this.startPromise;
      return;
    }
    if (this.closing || signal.aborted) {
      throw abortError(signal);
    }
    this.startOperation = this.startImpl(signal);
    // The abort facade releases the runtime-manager slot even when JavaScript
    // module evaluation itself cannot be interrupted. startImpl retains its
    // own rejection handler and epoch checks so a late completion is harmless.
    this.startPromise = raceWithAbort(this.startOperation, signal);
    void this.startOperation.catch(() => undefined);
    await this.startPromise;
  }

  private async startImpl(signal: AbortSignal): Promise<void> {
    this.configReloadChain = Promise.resolve();
    this.lifecycleEpoch++;
    const epoch = this.lifecycleEpoch;
    const pluginLintPool = this.pluginLintPool;
    this.pluginDependencyRevision = 0;

    const binPath = this.installation.binaryPath;
    this.assertStartCurrent(epoch, signal);
    this.logger.info('Rslint binary path:', binPath);

    const serverProcessOwner = new LanguageServerProcessOwner(
      binPath,
      ['--lsp'],
      this.workspaceFolder.uri.fsPath,
    );
    this.serverProcessOwner = serverProcessOwner;
    const serverOptions: ServerOptions = async () => {
      const process = await serverProcessOwner.start();
      return process;
    };

    // Check if LSP tracing is enabled
    const traceServer = workspace
      .getConfiguration('rslint', this.workspaceFolder.uri)
      .get<string>('trace.server', 'off');
    const traceEnabled = traceServer !== 'off';

    const clientOptions = createLanguageClientOptions(
      this.workspaceFolder,
      this.outputChannel,
      this.router.createMiddleware(this),
    );
    const errorHandlerHolder: { current?: ErrorHandler } = {};
    clientOptions.errorHandler = {
      error: async (error, message, count) => {
        const result = await Promise.resolve(
          errorHandlerHolder.current?.error(error, message, count) ?? {
            action: ErrorAction.Shutdown,
          },
        );
        return result;
      },
      closed: async () => {
        if (this.closing) {
          return { action: CloseAction.DoNotRestart, handled: true };
        }
        const result = await Promise.resolve(
          errorHandlerHolder.current?.closed() ?? {
            action: CloseAction.DoNotRestart,
          },
        );
        return result;
      },
    };

    if (traceEnabled) {
      clientOptions.traceOutputChannel = this.lspOutputChannel;
      this.logger.info(
        'LSP tracing enabled, output will be logged to "Rslint LSP trace" channel',
      );
    } else {
      this.logger.debug('LSP tracing disabled by configuration');
    }

    const client = new ManagedLanguageClient(
      'rslint',
      `Rslint Language Server (${this.workspaceFolder.name})`,
      serverOptions,
      clientOptions,
    );
    errorHandlerHolder.current = client.createDefaultErrorHandler();
    this.client = client;
    this.stateWatcher = client.onDidChangeState((event) => {
      this.logger.debug(
        `Language client state ${event.oldState} -> ${event.newState}`,
      );
    });

    try {
      const clientStartPromise = client.start();
      this.clientStartPromise = clientStartPromise;
      await clientStartPromise;
      this.assertStartCurrent(epoch, signal, client);

      const adapter = new LspConfigTransactionAdapter(
        this.installation.createConfigModuleHost(),
        pluginLintPool,
        (activation) => this.computeActivationFingerprint(activation),
        this.installation.protocolVersion,
      );
      this.configTransactionAdapter = adapter;

      this.requestHandlers.push(
        client.onRequest(
          'rslint/loadConfigs',
          async (params: LoadConfigsRequest, token: CancellationToken) =>
            withCancellationSignal(token, async (requestSignal) =>
              adapter.loadConfigs(params, requestSignal),
            ),
        ),
      );
      this.requestHandlers.push(
        client.onRequest(
          'rslint/activateConfigs',
          async (params: ActivateConfigsRequest, token: CancellationToken) =>
            withCancellationSignal(token, async (requestSignal) =>
              adapter.activateConfigs(params, requestSignal),
            ),
        ),
      );
      this.requestHandlers.push(
        client.onRequest(
          'rslint/commitConfigs',
          async (params: ConfigTransactionControlRequest) =>
            adapter.commitConfigs(params),
        ),
      );
      this.requestHandlers.push(
        client.onRequest(
          'rslint/abortConfigs',
          async (params: ConfigTransactionControlRequest) =>
            adapter.abortConfigs(params),
        ),
      );

      // Answer Go's reverse `rslint/pluginLint` requests: Go lints
      // natively but dispatches rules mounted via a config's object-form
      // `plugins` back to us, where the JS WorkerPool runs them. The generic
      // string-method overload of `onRequest` handles server-initiated custom
      // requests. The handler's CancellationToken — fired when Go sends
      // $/cancelRequest for a superseded keystroke / closed document — is
      // threaded through to the pool, which bridges it to an AbortSignal and
      // cancels the in-flight worker tasks.
      this.requestHandlers.push(
        client.onRequest(
          'rslint/pluginLint',
          async (params: EslintPluginLintRequest, token: CancellationToken) =>
            pluginLintPool.lint(params, token),
        ),
      );

      // client.start() has already emitted the initial Running transition. Any
      // later Running event belongs to an automatic native-server restart.
      // Reset router-side server-open state before LanguageClient replays open
      // documents, then rebuild the replacement Go process's config catalog.
      this.serverRestartWatcher = client.onDidChangeState((event) => {
        if (
          shouldResetDocumentSessionOnServerState(
            event.oldState,
            event.newState,
          )
        ) {
          this.router.resetServerSession(this).catch((error: unknown) => {
            this.logger.error(
              'Failed to reset documents after server exit',
              error,
            );
          });
        }
        const recovery = recoverConfigDiscoveryOnServerState(
          event.newState,
          async (reason, beforeRequest) => {
            await this.router.resetServerSession(this);
            await this.requestConfigRefresh(reason, beforeRequest);
          },
        );
        recovery?.then(
          () => {
            this.logger.info(
              'Documents and config discovery recovered after server restart',
            );
          },
          (error: unknown) => {
            this.logger.error('Failed to recover after server restart', error);
          },
        );
      });

      if (traceEnabled) {
        const traceLevel =
          traceServer === 'verbose' ? Trace.Verbose : Trace.Messages;
        await client.setTrace(traceLevel);
        this.assertStartCurrent(epoch, signal, client);
        this.logger.info(`LSP trace level set to: ${traceServer}`);
      }

      this.installConfigRefreshWatcher();
      // The watcher is live before initial discovery, so a mutation during a
      // slow module evaluation schedules a second serialized transaction.
      // A plugin worker is prepared between two config fingerprints. If the
      // initial source changes in that window, Go correctly aborts the
      // generation. Retry once from the now-current bytes instead of tearing
      // down the language client before the already-live watcher can recover.
      const retried = await retryConfigRefreshOnSourceChange(
        async () => {
          await this.requestConfigRefresh('initial');
        },
        async () => {
          await this.requestConfigRefresh('config-change');
        },
      );
      this.assertStartCurrent(epoch, signal, client);
      if (retried) {
        this.logger.warn(
          'Config changed during initial activation; discovery recovered on retry',
        );
      }

      this.logger.info('Rslint language client started successfully');
    } catch (err: unknown) {
      this.logger.error('Failed to start Rslint language client', err);
      throw err;
    }
  }

  private assertStartCurrent(
    epoch: number,
    signal: AbortSignal,
    client?: LanguageClient,
  ): void {
    throwIfAborted(signal);
    if (
      this.closing ||
      epoch !== this.lifecycleEpoch ||
      (client !== undefined && client !== this.client)
    ) {
      throw abortError(signal);
    }
  }

  private installConfigRefreshWatcher(): void {
    this.configWatcher = workspace.createFileSystemWatcher(
      // Go owns the config-scoped .gitignore watcher and refresh transaction.
      // Keeping it out of this direct watcher prevents one mutation from
      // starting both a didChangeWatchedFiles and a configRefresh transaction.
      new RelativePattern(this.workspaceFolder, CONFIG_REFRESH_WATCH_GLOB),
    );
    const refreshConfig = (uri: Uri) => {
      const reason = configRefreshReasonForPath(uri.fsPath);
      if (reason === 'dependency-change') {
        // The actual package contents can change while config source bytes stay
        // identical. Feed a monotonic dependency revision into the staged host
        // fingerprint so any lockfile mutation forces a worker rebuild.
        this.pluginDependencyRevision++;
      }
      this.logger.debug(`${reason}: ${uri.fsPath}`);
      clearTimeout(this.configReloadTimer);
      this.configReloadTimer = setTimeout(() => {
        this.configReloadTimer = undefined;
        this.requestConfigRefresh(reason).catch((err: unknown) => {
          this.logger.error('Failed to refresh config discovery', err);
        });
      }, 300);
    };
    this.configWatcher.onDidChange(refreshConfig);
    this.configWatcher.onDidCreate(refreshConfig);
    this.configWatcher.onDidDelete(refreshConfig);
  }

  private async requestConfigRefresh(
    reason: ConfigRefreshReason,
    beforeRequest?: (adapter: LspConfigTransactionAdapter) => Promise<void>,
  ): Promise<void> {
    const epoch = this.lifecycleEpoch;
    const client = this.client;
    const pluginLintPool = this.pluginLintPool;
    const adapter = this.configTransactionAdapter;
    if (!client || !adapter || this.pluginLintPoolDisposed) {
      return;
    }
    const refresh = this.configReloadChain.then(async () => {
      if (
        !this.isLifecycleCurrent(epoch, client, pluginLintPool) ||
        adapter !== this.configTransactionAdapter
      ) {
        return;
      }
      await beforeRequest?.(adapter);
      if (
        !this.isLifecycleCurrent(epoch, client, pluginLintPool) ||
        adapter !== this.configTransactionAdapter
      ) {
        return;
      }
      const request: ConfigRefreshRequest = {
        protocolVersion: this.installation.protocolVersion,
        reason,
      };
      await client.sendRequest('rslint/configRefresh', request);
    });
    this.configReloadChain = refresh.catch(() => undefined);
    await refresh;
  }

  private isLifecycleCurrent(
    epoch: number,
    client: LanguageClient,
    pluginLintPool: PluginLintPool,
  ): boolean {
    return (
      epoch === this.lifecycleEpoch &&
      client === this.client &&
      pluginLintPool === this.pluginLintPool &&
      !this.pluginLintPoolDisposed &&
      !this.closing
    );
  }

  /**
   * Fingerprint the inputs that decide whether the plugin host must rebuild:
   * Go's selected config snapshots plus each workspace-root lockfile's
   * existence, mtime, and size. A dependency install can replace plugin code
   * without changing config, so the lockfile also feeds the key.
   */
  private computeMetadataFingerprint(filePath: string): string {
    try {
      const stat = fs.statSync(filePath);
      return `${stat.mtimeMs}:${stat.size}`;
    } catch {
      return 'absent';
    }
  }

  private computeActivationFingerprint(
    activation: ConfigModuleActivationPlan,
  ): string {
    const sourceFingerprint = this.computeFingerprint(
      activation.pluginConfigs.map((config) => config.configPath),
      activation.configs,
    );
    return `${sourceFingerprint}|dependency-revision:${this.pluginDependencyRevision}`;
  }

  private computeFingerprint(
    configPaths: string[],
    configs: ReadonlyArray<{
      configPath: string;
      sourceFingerprint: string;
    }>,
  ): string {
    const parts: string[] = [];
    const sourceFingerprintByPath = new Map(
      configs.map((config) => [
        path.normalize(config.configPath),
        config.sourceFingerprint,
      ]),
    );
    for (const p of [...configPaths].sort()) {
      const sourceFingerprint = sourceFingerprintByPath.get(path.normalize(p));
      if (sourceFingerprint === undefined) {
        throw new Error(`missing source fingerprint for plugin config ${p}`);
      }
      parts.push(`${p}:${sourceFingerprint}`);
    }
    for (const name of LOCKFILE_NAMES) {
      const lockPath = path.join(this.workspaceFolder.uri.fsPath, name);
      parts.push(`lock:${name}:${this.computeMetadataFingerprint(lockPath)}`);
    }
    return parts.join('|');
  }

  public async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private async closeImpl(): Promise<void> {
    const errors: unknown[] = [];
    const disposeSafely = (resource: Disposable | undefined): void => {
      if (!resource) return;
      try {
        resource.dispose();
      } catch (error) {
        errors.push(error);
      }
    };

    this.closing = true;
    this.lifecycleEpoch++;
    clearTimeout(this.configReloadTimer);
    this.configReloadTimer = undefined;
    disposeSafely(this.serverRestartWatcher);
    this.serverRestartWatcher = undefined;
    disposeSafely(this.configWatcher);
    this.configWatcher = undefined;
    disposeSafely(this.configTransactionAdapter);
    this.configTransactionAdapter = undefined;
    for (const handler of this.requestHandlers.splice(0)) {
      disposeSafely(handler);
    }
    disposeSafely(this.stateWatcher);
    this.stateWatcher = undefined;
    // Do not await startOperation/configReloadChain: user module evaluation can
    // contain a non-settling top-level await. Epoch/closing checks fence every
    // late continuation from publishing resources or state.
    this.configReloadChain = Promise.resolve();
    this.pluginLintPoolDisposed = true;

    const client = this.client;
    this.client = undefined;
    const clientStartPromise = this.clientStartPromise;
    this.clientStartPromise = undefined;
    const clientStopped =
      client?.state === State.Starting
        ? observeClientStopped(client)
        : undefined;
    const serverProcessOwner = this.serverProcessOwner;
    this.serverProcessOwner = undefined;
    // Block vscode-languageclient's automatic restart callback before its
    // graceful client shutdown begins. The owner force-terminates and awaits
    // any surviving child after the bounded protocol shutdown finishes.
    serverProcessOwner?.beginClose();
    const asynchronousCleanups: Promise<void>[] = [
      (async () => {
        await this.pluginLintPool.dispose();
      })(),
    ];
    if (client) {
      asynchronousCleanups.push(
        (async () => {
          const clientErrors: unknown[] = [];
          try {
            await disposeLanguageClient(client);
          } catch (error) {
            clientErrors.push(error);
          }
          try {
            await serverProcessOwner?.close();
          } catch (error) {
            clientErrors.push(error);
          }
          if (clientStartPromise) {
            try {
              await waitForPromiseSettlement(
                clientStartPromise,
                2_000,
                'language client start',
              );
            } catch (error) {
              clientErrors.push(error);
            }
          }
          if (clientStopped) {
            try {
              await waitForPromiseSettlement(
                clientStopped.promise,
                2_000,
                'language client terminal state',
              );
            } catch (error) {
              clientErrors.push(error);
            } finally {
              clientStopped.dispose();
            }
          }
          if (clientErrors.length > 0) {
            throw new AggregateError(
              clientErrors,
              'failed to close language client resources',
            );
          }
        })(),
      );
    } else if (serverProcessOwner) {
      asynchronousCleanups.push(
        (async () => {
          await serverProcessOwner.close();
        })(),
      );
    }
    const results = await Promise.allSettled(asynchronousCleanups);
    this.pluginDependencyRevision = 0;

    for (const result of results) {
      if (result.status === 'rejected') {
        const reason: unknown = result.reason;
        errors.push(reason);
      }
    }
    try {
      for (const error of errors) {
        this.logger.error('Failed to close Rslint workspace resource', error);
      }
      if (errors.length === 0) {
        this.logger.info('Rslint language client closed');
      }
    } catch (error) {
      errors.push(error);
    }
    try {
      this.logger.dispose();
    } catch (error) {
      errors.push(error);
    }
    if (errors.length > 0) {
      throw new AggregateError(errors, 'failed to close Rslint workspace');
    }
  }

  public isRunning(): boolean {
    return this.client?.state === State.Running;
  }

  public async sendDocumentOpen(document: TextDocument): Promise<void> {
    const provider = this.client
      ?.getFeature(DidOpenTextDocumentNotification.method)
      .getProvider(document);
    if (!provider) {
      throw new Error(`didOpen provider is unavailable for ${document.uri}`);
    }
    await provider.send(document);
  }

  public async sendDocumentClose(document: TextDocument): Promise<void> {
    const provider = this.client
      ?.getFeature(DidCloseTextDocumentNotification.method)
      .getProvider(document);
    if (!provider) {
      throw new Error(`didClose provider is unavailable for ${document.uri}`);
    }
    await provider.send(document);
  }

  public clearDocumentDiagnostics(uri: Uri): void {
    this.client?.diagnostics?.delete(uri);
  }

  public dispose(): void {
    void this.close().catch(() => undefined);
  }
}
