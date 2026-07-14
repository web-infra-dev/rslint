import {
  workspace,
  Uri,
  Disposable,
  FileSystemWatcher,
  RelativePattern,
  WorkspaceFolder,
  OutputChannel,
  type CancellationToken,
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
import { createHash } from 'node:crypto';
import {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  ConfigModuleHost,
  loadConfigFileFresh,
  normalizeConfig,
  collectPluginMeta,
  filterConfigsByParentIgnores,
  JS_CONFIG_FILES,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigModuleEslintPluginEntry,
  type ConfigModulePluginDescriptor,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
} from '@rslint/core/config-loader';
import { PluginLintPool } from './PluginLintPool';
import type { EslintPluginLintRequest } from '@rslint/core/eslint-plugin';

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

export const JS_CONFIG_SEARCH_GLOB = `**/{${JS_CONFIG_FILES.join(',')}}`;
export const JS_CONFIG_SEARCH_EXCLUDE_PATTERN = '**/{node_modules,.git}/**';
export const CONFIG_WATCH_GLOB = `**/{${[
  ...JS_CONFIG_FILES,
  'rslint.json',
  'rslint.jsonc',
  '.gitignore',
  ...LOCKFILE_NAMES,
].join(',')}}`;
export const CONFIG_REFRESH_WATCH_GLOB = `**/{${[
  ...JS_CONFIG_FILES,
  'rslint.json',
  'rslint.jsonc',
  ...LOCKFILE_NAMES,
].join(',')}}`;

export type ConfigRefreshReason =
  | 'initial'
  | 'config-change'
  | 'gitignore-change'
  | 'dependency-change';

interface ConfigRefreshRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  reason: ConfigRefreshReason;
}

export interface ConfigActivationWireResponse {
  transactionId: string;
  /** Plugin-lint requests use the discovery transaction as their generation. */
  generation: string;
  /** Empty when no matching worker generation could be staged. */
  eslintPluginEntries: ConfigModuleEslintPluginEntry[];
  /** False lets Go preserve its last-good catalog instead of committing. */
  pluginHostReady: boolean;
}

export interface ConfigTransactionControlRequest {
  protocolVersion: typeof CONFIG_DISCOVERY_PROTOCOL_VERSION;
  transactionId: string;
}

export interface ConfigCommitWireResponse {
  transactionId: string;
  generation: string;
  committed: true;
}

export interface ConfigAbortWireResponse {
  transactionId: string;
  generation: string;
  aborted: true;
}

/** Structural seams keep the JSON-RPC transaction adapter independently testable. */
export interface ConfigModuleHostAdapter {
  loadConfigs(
    request: LoadConfigsRequest,
    signal?: AbortSignal,
  ): Promise<LoadConfigsResponse>;
  activateConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
    prepare?: (activation: ActivateConfigsResponse) => Promise<void>,
  ): Promise<ActivateConfigsResponse>;
  deleteSession(transactionId: string): boolean;
}

export interface PluginLintPoolAdapter {
  prepare(
    descriptors: ConfigModulePluginDescriptor[],
    fingerprint: string,
    generation: string,
  ): Promise<boolean>;
  commit(generation: string): Promise<boolean>;
  abort(generation: string): Promise<void>;
}

export function normalizeConfigTransactionVersion(value: unknown): 0 | 1 | 2 {
  if (typeof value !== 'number' || !Number.isInteger(value)) return 0;
  if (value >= 2) return 2;
  return value === 1 ? 1 : 0;
}

export function configRefreshReasonForPath(
  filePath: string,
): Exclude<ConfigRefreshReason, 'initial'> {
  const basename = path.basename(filePath);
  if (basename === '.gitignore') return 'gitignore-change';
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

function throwIfAborted(signal?: AbortSignal): void {
  if (!signal?.aborted) return;
  if (signal.reason instanceof Error) throw signal.reason;
  throw new Error('config transaction was cancelled');
}

function assertTransactionControlRequest(
  request: ConfigTransactionControlRequest,
): void {
  if (!request || typeof request !== 'object') {
    throw new Error('config transaction request must be an object');
  }
  if (request.protocolVersion !== CONFIG_DISCOVERY_PROTOCOL_VERSION) {
    throw new Error(
      `unsupported config transaction protocol ${String(request.protocolVersion)}`,
    );
  }
  if (
    typeof request.transactionId !== 'string' ||
    request.transactionId.length === 0
  ) {
    throw new Error('config transactionId must be a non-empty string');
  }
}

/**
 * LSP transport adapter for the shared config module host.
 *
 * Go owns discovery, ignore semantics, last-good selection and catalog commit.
 * This adapter only evaluates Go's candidates, stages the matching plugin host,
 * and mirrors Go's final commit/abort for the same transaction ID.
 */
export class LspConfigTransactionAdapter {
  private readonly transactions = new Set<string>();
  private disposed = false;

  constructor(
    private readonly host: ConfigModuleHostAdapter,
    private readonly pluginLintPool: PluginLintPoolAdapter,
    private readonly fingerprint: (
      activation: ActivateConfigsResponse,
    ) => string,
  ) {}

  async loadConfigs(
    request: LoadConfigsRequest,
    signal?: AbortSignal,
  ): Promise<LoadConfigsResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    throwIfAborted(signal);
    const transactionId = request.transactionId;
    this.transactions.add(transactionId);
    try {
      // Editor reloads must not reuse the config entry module. Go still sends
      // the shared envelope, but the LSP transport makes that entry-freshness
      // invariant explicit for every frontier. Static transitive imports retain
      // Node's normal module-cache semantics; full graph isolation requires a
      // separate evaluator realm rather than query-busting only the entry URL.
      const response = await this.host.loadConfigs(
        { ...request, loadMode: 'fresh' },
        signal,
      );
      this.assertActive();
      throwIfAborted(signal);
      return response;
    } catch (error) {
      this.cleanup(transactionId);
      throw error;
    }
  }

  async activateConfigs(
    request: ActivateConfigsRequest,
    signal?: AbortSignal,
  ): Promise<ConfigActivationWireResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    throwIfAborted(signal);
    const generation = request.transactionId;
    try {
      let pluginHostReady = false;
      const activation = await this.host.activateConfigs(
        request,
        signal,
        async (candidate) => {
          this.assertActive();
          throwIfAborted(signal);
          pluginHostReady = await this.pluginLintPool.prepare(
            candidate.pluginConfigs,
            this.fingerprint(candidate),
            generation,
          );
          this.assertActive();
          throwIfAborted(signal);
        },
      );
      this.assertActive();
      throwIfAborted(signal);
      return {
        transactionId: activation.transactionId,
        generation,
        // Never ask Go to register/dispatch placeholder rules without the
        // matching worker generation. On first startup Go may still commit the
        // ordinary native config as a degraded no-host generation; with a
        // last-good generation it instead aborts this transaction.
        eslintPluginEntries: pluginHostReady
          ? activation.eslintPluginEntries
          : [],
        pluginHostReady,
      };
    } catch (error) {
      await this.pluginLintPool.abort(generation).catch(() => undefined);
      this.cleanup(generation);
      throw error;
    }
  }

  async commitConfigs(
    request: ConfigTransactionControlRequest,
  ): Promise<ConfigCommitWireResponse> {
    this.assertActive();
    assertTransactionControlRequest(request);
    const generation = request.transactionId;
    if (!(await this.pluginLintPool.commit(generation))) {
      throw new Error(
        `failed to commit plugin-host generation ${JSON.stringify(generation)}`,
      );
    }
    this.cleanup(generation);
    return {
      transactionId: generation,
      generation,
      committed: true,
    };
  }

  async abortConfigs(
    request: ConfigTransactionControlRequest,
  ): Promise<ConfigAbortWireResponse> {
    assertTransactionControlRequest(request);
    const generation = request.transactionId;
    try {
      await this.pluginLintPool.abort(generation);
    } finally {
      this.cleanup(generation);
    }
    return {
      transactionId: generation,
      generation,
      aborted: true,
    };
  }

  dispose(): void {
    if (this.disposed) return;
    this.disposed = true;
    for (const transactionId of this.transactions) {
      this.host.deleteSession(transactionId);
    }
    this.transactions.clear();
  }

  private cleanup(transactionId: string): void {
    this.host.deleteSession(transactionId);
    this.transactions.delete(transactionId);
  }

  private assertActive(): void {
    if (this.disposed) {
      throw new Error('config transaction adapter is disposed');
    }
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

/** A loaded + normalized config file with its source path. */
interface LoadedConfig {
  configPath: string;
  hierarchyDirectory: string;
  configDirectory: string;
  entries: Record<string, unknown>[];
  sourceFingerprint: string;
  /** Marks a failed config boundary that must not lint its owned subtree. */
  unavailable?: boolean;
}

function isSameOrChildDirectory(parent: string, candidate: string): boolean {
  const relative = path.relative(parent, candidate);
  return (
    relative === '' ||
    (relative !== '..' &&
      !relative.startsWith(`..${path.sep}`) &&
      !path.isAbsolute(relative))
  );
}

export function selectUnavailableConfigBoundaryDirectories(
  usableDirectories: readonly string[],
  unavailableDirectories: readonly string[],
): string[] {
  const usable = usableDirectories.map((directory) =>
    path.normalize(directory),
  );
  const unavailable = [
    ...new Set(
      unavailableDirectories.map((directory) => path.normalize(directory)),
    ),
  ];
  const withoutUsableAncestor = unavailable.filter(
    (directory) =>
      !usable.some((ancestor) => isSameOrChildDirectory(ancestor, directory)),
  );
  return withoutUsableAncestor.filter(
    (directory, index) =>
      !withoutUsableAncestor.some(
        (ancestor, ancestorIndex) =>
          ancestorIndex !== index &&
          isSameOrChildDirectory(ancestor, directory),
      ),
  );
}

function selectEffectiveConfigFiles(configFiles: readonly Uri[]): Uri[] {
  const selectedByDirectory = new Map<string, { uri: Uri; priority: number }>();

  for (const uri of configFiles) {
    const configName = path.basename(uri.fsPath);
    const priority = JS_CONFIG_FILES.findIndex(
      (candidateName) => candidateName === configName,
    );
    if (priority < 0) continue;

    const directory = path.normalize(path.dirname(uri.fsPath));
    const selected = selectedByDirectory.get(directory);
    if (!selected || priority < selected.priority) {
      selectedByDirectory.set(directory, { uri, priority });
    }
  }

  return [...selectedByDirectory.values()]
    .map(({ uri }) => uri)
    .sort((a, b) => a.fsPath.localeCompare(b.fsPath));
}

export function filterEffectiveConfigCatalog<
  T extends { hierarchyDirectory: string; entries: unknown[] },
>(configs: T[]): T[] {
  return filterConfigsByParentIgnores(
    configs.map((config) => ({
      configDirectory: config.hierarchyDirectory,
      entries: config.entries,
      config,
    })),
  ).map(({ config }) => config);
}

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;
  private readonly lspOutputChannel: OutputChannel | undefined;
  private readonly outputChannel: OutputChannel | undefined;
  /**
   * Legacy language-client synchronization watcher. V2 replaces it with the
   * scoped direct watcher below so one workspace mutation cannot arrive both
   * as didChangeWatchedFiles and rslint/configRefresh.
   */
  private synchronizedConfigWatcher: FileSystemWatcher | undefined;
  private configWatcher: FileSystemWatcher | undefined;
  private dependencyWatcher: FileSystemWatcher | undefined;
  private configReloadTimer: ReturnType<typeof setTimeout> | undefined;
  private configReloadChain: Promise<void> = Promise.resolve();
  private lastGoodConfigs = new Map<string, LoadedConfig>();
  private hasSentJSConfig = false;
  private lifecycleEpoch = 0;
  private configGeneration = 0;
  private pluginDependencyRevision = 0;
  private pluginLintPoolDisposed = false;
  private configTransactionVersion = 0;
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
    this.lifecycleEpoch++;
    const pluginLintPool = this.pluginLintPool;
    this.hasSentJSConfig = false;
    this.lastGoodConfigs.clear();
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

    this.synchronizedConfigWatcher =
      workspace.createFileSystemWatcher(CONFIG_WATCH_GLOB);
    const clientOptions: LanguageClientOptions = {
      documentSelector: [
        { scheme: 'file', language: 'typescript' },
        { scheme: 'file', language: 'typescriptreact' },
        { scheme: 'file', language: 'javascript' },
        { scheme: 'file', language: 'javascriptreact' },
      ],
      synchronize: {
        fileEvents: this.synchronizedConfigWatcher,
      },
      outputChannel: this.outputChannel,
    };

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

      this.configTransactionVersion = await this.detectConfigTransactionVersion(
        this.client,
      );

      if (this.configTransactionVersion >= 2) {
        const adapter = new LspConfigTransactionAdapter(
          new ConfigModuleHost(),
          pluginLintPool,
          (activation) => this.computeActivationFingerprint(activation),
        );
        this.configTransactionAdapter = adapter;

        this.client.onRequest(
          'rslint/loadConfigs',
          async (params: LoadConfigsRequest, token: CancellationToken) => {
            const response = await withCancellationSignal(
              token,
              async (signal) => {
                const loaded = await adapter.loadConfigs(params, signal);
                return loaded;
              },
            );
            return response;
          },
        );
        this.client.onRequest(
          'rslint/activateConfigs',
          async (params: ActivateConfigsRequest, token: CancellationToken) => {
            const response = await withCancellationSignal(
              token,
              async (signal) => {
                const activation = await adapter.activateConfigs(
                  params,
                  signal,
                );
                return activation;
              },
            );
            return response;
          },
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
      }

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

      if (this.configTransactionVersion >= 2) {
        this.installConfigRefreshWatcher();
        // The direct watcher is now live. Stop the language-client forwarding
        // watcher before initial discovery so later workspace edits have one
        // and only one transaction owner. Legacy v0/v1 keeps that watcher for
        // JSON config and .gitignore notifications to Go.
        this.synchronizedConfigWatcher?.dispose();
        this.synchronizedConfigWatcher = undefined;
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
      } else {
        // Legacy binaries still expect Node to scan the workspace and push one
        // configUpdate payload. Keep that path byte-for-byte compatible.
        this.installLegacyConfigWatchers();
        await this.loadAndSendConfig();
      }

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
      // Keeping it out of this direct v2 watcher prevents one mutation from
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

  private installLegacyConfigWatchers(): void {
    // Install the watchers before the initial load. A file changed while the
    // first import is running then queues a second, serialized reload instead
    // of being missed between initial load and watcher registration.
    this.configWatcher = workspace.createFileSystemWatcher(
      new RelativePattern(this.workspaceFolder, JS_CONFIG_SEARCH_GLOB),
    );
    const reloadConfig = (kind: string) => (uri: Uri) => {
      this.logger.debug(`${kind} changed: ${uri.fsPath}`);
      clearTimeout(this.configReloadTimer);
      this.configReloadTimer = setTimeout(() => {
        this.configReloadTimer = undefined;
        this.loadAndSendConfig().catch((err: unknown) => {
          this.logger.error('Failed to reload JS config', err);
        });
      }, 300);
    };
    const reloadJSConfig = reloadConfig('JS config file');
    this.configWatcher.onDidChange(reloadJSConfig);
    this.configWatcher.onDidCreate(reloadJSConfig);
    this.configWatcher.onDidDelete(reloadJSConfig);

    // computeFingerprint includes workspace-root dependency lockfiles because
    // an install can replace a mounted plugin without changing its config.
    // The language client's watcher only informs Go; this watcher is what
    // rebuilds the Node plugin host.
    this.dependencyWatcher = workspace.createFileSystemWatcher(
      new RelativePattern(
        this.workspaceFolder,
        '{package-lock.json,pnpm-lock.yaml,yarn.lock}',
      ),
    );
    const reloadDependencies = reloadConfig('Dependency lockfile');
    this.dependencyWatcher.onDidChange(reloadDependencies);
    this.dependencyWatcher.onDidCreate(reloadDependencies);
    this.dependencyWatcher.onDidDelete(reloadDependencies);
  }

  private async requestConfigRefresh(
    reason: ConfigRefreshReason,
  ): Promise<void> {
    const epoch = this.lifecycleEpoch;
    const client = this.client;
    const pluginLintPool = this.pluginLintPool;
    const adapter = this.configTransactionAdapter;
    if (
      !client ||
      !adapter ||
      this.configTransactionVersion < 2 ||
      this.pluginLintPoolDisposed
    ) {
      return;
    }
    const refresh = this.configReloadChain.then(async () => {
      if (
        !this.isLifecycleCurrent(epoch, client, pluginLintPool) ||
        adapter !== this.configTransactionAdapter ||
        this.configTransactionVersion < 2
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

  private async loadAndSendConfig(): Promise<void> {
    const epoch = this.lifecycleEpoch;
    const client = this.client;
    const pluginLintPool = this.pluginLintPool;
    if (!client || this.pluginLintPoolDisposed) return;
    const reload = this.configReloadChain.then(async () => {
      await this.performLoadAndSendConfig(epoch, client, pluginLintPool);
    });
    // Keep future reloads usable after a failed import while still returning
    // this reload's rejection to the caller for logging.
    this.configReloadChain = reload.catch(() => undefined);
    await reload;
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

  private nextConfigGeneration(): string {
    this.configGeneration++;
    return String(this.configGeneration);
  }

  private async detectConfigTransactionVersion(
    client: LanguageClient,
  ): Promise<number> {
    try {
      const capabilities = await client.sendRequest<{
        transactionVersion?: number;
      }>('rslint/configCapabilities');
      return normalizeConfigTransactionVersion(capabilities.transactionVersion);
    } catch {
      // An older custom binary does not implement the capability request. Keep
      // its historical notification/single-host protocol instead of hanging on
      // a configUpdate request that never acknowledges.
      return 0;
    }
  }

  private async sendConfigUpdate(
    client: LanguageClient,
    pluginLintPool: PluginLintPool,
    generation: string,
    payload: Record<string, unknown>,
  ): Promise<void> {
    if (this.configTransactionVersion === 1) {
      const ack = await client.sendRequest<{ generation: string }>(
        'rslint/configUpdate',
        { generation, ...payload },
      );
      if (ack.generation !== generation) {
        throw new Error(
          `config update acknowledged generation ${ack.generation}, expected ${generation}`,
        );
      }
      if (!(await pluginLintPool.commit(generation))) {
        throw new Error(
          `failed to commit plugin-host generation ${generation}`,
        );
      }
      return;
    }

    // Compatibility mode for an older Go binary: it sends reverse requests
    // without generation and accepts only the historical notification. This is
    // necessarily best-effort rather than transactional: commit the host first
    // so requests after the notification use the new plugins, accepting the
    // brief pre-notification window where old config may reach the new host.
    if (!(await pluginLintPool.commit(generation))) {
      throw new Error(`failed to commit plugin-host generation ${generation}`);
    }
    await client.sendNotification('rslint/configUpdate', payload);
  }

  private async performLoadAndSendConfig(
    epoch: number,
    client: LanguageClient,
    pluginLintPool: PluginLintPool,
  ): Promise<void> {
    if (!this.isLifecycleCurrent(epoch, client, pluginLintPool)) return;
    const discoveredConfigFiles = await workspace.findFiles(
      new RelativePattern(this.workspaceFolder, JS_CONFIG_SEARCH_GLOB),
      JS_CONFIG_SEARCH_EXCLUDE_PATTERN,
    );
    if (!this.isLifecycleCurrent(epoch, client, pluginLintPool)) return;
    const configFiles = selectEffectiveConfigFiles(discoveredConfigFiles);
    this.logger.debug(
      `Found ${discoveredConfigFiles.length} JS config file(s); selected ${configFiles.length}`,
    );

    if (configFiles.length === 0) {
      const generation = this.nextConfigGeneration();
      await pluginLintPool.prepare([], this.computeFingerprint([]), generation);
      if (!this.isLifecycleCurrent(epoch, client, pluginLintPool)) {
        await pluginLintPool.abort(generation);
        return;
      }
      try {
        await this.sendConfigUpdate(client, pluginLintPool, generation, {
          configs: [],
        });
      } catch (error) {
        await pluginLintPool.abort(generation);
        throw error;
      }
      this.lastGoodConfigs.clear();
      this.hasSentJSConfig = false;
      return;
    }

    const results = await Promise.allSettled(
      configFiles.map(async (uri) => {
        const fingerprintBeforeLoad = this.computeFileFingerprint(uri.fsPath);
        const rawConfig = await loadConfigFileFresh(uri.fsPath);
        const entries = normalizeConfig(rawConfig);
        const sourceFingerprint = this.computeFileFingerprint(uri.fsPath);
        if (sourceFingerprint !== fingerprintBeforeLoad) {
          throw new Error('config changed while it was being loaded');
        }
        return {
          configPath: uri.fsPath,
          hierarchyDirectory: path.dirname(uri.fsPath),
          // URI form, kept byte-identical to the per-file `configKey` Go
          // returns (its `getConfigForURI` cwd) so the worker's per-config
          // dispatch keys match. Also the key Go uses in `s.jsConfigs`.
          configDirectory: Uri.file(path.dirname(uri.fsPath)).toString(),
          entries,
          sourceFingerprint,
        };
      }),
    );

    const loadedConfigs = new Map<string, LoadedConfig>();
    for (let i = 0; i < results.length; i++) {
      const result = results[i];
      if (result.status === 'fulfilled') {
        loadedConfigs.set(
          path.normalize(result.value.configPath),
          result.value,
        );
      } else {
        this.logger.error(
          `Failed to load JS config: ${configFiles[i].fsPath}`,
          result.reason,
        );
      }
    }

    const configs: LoadedConfig[] = [];
    const nextLastGoodConfigs = new Map<string, LoadedConfig>();
    const unavailableConfigs: Uri[] = [];
    for (const uri of configFiles) {
      const configPath = path.normalize(uri.fsPath);
      const loaded = loadedConfigs.get(configPath);
      if (loaded) {
        configs.push(loaded);
        nextLastGoodConfigs.set(configPath, loaded);
        continue;
      }

      const lastGood = this.lastGoodConfigs.get(configPath);
      if (lastGood) {
        configs.push(lastGood);
        nextLastGoodConfigs.set(configPath, lastGood);
        this.logger.error(
          `Preserving the last good JS config for ${uri.fsPath}`,
        );
        continue;
      }

      unavailableConfigs.push(uri);
    }

    // A failed candidate without a last-good value still owns its subtree
    // unless a usable JS ancestor can take over. Keep only the outermost empty
    // boundary when several unavailable candidates are nested.
    const boundaryDirectories = selectUnavailableConfigBoundaryDirectories(
      configs.map((config) => config.hierarchyDirectory),
      unavailableConfigs.map((uri) => path.dirname(uri.fsPath)),
    );
    const unavailableByDirectory = new Map(
      unavailableConfigs.map((uri) => [
        path.normalize(path.dirname(uri.fsPath)),
        uri,
      ]),
    );
    for (const hierarchyDirectory of boundaryDirectories) {
      const uri = unavailableByDirectory.get(hierarchyDirectory);
      if (!uri) continue;
      configs.push({
        configPath: uri.fsPath,
        hierarchyDirectory,
        configDirectory: Uri.file(hierarchyDirectory).toString(),
        entries: [],
        sourceFingerprint: 'unavailable',
        unavailable: true,
      });
    }

    if (boundaryDirectories.length > 0) {
      this.logger.error(
        'JS configs without a usable ancestor were replaced by empty config boundaries until they load successfully',
      );
    }
    if (unavailableConfigs.length > boundaryDirectories.length) {
      this.logger.error(
        'JS configs covered by an ancestor config or empty boundary were omitted from the catalog',
      );
    }

    configs.sort((a, b) => a.configPath.localeCompare(b.configPath));
    const effectiveConfigs = filterEffectiveConfigCatalog(configs);
    effectiveConfigs.sort((a, b) => a.configPath.localeCompare(b.configPath));

    // Collect the ESLint-plugin metadata once: the worker-pool descriptors
    // (one per config that actually mounts plugins — others stay zero-overhead)
    // and the aggregated {prefix, ruleNames} entries Go registers as
    // placeholder rules. Single pass so the two never disagree about which
    // configs carry plugins.
    const { pluginConfigs: descriptors, eslintPluginEntries } =
      collectPluginMeta(effectiveConfigs);

    // Spin up / refresh the host. Fingerprint over the config files + lockfiles
    // so an edit or dependency install rebuilds the workers.
    const generation = this.nextConfigGeneration();
    const pluginHostReady = await pluginLintPool.prepare(
      descriptors,
      this.computeFingerprint(
        descriptors.map((c) => c.configPath),
        effectiveConfigs,
      ),
      generation,
    );

    if (!this.isLifecycleCurrent(epoch, client, pluginLintPool)) {
      await pluginLintPool.abort(generation);
      return;
    }

    // Workers re-import plugin-bearing configs. Validate every freshly-loaded
    // source again after prepare so Go's normalized entries and the staged
    // workers cannot commit across a concurrent config rewrite.
    for (const loaded of loadedConfigs.values()) {
      if (
        this.computeFileFingerprint(loaded.configPath) !==
        loaded.sourceFingerprint
      ) {
        await pluginLintPool.abort(generation);
        throw new Error(
          `config changed while staging workers: ${loaded.configPath}`,
        );
      }
    }

    // A cached config can remain usable while its file is temporarily broken,
    // but a changed community-plugin topology cannot be reconstructed from the
    // normalized wire entries alone. If the transactional host rebuild fails,
    // keep the previous Go payload and previous host together. The payload is
    // retained as a unit because plugin metadata and ordinary config entries
    // share one configUpdate notification.
    if (!pluginHostReady && this.hasSentJSConfig) {
      await pluginLintPool.abort(generation);
      this.logger.error(
        'Failed to refresh the ESLint-plugin host; preserving the previous config payload',
      );
      return;
    }

    // Wire shape Go's handleConfigUpdate parses: per-config configDirectory,
    // entries, and an optional unavailable-boundary marker, plus the top-level
    // `eslintPlugins` aggregate. `configPath` is a host-local worker detail.
    try {
      await this.sendConfigUpdate(client, pluginLintPool, generation, {
        configs: effectiveConfigs.map((c) => ({
          configDirectory: c.configDirectory,
          entries: c.entries,
          unavailable: c.unavailable || undefined,
        })),
        eslintPlugins: eslintPluginEntries,
      });
    } catch (error) {
      await pluginLintPool.abort(generation);
      throw error;
    }
    this.lastGoodConfigs = nextLastGoodConfigs;
    this.hasSentJSConfig = true;
  }

  /**
   * Fingerprint the inputs that decide whether the plugin host must rebuild:
   * each loaded config snapshot's content plus each workspace-root lockfile's
   * existence, mtime, and size. A last-good config retains its successful snapshot while
   * the current file is broken, keeping the worker and Go payload on the same
   * generation. A dependency install can replace plugin code without changing
   * config, so the lockfile also feeds the key.
   */
  private computeFileFingerprint(configPath: string): string {
    try {
      const content = fs.readFileSync(configPath);
      return `${content.byteLength}:${createHash('sha256').update(content).digest('hex')}`;
    } catch {
      return 'absent';
    }
  }

  private computeMetadataFingerprint(filePath: string): string {
    try {
      const stat = fs.statSync(filePath);
      return `${stat.mtimeMs}:${stat.size}`;
    } catch {
      return 'absent';
    }
  }

  private computeActivationFingerprint(
    activation: ActivateConfigsResponse,
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
    }> = [],
  ): string {
    const parts: string[] = [];
    const sourceFingerprintByPath = new Map(
      configs.map((config) => [
        path.normalize(config.configPath),
        config.sourceFingerprint,
      ]),
    );
    for (const p of [...configPaths].sort()) {
      const sourceFingerprint =
        sourceFingerprintByPath.get(path.normalize(p)) ??
        this.computeFileFingerprint(p);
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
    this.synchronizedConfigWatcher?.dispose();
    this.synchronizedConfigWatcher = undefined;
    this.configWatcher?.dispose();
    this.configWatcher = undefined;
    this.dependencyWatcher?.dispose();
    this.dependencyWatcher = undefined;
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
      this.lastGoodConfigs.clear();
      this.hasSentJSConfig = false;
      this.logger.debug('Rslint client is not running');
    } else if (clientResult.status === 'fulfilled') {
      this.logger.info('Rslint language client stopped');
    }
    this.lastGoodConfigs.clear();
    this.hasSentJSConfig = false;
    this.configTransactionVersion = 0;
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
    this.synchronizedConfigWatcher?.dispose();
    this.synchronizedConfigWatcher = undefined;
    this.dependencyWatcher?.dispose();
    this.dependencyWatcher = undefined;
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
