import {
  ExtensionContext,
  Disposable,
  workspace,
  WorkspaceFolder,
} from 'vscode';
import { State } from 'vscode-languageclient/node';
import { LogLevel, Logger } from './logger';
import { Rslint } from './Rslint';

export class Extension implements Disposable {
  private rslintInstances: Map<string, Rslint> = new Map();
  private disposables: Disposable[] = [];
  private logger: Logger;
  public context: ExtensionContext;

  constructor(context: ExtensionContext) {
    this.context = context;
    this.logger = new Logger('Extension');
    this.setupLogging();
  }

  public async activate(): Promise<void> {
    this.logger.info('Rslint extension activating...');

    const folders = workspace.workspaceFolders ?? [];
    for (const folder of folders) {
      const workspaceRslint = await this.createRslintInstance(
        'default',
        folder,
      );
      await workspaceRslint.start();
    }

    this.logger.info('Rslint extension activated successfully');
  }

  public async deactivate(): Promise<void> {
    this.logger.info('Rslint extension deactivating...');

    const stopPromises = Array.from(this.rslintInstances.values()).map(
      instance => instance.stop(),
    );

    try {
      await Promise.all(stopPromises);
      this.logger.info('All Rslint instances stopped');
    } catch (err: unknown) {
      this.logger.error('Error stopping some Rslint instances', err);
    }

    this.dispose();
    this.logger.info('Rslint extension deactivated');
  }

  public async createRslintInstance(
    id: string,
    workspaceFolder: WorkspaceFolder,
  ): Promise<Rslint> {
    if (this.rslintInstances.has(id)) {
      this.logger.warn(`Rslint instance with id '${id}' already exists`);
      return this.rslintInstances.get(id)!;
    }

    // TODO: single file mode
    const rslint = new Rslint(this, workspaceFolder);
    this.rslintInstances.set(id, rslint);

    this.setupStateChangeMonitoring(rslint, id);

    this.logger.debug(`Created Rslint instance with id: ${id}`);
    return rslint;
  }

  public getRslintInstance(id: string): Rslint | undefined {
    return this.rslintInstances.get(id);
  }

  public async removeRslintInstance(id: string): Promise<void> {
    const instance = this.rslintInstances.get(id);
    if (!instance) {
      this.logger.warn(`Rslint instance with id '${id}' not found`);
      return;
    }

    await instance.stop();
    instance.dispose();
    this.rslintInstances.delete(id);

    this.logger.debug(`Removed Rslint instance with id: ${id}`);
  }

  public getAllRslintInstances(): Map<string, Rslint> {
    return new Map(this.rslintInstances);
  }

  public dispose(): void {
    this.rslintInstances.forEach(instance => {
      instance.dispose();
    });
    this.rslintInstances.clear();

    this.disposables.forEach(disposable => {
      disposable.dispose();
    });
    this.disposables = [];

    this.logger.dispose();
  }

  private setupLogging(): void {
    const isDevelopment = this.context.extensionMode === 1; // Development mode
    this.logger.setLogLevel(isDevelopment ? LogLevel.DEBUG : LogLevel.INFO);
  }

  private setupStateChangeMonitoring(rslint: Rslint, instanceId: string): void {
    const stateChangeDisposable = rslint.onDidChangeState(event => {
      this.logger.debug(
        `Rslint client state changed for instance '${instanceId}':`,
        event.oldState,
        '->',
        event.newState,
      );

      if (event.newState === State.Running) {
        this.logger.info(
          `Rslint language server started for instance '${instanceId}'`,
        );
      } else if (event.newState === State.Stopped) {
        this.logger.info(
          `Rslint language server stopped for instance '${instanceId}'`,
        );
      }
    });

    this.disposables.push(stateChangeDisposable);
    this.context.subscriptions.push(stateChangeDisposable);
  }
}
