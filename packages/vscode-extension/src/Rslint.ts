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

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;
  private lspOutputChannel: OutputChannel | undefined;
  private outputChannel: OutputChannel | undefined;
  private configWatcher: FileSystemWatcher | undefined;

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

      // Load JS config and send to LSP server
      await this.loadAndSendConfig();

      // Watch for JS config file changes anywhere in the workspace
      this.configWatcher = workspace.createFileSystemWatcher(
        new RelativePattern(
          this.workspaceFolder,
          '**/rslint.config.{ts,mts,js,mjs}',
        ),
      );
      const reloadConfig = () => {
        this.loadAndSendConfig().catch((err: unknown) => {
          this.logger.error('Failed to reload JS config', err);
        });
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
    const configFiles = await workspace.findFiles(
      new RelativePattern(
        this.workspaceFolder,
        '**/rslint.config.{js,mjs,ts,mts}',
      ),
    );

    if (configFiles.length === 0) {
      // No JS configs found — send empty configs to clear any previous state
      if (!this.client) return;
      await this.client.sendNotification('rslint/configUpdate', {
        configs: [],
      });
      return;
    }

    const results = await Promise.allSettled(
      configFiles.map(async uri => {
        const rawConfig = await this.loadConfigFresh(uri.fsPath);
        const entries = normalizeConfig(rawConfig);
        this.logger.info(`Loaded JS config: ${uri.fsPath}`);
        return {
          configDirectory: Uri.file(path.dirname(uri.fsPath)).toString(),
          entries,
        };
      }),
    );

    const configs: { configDirectory: string; entries: unknown[] }[] = [];
    for (let i = 0; i < results.length; i++) {
      const result = results[i];
      if (result.status === 'fulfilled') {
        configs.push(result.value);
      } else {
        this.logger.error(
          `Failed to load JS config: ${configFiles[i].fsPath}`,
          result.reason,
        );
      }
    }

    if (!this.client) return;
    await this.client.sendNotification('rslint/configUpdate', { configs });
  }

  /**
   * Load a JS/TS config file with ESM cache busting for hot reload.
   * Appends ?t=timestamp to the file URL so that Node.js treats each
   * reload as a new module (bypassing ESM cache). For .ts/.mts, if
   * native import fails (e.g. Electron without strip-types), falls
   * back to loadConfigFile which uses jiti (no caching issue).
   */
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
    } catch {
      return loadConfigFile(configPath);
    }
  }

  public async stop(): Promise<void> {
    if (!this.client) {
      this.logger.debug('Rslint client is not running');
      return;
    }

    try {
      await this.client.stop();
      this.logger.info('Rslint language client stopped');
    } catch (err: unknown) {
      this.logger.error('Error stopping Rslint language client', err);
      throw err;
    } finally {
      this.client = undefined;
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
    if (this.client) {
      this.client.stop().catch((err: unknown) => {
        this.logger.error('Error disposing Rslint client', err);
      });
    }
    this.configWatcher?.dispose();
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
      const platformPackageBinPath = require.resolve(PLATFORM_BIN_REQUEST, {
        paths: [pathToRslintCorePackage],
      });

      this.logger.debug(
        `Using Rslint binary from node_modules: ${platformPackageBinPath}`,
      );
      return platformPackageBinPath;
    } catch (err) {
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
        // eslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
        const yarnPnpApi = require(yarnPnpFile.fsPath);

        const rslintCorePackage = yarnPnpApi.resolveRequest(
          '@rslint/core/package.json',
          folder.uri.fsPath,
        );

        if (!rslintCorePackage) {
          continue;
        }

        const rslintPlatformPkg = Uri.file(
          yarnPnpApi.resolveRequest(
            PLATFORM_BIN_REQUEST,
            rslintCorePackage,
          ) as string,
        );

        if (await fileExists(rslintPlatformPkg)) {
          this.logger.debug(
            `Using Rslint binary from PnP: ${rslintPlatformPkg.fsPath}`,
          );
          return rslintPlatformPkg.fsPath;
        }
      } catch (err) {
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
      .get<RslintBinPath>('rslint.binPath')!;

    let finalBinPath: string | null = null;
    if (binPathConfig === 'local') {
      // 1. Check if the binary exists in node_modules or PnP
      // 2. Fallback to built-in binary if not found
      let localBinPath =
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
