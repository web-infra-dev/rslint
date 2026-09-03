import { window, workspace, type Disposable, type OutputChannel } from 'vscode';
import { Logger } from './logger';
import { Rslint } from './Rslint';
import { setupStatusBar } from './statusBar';
import { registerCommands } from './commands';
import { WorkspaceDocumentRouter } from './WorkspaceDocumentRouter';
import { CoreResolver } from './CoreResolver';
import { RuntimeManager } from './RuntimeManager';

const CORE_TOPOLOGY_GLOB =
  '**/{package-lock.json,pnpm-lock.yaml,yarn.lock,node_modules/@rslint/core/package.json}';

/** Extension-wide owner for shared UI, channels and workspace runtimes. */
export class Extension {
  private readonly logger: Logger;
  private readonly globalDisposables: Disposable[] = [];
  private outputChannel: OutputChannel | undefined;
  private lspOutputChannel: OutputChannel | undefined;
  private router: WorkspaceDocumentRouter | undefined;
  private runtimeManager: RuntimeManager | undefined;
  private closePromise: Promise<void> | undefined;
  private activated = false;

  constructor(logger: Logger) {
    this.logger = logger;
  }

  public activate(): void {
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
    const runtimeManager = new RuntimeManager(
      router,
      new CoreResolver(),
      (resolved) =>
        new Rslint({
          rootKey: resolved.key,
          workspaceFolder: resolved.workspaceFolder,
          installation: resolved.installation,
          outputChannel,
          traceOutputChannel: lspOutputChannel,
          router,
        }),
      this.logger,
    );
    this.router = router;
    this.runtimeManager = runtimeManager;

    // Install every global facility before awaiting a root. A slow or broken
    // config must not create an event-listener gap or hide commands/channels.
    this.globalDisposables.push(setupStatusBar());
    this.globalDisposables.push(
      ...registerCommands(outputChannel, lspOutputChannel),
    );
    const reconcile = (reason: string) => {
      runtimeManager.clearResolutionCache();
      void runtimeManager.reconcileOpenDocuments().catch((error: unknown) => {
        this.logger.error(
          `Failed to reconcile runtimes after ${reason}`,
          error,
        );
      });
    };
    this.globalDisposables.push(
      workspace.onDidOpenTextDocument((document) => {
        void runtimeManager.reconcile(document).catch((error: unknown) => {
          this.logger.error(
            `Failed to open ${document.uri} with Rslint`,
            error,
          );
        });
      }),
      workspace.onDidCloseTextDocument((document) => {
        runtimeManager.documentClosed(document);
      }),
      workspace.onDidChangeWorkspaceFolders(() => {
        reconcile('workspace-folder change');
      }),
      workspace.onDidChangeConfiguration((event) => {
        if (event.affectsConfiguration('rslint.corePath')) {
          reconcile('configuration change');
        }
      }),
    );
    const topologyWatcher =
      workspace.createFileSystemWatcher(CORE_TOPOLOGY_GLOB);
    topologyWatcher.onDidCreate(() => {
      reconcile('dependency change');
    });
    topologyWatcher.onDidChange(() => {
      reconcile('dependency change');
    });
    topologyWatcher.onDidDelete(() => {
      reconcile('dependency change');
    });
    this.globalDisposables.push(topologyWatcher);

    runtimeManager.initialize(workspace.textDocuments);
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
    // Stop accepting document/topology changes before withdrawing runtimes.
    for (const disposable of this.globalDisposables.splice(0).reverse()) {
      try {
        disposable.dispose();
      } catch (error) {
        errors.push(error);
      }
    }
    try {
      await this.runtimeManager?.close();
    } catch (error) {
      errors.push(error);
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
    this.runtimeManager = undefined;
    this.router = undefined;

    for (const error of errors) {
      this.logger.error('Failed to close an extension resource', error);
    }
    this.logger.info('Rslint extension deactivated');
    if (errors.length > 0) {
      throw new AggregateError(errors, 'failed to deactivate Rslint extension');
    }
  }
}
