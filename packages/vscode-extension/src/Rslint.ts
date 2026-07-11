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
  loadConfigFileFresh,
  normalizeConfig,
  collectPluginMeta,
  filterConfigsByParentIgnores,
  JS_CONFIG_FILES,
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

export const JS_CONFIG_SEARCH_GLOB = '**/rslint.config.{js,mjs,cjs,ts,mts,cts}';
export const JS_CONFIG_SEARCH_EXCLUDE_PATTERN = '**/{node_modules,.git}/**';

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
  private configWatcher: FileSystemWatcher | undefined;
  private dependencyWatcher: FileSystemWatcher | undefined;
  private configReloadTimer: ReturnType<typeof setTimeout> | undefined;
  private configReloadChain: Promise<void> = Promise.resolve();
  private lastGoodConfigs = new Map<string, LoadedConfig>();
  private hasSentJSConfig = false;
  private lifecycleEpoch = 0;
  private configGeneration = 0;
  private pluginLintPoolDisposed = false;
  private configTransactionVersion = 0;
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

    const clientOptions: LanguageClientOptions = {
      documentSelector: [
        { scheme: 'file', language: 'typescript' },
        { scheme: 'file', language: 'typescriptreact' },
        { scheme: 'file', language: 'javascript' },
        { scheme: 'file', language: 'javascriptreact' },
      ],
      synchronize: {
        fileEvents: workspace.createFileSystemWatcher(
          '**/{rslint.config.{js,mjs,cjs,ts,mts,cts},rslint.{json,jsonc},package-lock.json,pnpm-lock.yaml,yarn.lock}',
        ),
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

      // Load JS config and send it to the LSP server. Reloads queued by either
      // watcher above run after this call through configReloadChain.
      await this.loadAndSendConfig();

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
      return capabilities.transactionVersion === 1 ? 1 : 0;
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

  private computeFingerprint(
    configPaths: string[],
    configs: readonly LoadedConfig[] = [],
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
    this.configWatcher?.dispose();
    this.configWatcher = undefined;
    this.dependencyWatcher?.dispose();
    this.dependencyWatcher = undefined;
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
    void this.pluginLintPool.dispose();
    const client = this.client;
    this.client = undefined;
    if (client) {
      client.stop().catch((err: unknown) => {
        this.logger.error('Error disposing Rslint client', err);
      });
    }
    this.configWatcher?.dispose();
    this.dependencyWatcher?.dispose();
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
