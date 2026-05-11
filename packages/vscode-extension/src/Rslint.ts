import {
  workspace,
  Uri,
  Disposable,
  FileSystemWatcher,
  RelativePattern,
  WorkspaceFolder,
  OutputChannel,
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
import { fileExists, PLATFORM_BIN_REQUEST, RslintBinPath } from './utils';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { loadConfigFile, normalizeConfig } from '@rslint/core/config-loader';
import { applyParentIgnoresFilter } from './workspace-config';
import {
  CompatPool,
  type LintCompatBatchParams,
  type LintCompatBatchResult,
  type NormalizedConfig,
} from './CompatPool';

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;
  private readonly lspOutputChannel: OutputChannel | undefined;
  private readonly outputChannel: OutputChannel | undefined;
  private configWatcher: FileSystemWatcher | undefined;
  private configReloadTimer: ReturnType<typeof setTimeout> | undefined;
  // Owns the WorkerPool that executes ESLint-plugin rules on behalf of
  // the Go LSP server (via `rslint/lintCompatBatch` requests). Created
  // in start(); reconfigured from loadAndSendConfig; drained in stop().
  private compatPool: CompatPool | undefined;
  private compatBatchHandlerDisposable: Disposable | undefined;
  // Monotonic counter bumped at the start of every loadAndSendConfig
  // call. The current task captures the value at entry and re-checks
  // it before publishing to Go / CompatPool. If a newer call started
  // mid-flight, the stale task short-circuits without overwriting
  // newer state with older configs.
  //
  // Without this, two configWatcher events firing inside the debounce
  // window (rare but possible — e.g. saving a config triggers a
  // file-change ping then a save ping back-to-back) race against the
  // LSP server: Go could see configs from the slower call AFTER the
  // faster call's configs. The result is that newer edits silently
  // revert until the next watcher tick.
  private configGeneration = 0;

  // Per-config mtime/size cache for `loadAndSendConfig`. Workspaces
  // with N config files used to re-import EVERY config on every
  // watcher tick (each `loadConfigFresh` is a cache-busted
  // `await import(...)`, ~50 ms for a config that pulls in
  // unicorn/typescript-eslint). When a watcher fires after a single
  // config save, only that file changed — the other N-1 stayed put.
  // Comparing the on-disk mtimeMs + byteLength lets us skip the
  // unchanged ones and reuse the previous live plugin instances.
  //
  // Cache value retains the LIVE plugin objects (raw config) because
  // they are JS module-scoped singletons; the worker pool already
  // depends on identity equality of plugin instances across
  // reconfigures, so reusing them here is consistent (and is what
  // Node's own ESM cache does internally when there's no cache-bust
  // query parameter).
  //
  // Invalidation events:
  //   - mtimeMs or size changed → re-import (the typical edit)
  //   - stat throws (file deleted, permissions changed) → re-import,
  //     letting the upstream error propagate to the user
  //   - extension stop()/dispose() → cleared
  //
  // The cache is per-Rslint-instance, so multi-workspace folders
  // each have their own cache and never share live plugin instances
  // (which is important: a plugin instance imported through folder A's
  // node_modules may differ from the same package imported through
  // folder B's).
  private readonly configCache: Map<
    string,
    { mtimeMs: number; size: number; raw: unknown }
  > = new Map();

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
  }

  public async start(): Promise<void> {
    if (this.client) {
      this.logger.warn('Rslint client is already running');
      return;
    }

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
          '**/{rslint.config.{ts,mts,js,mjs},rslint.{json,jsonc},package-lock.json,pnpm-lock.yaml,yarn.lock}',
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

      if (traceEnabled) {
        const traceLevel =
          traceServer === 'verbose' ? Trace.Verbose : Trace.Messages;
        await this.client.setTrace(traceLevel);
        this.logger.info(`LSP trace level set to: ${traceServer}`);
      }

      // Build the in-extension compat WorkerPool and wire it up to the
      // Go LSP server's `rslint/lintCompatBatch` custom request.
      //
      // Lifecycle:
      //   - CompatPool itself doesn't spawn worker_threads here — it
      //     lazily spawns when the first lintBatch arrives.
      //   - The onRequest handler is the only entry point into the
      //     pool from the LSP side. It dispatches each batch to the
      //     pool's lintBatch and threads cancel through the
      //     CancellationToken.
      //   - The handler returns a plain object that vscode-jsonrpc
      //     serializes into the JSON-RPC response. Go's
      //     LintCompatBatchResult shape matches what we return here.
      this.compatPool = new CompatPool(this.logger);
      this.compatBatchHandlerDisposable = this.client.onRequest(
        'rslint/lintCompatBatch',
        async (params: LintCompatBatchParams, token) =>
          this.handleLintCompatBatch(params, token),
      );

      // Load JS config and send to LSP server. Also pushes the
      // `ConfigDescriptor[]` (configPath + configDirectory) into
      // compatPool via `extractConfigDescriptors` — workers import
      // each `configPath` directly under the configs-flow design.
      await this.loadAndSendConfig();

      // Watch for JS config file changes anywhere in the workspace
      this.configWatcher = workspace.createFileSystemWatcher(
        new RelativePattern(
          this.workspaceFolder,
          '**/rslint.config.{ts,mts,js,mjs}',
        ),
      );
      const reloadConfig = (uri: Uri) => {
        this.logger.debug(`JS config file changed: ${uri.fsPath}`);
        clearTimeout(this.configReloadTimer);
        this.configReloadTimer = setTimeout(() => {
          this.configReloadTimer = undefined;
          this.loadAndSendConfig().catch((err: unknown) => {
            this.logger.error('Failed to reload JS config', err);
          });
        }, 300);
      };
      this.configWatcher.onDidChange(reloadConfig);
      this.configWatcher.onDidCreate(reloadConfig);
      this.configWatcher.onDidDelete(reloadConfig);

      this.logger.info('Rslint language client started successfully');
    } catch (err: unknown) {
      this.logger.error('Failed to start Rslint language client', err);
      throw err;
    }
  }

  private async loadAndSendConfig(): Promise<void> {
    // Bump-on-entry generation counter. Each task remembers its gen,
    // and re-checks `this.configGeneration === myGen` before every
    // observable side-effect (pool.reconfigure / sendNotification).
    // If a newer call superseded us, abort so the newer task's writes
    // remain authoritative.
    const myGen = ++this.configGeneration;
    const isCurrent = () => this.configGeneration === myGen;

    const configFiles = await workspace.findFiles(
      new RelativePattern(
        this.workspaceFolder,
        '**/rslint.config.{js,mjs,ts,mts}',
      ),
      '**/node_modules/**',
    );
    if (!isCurrent()) return;
    this.logger.debug(`Found ${configFiles.length} JS config file(s)`);

    if (configFiles.length === 0) {
      // No JS configs found — clear state on BOTH sides:
      //   1. Go LSP server: send empty configs notification so it
      //      drops jsConfigs and any registered placeholder plugin
      //      rules.
      //   2. CompatPool: reconfigure with empty entries so any
      //      previously-spawned WorkerPool drains and the
      //      `currentConfigs` set empties. Without this, a workspace
      //      that previously had eslintPlugins configured keeps
      //      its WorkerPool + cached plugin instances alive — a
      //      resource leak and a correctness footgun (subsequent
      //      lintBatch requests from Go after some other source
      //      re-triggers them would still see the old pool).
      if (!this.client) return;
      if (!isCurrent()) return;
      // ORDER MATTERS: drain the worker pool BEFORE telling Go to drop
      // configs. sendNotification is fire-and-forget, so once Go receives
      // configUpdate({configs:[]}) it may immediately emit lintCompatBatch
      // requests against the new state — those would race a still-alive
      // pool full of stale plugins. Reconfiguring the pool first means
      // any in-flight or queued lintCompatBatch sees the drained state
      // (CompatPool returns empty results when pool is null).
      if (this.compatPool) {
        await this.compatPool.reconfigure(
          [],
          Uri.file(this.workspaceFolder.uri.fsPath).toString(),
        );
      }
      if (!isCurrent()) return;
      await this.client.sendNotification('rslint/configUpdate', {
        configs: [],
      });
      return;
    }

    const results = await Promise.allSettled(
      configFiles.map(async (uri) => {
        const rawConfig = await this.loadConfigCached(uri.fsPath);
        // configs-flow: workers import the config file directly so we
        // no longer need to resolve plugin paths on the host side.
        // `configPath` is passed through to the WorkerPool; the
        // worker's `loadPluginsFromConfigs` imports it and pulls
        // plugin instances out of the result.
        //
        // We track BOTH the filesystem path and the URI form here.
        // `filterConfigsByParentIgnores` (shared with the CLI) expects
        // filesystem paths (it calls `fs.realpathSync` + ancestor
        // string-startsWith); Go's `jsConfigs` map keys off the URI
        // form. Carrying both avoids a second round of conversion
        // after filtering.
        const configDir = path.dirname(uri.fsPath);
        const entries = normalizeConfig(rawConfig);
        return {
          configDirectoryFsPath: configDir,
          configDirectoryUri: Uri.file(configDir).toString(),
          configPath: uri.fsPath,
          entries,
        };
      }),
    );

    // Garbage-collect cache entries for configs that disappeared (file
    // deleted, glob pattern changed). Without this the cache would grow
    // monotonically over a long LSP session.
    const aliveSet = new Set(configFiles.map((u) => u.fsPath));
    for (const key of this.configCache.keys()) {
      if (!aliveSet.has(key)) this.configCache.delete(key);
    }

    if (!isCurrent()) return;

    // First collect SUCCESSFUL loads, keyed by fs path. We feed those
    // to `filterConfigsByParentIgnores` (same routine the CLI uses
    // at cli.ts:253) so a root config's global `ignores` ALSO drops
    // nested configs inside ignored directories — matches ESLint v10
    // behavior. Without this, CLI and LSP diverged: CLI silently
    // skipped the nested config, LSP loaded it and `getConfigForURI`
    // could then route a file under the ignored directory through
    // the nested config, producing diagnostics the CLI would never
    // emit. Loud "fake green / fake red" behavior split.
    type LoadedEntry = {
      configDirectoryFsPath: string;
      configDirectoryUri: string;
      configPath: string;
      entries: ReturnType<typeof normalizeConfig>;
    };
    const loadedFs: LoadedEntry[] = [];
    for (let i = 0; i < results.length; i++) {
      const result = results[i];
      if (result.status === 'fulfilled') {
        loadedFs.push(result.value);
      } else {
        this.logger.error(
          `Failed to load JS config: ${configFiles[i].fsPath}`,
          result.reason,
        );
      }
    }

    // Delegate the parent-ignores filter to a small pure helper so it
    // stays unit-testable without spinning up the VS Code runtime.
    // See `workspace-config.ts` for the projection + back-projection
    // logic. The helper internally calls the same
    // `filterConfigsByParentIgnores` the CLI uses at cli.ts:253,
    // keeping LSP and CLI in lockstep on what counts as a discovered
    // config.
    const keptEntries = applyParentIgnoresFilter(loadedFs);

    const configs: NormalizedConfig[] = keptEntries.map((e) => ({
      configDirectory: e.configDirectoryUri,
      configPath: e.configPath,
      entries: e.entries,
    }));

    const droppedCount = loadedFs.length - keptEntries.length;
    if (droppedCount > 0) {
      this.logger.debug(
        `Dropped ${droppedCount} nested config(s) covered by parent-config global ignores`,
      );
    }

    if (!this.client) return;
    if (!isCurrent()) return;
    // Push the same set to both consumers, ORDERED: pool FIRST, Go SECOND.
    //
    //   1. compatPool — reconfigure (await) so the pool is fully built
    //      with the new plugins before Go can emit any lintCompatBatch
    //      against the new state. `sendNotification` is fire-and-forget,
    //      so the moment Go receives configUpdate it may immediately
    //      issue batches; if the pool were not ready, those batches
    //      would race against a draining/half-spawned pool.
    //
    //   2. Go LSP server — gets the wire form (configDirectory + entries
    //      including eslintPlugins metadata) via the existing custom
    //      notification. Go uses this for native-rule enabling and the
    //      enforcePlugins gate; it does NOT itself execute plugins.
    //
    // (CompatPool serializes everything through `opChain` internally,
    // so even if a batch DID race, it would queue behind reconfigure —
    // but reordering here keeps the user-visible behavior sequential
    // without relying on that downstream serialization.)
    if (this.compatPool) {
      // Workspace fallback URL is used only when configs is empty; the
      // pool's reconfigure ignores it in tier-3 (clear) anyway, so
      // passing the workspace root URI is fine even without configs.
      await this.compatPool.reconfigure(
        configs,
        Uri.file(this.workspaceFolder.uri.fsPath).toString(),
      );
    }
    if (!isCurrent()) return;
    await this.client.sendNotification('rslint/configUpdate', { configs });
  }

  /**
   * Handle one `rslint/lintCompatBatch` request from the Go LSP server.
   * Delegates to compatPool.lintBatch; any thrown error surfaces as a
   * JSON-RPC error response on the wire.
   *
   * The Go side has a strict contract: response.results.length MUST
   * equal request.files.length. CompatPool's own implementation
   * upholds that — see CompatPool.lintBatch's per-task mapping.
   */
  private async handleLintCompatBatch(
    params: LintCompatBatchParams,
    token: import('vscode').CancellationToken,
  ): Promise<LintCompatBatchResult> {
    if (!this.compatPool) {
      // Shouldn't happen — the handler is only registered after
      // compatPool is constructed. Defensive: throwing here surfaces
      // as a JSON-RPC error which the Go dispatcher marks as a
      // batch-level failure rather than silently dropping diagnostics.
      throw new Error(
        'rslint/lintCompatBatch received before compat pool was constructed',
      );
    }
    return this.compatPool.lintBatch(params, token);
  }

  /**
   * Load a JS/TS config file with ESM cache busting for hot reload.
   * Appends ?t=timestamp to the file URL so that Node.js treats each
   * reload as a new module (bypassing ESM cache). For .ts/.mts, if
   * native import fails (e.g. Electron without strip-types), falls
   * back to loadConfigFile which goes through jiti — and jiti has its
   * OWN internal module cache that the URL ?t= trick does NOT bust.
   * On a hot edit of a .ts config on a Node without native TS support,
   * the fallback path returns the previously-cached module. We log a
   * warning so the user can spot stale-config behavior; a clean fix
   * requires jiti exposing a cache-clear hook (not currently available).
   */
  /**
   * Stat-first wrapper around {@link loadConfigFresh}. Returns the
   * previously-imported `rawConfig` if the file's mtime + size are
   * unchanged since the last import; otherwise falls through to a
   * real import and refreshes the cache entry.
   *
   * Stat failure is non-fatal: we delete the cache entry and let
   * `loadConfigFresh` produce its own error path, so a deleted /
   * permission-changed config surfaces the same diagnostic as before.
   *
   * Identity reuse: returning the same `raw` reference across calls is
   * what allows the worker pool's plugin-identity fingerprint to
   * remain stable — re-importing the same source code through ESM
   * with `?t=...` would produce a NEW module record + new plugin
   * instances, and the pool's `currentFingerprint` change would
   * trigger an unnecessary full reconfigure on every save (even saves
   * to unrelated files in the workspace).
   */
  private async loadConfigCached(configPath: string): Promise<unknown> {
    let stat;
    try {
      const fs = await import('node:fs/promises');
      stat = await fs.stat(configPath);
    } catch {
      // Treat as cache miss; let loadConfigFresh raise the real error.
      this.configCache.delete(configPath);
      return this.loadConfigFresh(configPath);
    }
    const cached = this.configCache.get(configPath);
    if (
      cached &&
      cached.mtimeMs === stat.mtimeMs &&
      cached.size === stat.size
    ) {
      return cached.raw;
    }
    const raw = await this.loadConfigFresh(configPath);
    this.configCache.set(configPath, {
      mtimeMs: stat.mtimeMs,
      size: stat.size,
      raw,
    });
    return raw;
  }

  private async loadConfigFresh(configPath: string): Promise<unknown> {
    const ext = path.extname(configPath);
    if (ext === '.js' || ext === '.mjs') {
      const url = `${pathToFileURL(configPath).href}?t=${Date.now()}`;
      const mod: Record<string, unknown> = await import(url);
      return mod.default ?? mod;
    }
    // .ts/.mts: try native import with cache busting first (Node >= 22.6),
    // fall back to loadConfigFile (jiti) if native TS import is not supported.
    try {
      const url = `${pathToFileURL(configPath).href}?t=${Date.now()}`;
      const mod: Record<string, unknown> = await import(url);
      return mod.default ?? mod;
    } catch (err) {
      this.logger.warn(
        `Native TS import of ${configPath} failed; falling back to jiti. ` +
          `Note: jiti does NOT honor the ?t= cache-buster, so edits to this ` +
          `config may not be picked up until the extension reloads. ` +
          `Original error: ${err instanceof Error ? err.message : String(err)}`,
      );
      return loadConfigFile(configPath);
    }
  }

  public async stop(): Promise<void> {
    if (!this.client) {
      this.logger.debug('Rslint client is not running');
      return;
    }

    // Cancel any pending configWatcher-driven reload so it can't fire
    // after stop returns. dispose() also clears this, but stop() may
    // be called separately (e.g. on extension reload without unload),
    // and a late timer kicking `loadAndSendConfig()` against a null
    // client would log spurious errors.
    if (this.configReloadTimer !== undefined) {
      clearTimeout(this.configReloadTimer);
      this.configReloadTimer = undefined;
    }

    // Drain workers BEFORE stopping the client so any in-flight
    // lintCompatBatch handler completes (or is cancelled) cleanly.
    // After client.stop() the LSP connection is gone and a still-
    // running pool would have no one to send results back to anyway.
    this.compatBatchHandlerDisposable?.dispose();
    this.compatBatchHandlerDisposable = undefined;
    if (this.compatPool) {
      try {
        await this.compatPool.dispose();
      } catch (err) {
        this.logger.error('Error disposing compat pool', err);
      }
      this.compatPool = undefined;
    }

    try {
      await this.client.stop();
      this.logger.info('Rslint language client stopped');
    } catch (err: unknown) {
      this.logger.error('Error stopping Rslint language client', err);
      throw err;
    } finally {
      this.client = undefined;
      // Drop cached config raw imports — they hold references to the
      // plugin instances, which keep the entire plugin module graph
      // alive. On restart we want a fresh load through Node's ESM
      // resolver (mtime-keyed cache picks back up from disk).
      this.configCache.clear();
    }
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
    clearTimeout(this.configReloadTimer);
    this.configCache.clear();
    this.compatBatchHandlerDisposable?.dispose();
    this.compatBatchHandlerDisposable = undefined;
    if (this.compatPool) {
      // Fire-and-forget — dispose() is sync per VS Code's Disposable
      // contract, so we can't await pool drain. The pool's shutdown
      // is fast (worker_threads terminate within a few seconds even
      // if mid-task) and the alternative (blocking VS Code shutdown
      // on lint workers) is worse UX.
      this.compatPool.dispose().catch((err: unknown) => {
        this.logger.error('Error disposing compat pool', err);
      });
      this.compatPool = undefined;
    }
    if (this.client) {
      this.client.stop().catch((err: unknown) => {
        this.logger.error('Error disposing Rslint client', err);
      });
    }
    this.configWatcher?.dispose();
    // NOTE: lspOutputChannel is OWNED by Extension (one shared channel
    // across all folder instances) — Extension.dispose() disposes it. A
    // per-instance dispose here destroyed the shared channel for every
    // other folder's trace output (#13).
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
      const platformPackageBinPath = require.resolve(PLATFORM_BIN_REQUEST, {
        paths: [pathToRslintCorePackage],
      });

      this.logger.debug(
        `Using Rslint binary from node_modules: ${platformPackageBinPath}`,
      );
      return platformPackageBinPath;
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

        const rslintPlatformPkgPath = yarnPnpApi.resolveRequest(
          PLATFORM_BIN_REQUEST,
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
