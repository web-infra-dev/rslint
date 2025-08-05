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

export class Rslint implements Disposable {
  private client: LanguageClient | undefined;
  private readonly logger: Logger;
  private readonly extension: Extension;
  private readonly workspaceFolder: WorkspaceFolder;

  constructor(extension: Extension, workspaceFolder: WorkspaceFolder) {
    this.extension = extension;
    this.workspaceFolder = workspaceFolder;
    this.logger = new Logger('Rslint');
  }

  public async start(): Promise<void> {
    if (this.client) {
      this.logger.warn('Rslint client is already running');
      return;
    }

    const binPath = this.getBinaryPath();
    this.logger.debug('Rslint binary path:', binPath);

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

  private findBianryFromNodeModules(): Promise<string> {
    const searchRoot = this.workspaceFolder.uri.fsPath;
    // const tryFindInNodeModules = async (): Promise<string | undefined> => {
    // const rslint =
  }

  private findBianryFromBuiltIn(): Promise<string> {
    const searchRoot = this;
    // const tryFindInNodeModules = async (): Promise<string | undefined> => {
    // const rslint =
  }

  private findBianryFromUserSettings(): Promise<string> {
    const binPathConfig = (
      workspace.getConfiguration().get('rslint.binPath') as string
    ).trim();
    if (binPathConfig) {
      this.logger.debug(
        `Using Rslint binary path from user settings: ${binPathConfig}`,
      );

      // check if file exist by using workspace.fs.stat

      return Promise.resolve(binPathConfig);
    }
    // const searchRoot = this;
    // const tryFindInNodeModules = async (): Promise<string | undefined> => {
    // const rslint =
  }

  // Try resolve Rslint binary path in the following order:
  // 1. From workspace settings
  // 2. From `node_modules`
  // 3. From `node_modules` in PnP mode
  // 4. From extension built-in as fallback
  private getBinaryPath(): string {
    const extUri = this.extension.context.extensionUri;
    const binPathConfig = (
      workspace.getConfiguration().get('rslint.binPath') as string
    ).trim();

    if (extUri && extUri.trim() !== '') {
      return this.extension.activate.binPath;
    }

    if (binPathConfig && binPathConfig.trim() !== '') {
      return binPathConfig;
    }

    return Uri.joinPath(this.extension.extensionUri, 'dist', 'rslint').fsPath;
  }
}
