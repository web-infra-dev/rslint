import {
  workspace,
  Uri,
  Disposable,
  FileSystemWatcher,
  RelativePattern,
  WorkspaceFolder,
  OutputChannel,
  type CancellationToken,
  type DocumentFilter,
} from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  State,
  Trace,
} from 'vscode-languageclient/node';
import { Logger } from './logger';
import type { Extension } from './Extension';
import { fileExists, getPlatformBinRequests, RslintBinPath } from './utils';
import path from 'node:path';
import fs from 'node:fs';
import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  ConfigModuleHost,
  JS_CONFIG_FILES,
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

export const CONFIG_REFRESH_WATCH_GLOB = `**/{${[
  ...JS_CONFIG_FILES,
  'rslint.json',
  'rslint.jsonc',
  ...LOCKFILE_NAMES,
].join(',')}}`;

export type ConfigRefreshReason =
  | 'initial'
  | 'config-change'
  | 'dependency-change';

interface ConfigRefreshRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  reason: ConfigRefreshReason;
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

/** Bind each language client to the workspace whose Go process owns discovery. */
export function createLanguageClientOptions(
  workspaceFolder: WorkspaceFolder,
  outputChannel: OutputChannel | undefined,
): LanguageClientOptions {
  const workspacePattern = new RelativePattern(workspaceFolder, '**/*');
  const documentSelector = [
    {
      scheme: 'file',
      language: 'typescript',
      pattern: workspacePattern,
    },
    {
      scheme: 'file',
      language: 'typescriptreact',
      pattern: workspacePattern,
    },
    {
      scheme: 'file',
      language: 'javascript',
      pattern: workspacePattern,
    },
    {
      scheme: 'file',
      language: 'javascriptreact',
      pattern: workspacePattern,
    },
  ] satisfies DocumentFilter[];
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

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;
  private readonly lspOutputChannel: OutputChannel | undefined;
  private readonly outputChannel: OutputChannel | undefined;
  private configWatcher: FileSystemWatcher | undefined;
  private configReloadTimer: ReturnType<typeof setTimeout> | undefined;
  private configReloadChain: Promise<void> = Promise.resolve();
  private serverRestartWatcher: Disposable | undefined;
  private lifecycleEpoch = 0;
  private pluginDependencyRevision = 0;
  private pluginLintPoolDisposed = false;
  private configTransactionAdapter: LspConfigTransactionAdapter | undefined;
  /**
   * Hosts the in-process WorkerPool that answers Go's reverse
   * `rslint/pluginLint` requests for rules mounted via a config's
   * object-form `plugins`. It stays uninitialized until a config actually
   * mounts plugins.
   */
  private pluginLintPool: PluginLintPool;

  constructor(
    extension: Extension,
    workspaceFolder: WorkspaceFolder,
    outputChannel: OutputChannel,
    lspOutputChannel: OutputChannel,
  ) {
    this.extension = extension;
    this.workspaceFolder = workspaceFolder;
    this.logger = new Logger('Rslint (workspace)').useDefaultLogLevel();
    this.lspOutputChannel = lspOutputChannel;
    this.outputChannel = outputChannel;
    this.pluginLintPool = new PluginLintPool(this.logger);
  }

  public async start(): Promise<void> {
    if (this.client) {
      this.logger.warn('Rslint client is already running');
      return;
    }

    if (this.pluginLintPoolDisposed) {
      this.pluginLintPool = new PluginLintPool(this.logger);
      this.pluginLintPoolDisposed = false;
    }
    this.configReloadChain = Promise.resolve();
    this.serverRestartWatcher?.dispose();
    this.serverRestartWatcher = undefined;
    this.lifecycleEpoch++;
    const pluginLintPool = this.pluginLintPool;
    this.pluginDependencyRevision = 0;
    this.configTransactionAdapter?.dispose();
    this.configTransactionAdapter = undefined;

    const binPath = await this.getBinaryPath();
    this.logger.info('Rslint binary path:', binPath);

    const run: Executable = {
      command: binPath,
      args: ['--lsp'],
    };

    const serverOptions: ServerOptions = {
      run,
      debug: run,
    };

    // Check if LSP tracing is enabled
    const traceServer = workspace
      .getConfiguration('rslint')
      .get<string>('trace.server', 'off');
    const traceEnabled = traceServer !== 'off';

    const clientOptions = createLanguageClientOptions(
      this.workspaceFolder,
      this.outputChannel,
    );

    if (traceEnabled) {
      clientOptions.traceOutputChannel = this.lspOutputChannel;
      this.logger.info(
        'LSP tracing enabled, output will be logged to "Rslint LSP trace" channel',
      );
    } else {
      this.logger.debug('LSP tracing disabled by configuration');
    }

    this.client = new LanguageClient(
      'rslint',
      'Rslint Language Server',
      serverOptions,
      clientOptions,
    );

    try {
      await this.client.start();

      const adapter = new LspConfigTransactionAdapter(
        new ConfigModuleHost(),
        pluginLintPool,
        (activation) => this.computeActivationFingerprint(activation),
      );
      this.configTransactionAdapter = adapter;

      this.client.onRequest(
        'rslint/loadConfigs',
        async (params: LoadConfigsRequest, token: CancellationToken) =>
          withCancellationSignal(token, async (signal) =>
            adapter.loadConfigs(params, signal),
          ),
      );
      this.client.onRequest(
        'rslint/activateConfigs',
        async (params: ActivateConfigsRequest, token: CancellationToken) =>
          withCancellationSignal(token, async (signal) =>
            adapter.activateConfigs(params, signal),
          ),
      );
      this.client.onRequest(
        'rslint/commitConfigs',
        async (params: ConfigTransactionControlRequest) =>
          adapter.commitConfigs(params),
      );
      this.client.onRequest(
        'rslint/abortConfigs',
        async (params: ConfigTransactionControlRequest) =>
          adapter.abortConfigs(params),
      );

      // Answer Go's reverse `rslint/pluginLint` requests: Go lints
      // natively but dispatches rules mounted via a config's object-form
      // `plugins` back to us, where the JS WorkerPool runs them. The generic
      // string-method overload of `onRequest` handles server-initiated custom
      // requests. The handler's CancellationToken — fired when Go sends
      // $/cancelRequest for a superseded keystroke / closed document — is
      // threaded through to the pool, which bridges it to an AbortSignal and
      // cancels the in-flight worker tasks.
      this.client.onRequest(
        'rslint/pluginLint',
        async (params: EslintPluginLintRequest, token: CancellationToken) =>
          pluginLintPool.lint(params, token),
      );

      if (traceEnabled) {
        const traceLevel =
          traceServer === 'verbose' ? Trace.Verbose : Trace.Messages;
        await this.client.setTrace(traceLevel);
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
      if (retried) {
        this.logger.warn(
          'Config changed during initial activation; discovery recovered on retry',
        );
      }

      // client.start() has already emitted the initial Running transition. Any
      // later Running event belongs to an automatic native-server restart and
      // needs a fresh Go catalog; request handlers and the plugin pool survive on
      // the extension side.
      this.serverRestartWatcher = this.client.onDidChangeState((event) => {
        const recovery = recoverConfigDiscoveryOnServerState(
          event.newState,
          async (reason, beforeRequest) =>
            this.requestConfigRefresh(reason, beforeRequest),
        );
        recovery?.then(
          () => {
            this.logger.info('Config discovery recovered after server restart');
          },
          (error: unknown) => {
            this.logger.error(
              'Failed to recover config discovery after server restart',
              error,
            );
          },
        );
      });

      this.logger.info('Rslint language client started successfully');
    } catch (err: unknown) {
      this.logger.error('Failed to start Rslint language client', err);
      try {
        await this.stop();
      } catch (cleanupError: unknown) {
        this.logger.error(
          'Failed to clean up a partial Rslint start',
          cleanupError,
        );
      }
      throw err;
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
        protocolVersion: CONFIG_DISCOVERY_PROTOCOL_VERSION,
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
      !this.pluginLintPoolDisposed
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

  public async stop(): Promise<void> {
    this.lifecycleEpoch++;
    clearTimeout(this.configReloadTimer);
    this.serverRestartWatcher?.dispose();
    this.serverRestartWatcher = undefined;
    this.configWatcher?.dispose();
    this.configWatcher = undefined;
    this.configTransactionAdapter?.dispose();
    this.configTransactionAdapter = undefined;
    // Do not await configReloadChain: user config evaluation can contain a
    // non-settling top-level await. The epoch above makes any late completion
    // side-effect free, while resetting the chain keeps a future start usable.
    this.configReloadChain = Promise.resolve();

    const client = this.client;
    this.client = undefined;
    const pluginLintPool = this.pluginLintPool;
    this.pluginLintPoolDisposed = true;

    const [poolResult, clientResult] = await Promise.allSettled([
      pluginLintPool.dispose(),
      client ? client.stop() : Promise.resolve(),
    ]);
    if (!client) {
      this.logger.debug('Rslint client is not running');
    } else if (clientResult.status === 'fulfilled') {
      this.logger.info('Rslint language client stopped');
    }
    this.pluginDependencyRevision = 0;
    if (poolResult.status === 'rejected') throw poolResult.reason;
    if (clientResult.status === 'rejected') throw clientResult.reason;
  }

  public isRunning(): boolean {
    return this.client?.state === State.Running;
  }

  public getClient(): LanguageClient | undefined {
    return this.client;
  }

  public onDidChangeState(listener: (event: any) => void): Disposable {
    if (!this.client) {
      throw new Error('Client is not initialized');
    }
    return this.client.onDidChangeState(listener);
  }

  public dispose(): void {
    this.lifecycleEpoch++;
    clearTimeout(this.configReloadTimer);
    this.serverRestartWatcher?.dispose();
    this.serverRestartWatcher = undefined;
    this.configReloadChain = Promise.resolve();
    this.pluginLintPoolDisposed = true;
    this.configTransactionAdapter?.dispose();
    this.configTransactionAdapter = undefined;
    void this.pluginLintPool.dispose();
    const client = this.client;
    this.client = undefined;
    if (client) {
      client.stop().catch((err: unknown) => {
        this.logger.error('Error disposing Rslint client', err);
      });
    }
    this.configWatcher?.dispose();
    this.configWatcher = undefined;
    this.lspOutputChannel?.dispose();
  }

  private async findBinaryFromUserSettings(): Promise<string | null> {
    const customBinPathConfig = workspace
      .getConfiguration()
      .get<string>('rslint.customBinPath')
      ?.trim();

    if (!customBinPathConfig) {
      this.logger.warn(
        'rslint.binPath is set to "custom" but rslint.customBinPath is not configured',
      );
      return null;
    }

    this.logger.debug(
      `Try using Rslint binary path from user settings: ${customBinPathConfig}`,
    );

    const exist = await fileExists(Uri.file(customBinPathConfig));

    if (exist) {
      this.logger.debug(
        `Using Rslint binary from user settings: ${customBinPathConfig}`,
      );
      return customBinPathConfig;
    } else {
      this.logger.error(
        `Rslint binary path from user settings does not exist: ${customBinPathConfig}`,
      );
      return null;
    }
  }

  private findBinaryFromNodeModules(): string | null {
    const searchRoot = this.workspaceFolder.uri.fsPath;

    try {
      this.logger.debug('Looking for Rslint binary in node_modules');
      const pathToRslintCorePackage = path.dirname(
        require.resolve('@rslint/core/package.json', {
          paths: [searchRoot],
        }),
      );
      // Try each platform-package candidate in order, using the first that
      // resolves (linux ships gnu/musl variants — only one is installed).
      for (const request of getPlatformBinRequests()) {
        try {
          const platformPackageBinPath = require.resolve(request, {
            paths: [pathToRslintCorePackage],
          });

          this.logger.debug(
            `Using Rslint binary from node_modules: ${platformPackageBinPath}`,
          );
          return platformPackageBinPath;
        } catch {
          // Candidate not installed; try the next one.
        }
      }
    } catch {
      this.logger.debug('No binary found in node_modules');
    }

    return null;
  }

  private async findBinaryFromPnp(): Promise<string | null> {
    const folder = this.workspaceFolder;

    for (const extension of ['cjs', 'js']) {
      const yarnPnpFile = Uri.joinPath(folder.uri, `.pnp.${extension}`);

      if (!(await fileExists(yarnPnpFile))) {
        continue;
      }

      try {
        this.logger.debug('Looking for Rslint binary in PnP mode');
        const yarnPnpApi: {
          resolveRequest: (request: string, issuer: string) => string | null;
        } = require(yarnPnpFile.fsPath); // rslint-disable-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports

        const rslintCorePackage = yarnPnpApi.resolveRequest(
          '@rslint/core/package.json',
          folder.uri.fsPath,
        );

        if (!rslintCorePackage) {
          continue;
        }

        // Try each platform-package candidate in order, using the first that
        // resolves (linux ships gnu/musl variants — only one is installed).
        // PnP's resolveRequest throws (rather than returning null) for a
        // candidate absent from the dependency map, so each lookup needs its
        // own try/catch to fall through to the next tuple.
        for (const request of getPlatformBinRequests()) {
          try {
            const rslintPlatformPkgPath = yarnPnpApi.resolveRequest(
              request,
              rslintCorePackage,
            );

            if (!rslintPlatformPkgPath) {
              continue;
            }

            const rslintPlatformPkg = Uri.file(rslintPlatformPkgPath);

            if (await fileExists(rslintPlatformPkg)) {
              this.logger.debug(
                `Using Rslint binary from PnP: ${rslintPlatformPkg.fsPath}`,
              );
              return rslintPlatformPkg.fsPath;
            }
          } catch {
            // Candidate not registered in this PnP map; try the next one.
          }
        }
      } catch {
        this.logger.debug('No binary found in PnP mode');
      }
    }

    this.logger.debug('Not using PnP mode, skip resolving');
    return null;
  }

  private findBinaryFromBuiltIn(): string {
    const builtInBinPath = Uri.joinPath(
      this.extension.context.extensionUri,
      'dist',
      'rslint',
    ).fsPath;
    this.logger.debug(
      'Using built-in Rslint binary as fallback:',
      builtInBinPath,
    );
    return builtInBinPath;
  }

  private async getBinaryPath(): Promise<string> {
    const binPathConfig = workspace
      .getConfiguration()
      .get<RslintBinPath>('rslint.binPath');

    let finalBinPath: string | null = null;
    if (binPathConfig === 'local') {
      // 1. Check if the binary exists in node_modules or PnP
      // 2. Fallback to built-in binary if not found
      const localBinPath =
        this.findBinaryFromNodeModules() ?? (await this.findBinaryFromPnp());

      if (localBinPath === null) {
        this.logger.info(
          'No local Rslint binary found, falling back to built-in binary',
        );
      }

      finalBinPath = localBinPath ?? this.findBinaryFromBuiltIn();
    } else if (binPathConfig === 'built-in') {
      finalBinPath = this.findBinaryFromBuiltIn();
    } else if (binPathConfig === 'custom') {
      finalBinPath = await this.findBinaryFromUserSettings();
      if (finalBinPath === null) {
        throw new Error(
          'Customized Rslint binary path is not set or does not exist',
        );
      }
    }

    this.logger.debug('Final Rslint binary path:', finalBinPath);
    return finalBinPath!;
  }
}
