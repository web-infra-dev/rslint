import {
  window,
  workspace,
  type Disposable,
  type ExtensionContext,
  type OutputChannel,
} from 'vscode';
import { Logger } from './logger';
import { Rslint } from './Rslint';
import { setupStatusBar } from './statusBar';
import { registerCommands } from './commands';
import { WorkspaceDocumentRouter } from './WorkspaceDocumentRouter';
import { WorkspaceRslintCoordinator } from './WorkspaceRslintCoordinator';

/** Extension-wide owner for shared UI, channels and workspace runtimes. */
export class Extension {
  private readonly logger: Logger;
  private readonly globalDisposables: Disposable[] = [];
  private outputChannel: OutputChannel | undefined;
  private lspOutputChannel: OutputChannel | undefined;
  private workspaceFolderListener: Disposable | undefined;
  private router: WorkspaceDocumentRouter | undefined;
  private coordinator: WorkspaceRslintCoordinator | undefined;
  private closePromise: Promise<void> | undefined;
  private activated = false;

  constructor(private readonly context: ExtensionContext) {
    Logger.setDefaultLogLevel(context);
    this.logger = new Logger('Rslint (extension)').useDefaultLogLevel();
  }

  public async activate(): Promise<void> {
    if (this.activated) return;
    this.activated = true;
    this.logger.info('Rslint extension activating...');

    const outputChannel = (this.outputChannel = window.createOutputChannel(
      'Rslint Language Server',
      'log',
    ));
    const lspOutputChannel = (this.lspOutputChannel =
      window.createOutputChannel('Rslint Language Server(LSP)'));

    const router = new WorkspaceDocumentRouter();
    const coordinator = new WorkspaceRslintCoordinator(
      router,
      (workspaceFolder, rootKey) =>
        new Rslint({
          rootKey,
          extensionUri: this.context.extensionUri,
          workspaceFolder,
          outputChannel,
          lspOutputChannel,
          router,
        }),
      this.logger,
    );
    this.router = router;
    this.coordinator = coordinator;

    // Install every global facility before awaiting a root. A slow or broken
    // config must not create an event-listener gap or hide commands/channels.
    this.globalDisposables.push(setupStatusBar());
    this.globalDisposables.push(
      ...registerCommands(outputChannel, lspOutputChannel),
    );
    this.workspaceFolderListener = workspace.onDidChangeWorkspaceFolders(
      (event) => {
        coordinator.handleWorkspaceFoldersChanged(
          event,
          workspace.workspaceFolders ?? [],
        );
      },
    );

    await coordinator.initialize(workspace.workspaceFolders ?? []);
    this.logger.info('Rslint extension activated successfully');
  }

  public async deactivate(): Promise<void> {
    await this.close();
  }

  public async close(): Promise<void> {
    await (this.closePromise ??= this.closeImpl());
  }

  private async closeImpl(): Promise<void> {
    this.logger.info('Rslint extension deactivating...');
    const errors: unknown[] = [];
    // Stop accepting topology changes before withdrawing any active root.
    try {
      this.workspaceFolderListener?.dispose();
    } catch (error) {
      errors.push(error);
    }
    this.workspaceFolderListener = undefined;

    try {
      await this.coordinator?.close();
    } catch (error) {
      errors.push(error);
    }

    for (const disposable of this.globalDisposables.splice(0).reverse()) {
      try {
        disposable.dispose();
      } catch (error) {
        errors.push(error);
      }
    }
    try {
      this.outputChannel?.dispose();
    } catch (error) {
      errors.push(error);
    }
    this.outputChannel = undefined;
    try {
      this.lspOutputChannel?.dispose();
    } catch (error) {
      errors.push(error);
    }
    this.lspOutputChannel = undefined;
    this.coordinator = undefined;
    this.router = undefined;

    for (const error of errors) {
      this.logger.error('Failed to close an extension resource', error);
    }
    this.logger.info('Rslint extension deactivated');
    try {
      this.logger.dispose();
    } catch (error) {
      errors.push(error);
    }
    if (errors.length > 0) {
      throw new AggregateError(errors, 'failed to deactivate Rslint extension');
    }
  }
}
