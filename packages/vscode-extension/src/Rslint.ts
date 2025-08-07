import { workspace, Uri, Disposable, WorkspaceFolder } from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  State,
} from 'vscode-languageclient/node';
import { Logger } from './logger';
import type { Extension } from './Extension';
import { fileExists, PLATFORM_BIN_REQUEST } from './utils';
import { dirname } from 'node:path';
import { chmodSync } from 'node:fs';

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;

  constructor(extension: Extension, workspaceFolder: WorkspaceFolder) {
    this.extension = extension;
    this.workspaceFolder = workspaceFolder;
    this.logger = new Logger('Rslint (workspace)').useDefaultLogLevel();
  }

  public async start(): Promise<void> {
    if (this.client) {
      this.logger.warn('Rslint client is already running');
      return;
    }

    const binPath = await this.getBinaryPath();
    this.logger.info('Rslint binary path:', binPath);
    chmodSync(binPath, 0o755);

    const run: Executable = {
      command: binPath,
      args: ['--lsp'],
    };

    const serverOptions: ServerOptions = {
      run,
      debug: run,
    };

    const clientOptions: LanguageClientOptions = {
      documentSelector: [
        { scheme: 'file', language: 'typescript' },
        { scheme: 'file', language: 'typescriptreact' },
        { scheme: 'file', language: 'javascript' },
        { scheme: 'file', language: 'javascriptreact' },
      ],
      synchronize: {
        fileEvents: workspace.createFileSystemWatcher(
          '**/{rslint.{json,jsonc},package-lock.json,pnpm-lock.yaml,yarn.lock}',
        ),
      },
    };

    this.client = new LanguageClient(
      'rslint',
      'Rslint Language Server',
      serverOptions,
      clientOptions,
    );

    try {
      await this.client.start();
      this.logger.info('Rslint language client started successfully');
    } catch (err: unknown) {
      this.logger.error('Failed to start Rslint language client', err);
      throw err;
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
  }

  private async findBinaryFromUserSettings(): Promise<string | null> {
    const binPathConfig = (
      workspace.getConfiguration().get('rslint.binPath') as string
    ).trim();

    if (binPathConfig) {
      this.logger.debug(
        `Try using Rslint binary path from user settings: ${binPathConfig}`,
      );

      const exist = await fileExists(Uri.file(binPathConfig));

      if (exist) {
        this.logger.debug(
          `Using Rslint binary from user settings: ${binPathConfig}`,
        );
        return binPathConfig;
      } else {
        this.logger.warn(
          `Rslint binary path from user settings does not exist: ${binPathConfig}`,
        );
      }
    }

    this.logger.debug(
      'No Rslint binary path found in user settings, skip resolving',
    );
    return null;
  }

  private findBinaryFromNodeModules(): string | null {
    const searchRoot = this.workspaceFolder.uri.fsPath;

    try {
      this.logger.debug('Looking for Rslint binary in node_modules');
      const pathToRslintCorePackage = dirname(
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

  // Try resolve Rslint binary path in the following order:
  // 1. From workspace settings
  // 2. From `node_modules`
  // 3. From `node_modules` in PnP mode
  // 4. From extension built-in as fallback
  private async getBinaryPath(): Promise<string> {
    const finalBinPath =
      (await this.findBinaryFromUserSettings()) ??
      this.findBinaryFromNodeModules() ??
      (await this.findBinaryFromPnp()) ??
      this.findBinaryFromBuiltIn();

    this.logger.debug('Final Rslint binary path:', finalBinPath);
    return finalBinPath;
  }
}
