import {
  ExtensionContext,
  Disposable,
  workspace,
  WorkspaceFolder,
  OutputChannel,
  window,
} from 'vscode';
import { State } from 'vscode-languageclient/node';
import { Logger } from './logger';
import { Rslint } from './Rslint';
import { setupStatusBar } from './statusBar';
import { RegisterCommands } from './commands';
import { applyFolderChange, folderKey } from './workspace-folders';

export class Extension implements Disposable {
  private readonly rslintInstances = new Map<string, Rslint>();
  private disposables: Disposable[] = [];
  // Per-instance state-change listener disposables, tracked by id so a
  // folder removal releases its listener. Previously pushed only to
  // extension-lifetime arrays and never released on remove → one leaked
  // Disposable per add/remove cycle.
  private readonly stateChangeDisposables = new Map<string, Disposable>();
  private readonly logger: Logger;
  public context: ExtensionContext;
  // Stashed so onDidChangeWorkspaceFolders can reuse the SAME channels
  // when new folders join (one shared "Rslint Language Server" output
  // channel; per-folder spam in N channels is the wrong UX).
  private outputChannel?: OutputChannel;
  private lspOutputChannel?: OutputChannel;

  constructor(context: ExtensionContext) {
    Logger.setDefaultLogLevel(context);
    this.context = context;
    this.logger = new Logger('Rslint (extension)').useDefaultLogLevel();
  }

  public async activate(): Promise<void> {
    this.logger.info('Rslint extension activating...');

    const folders = workspace.workspaceFolders ?? [];
    this.outputChannel = window.createOutputChannel(
      'Rslint Language Server',
      'log',
    );
    this.lspOutputChannel = window.createOutputChannel(
      'Rslint Language Server(LSP)',
    );

    for (const folder of folders) {
      const workspaceRslint = this.createRslintInstance(
        folderKey(folder),
        folder,
        this.outputChannel,
        this.lspOutputChannel,
      );
      await workspaceRslint.start();
      this.setupStateChangeMonitoring(workspaceRslint, folderKey(folder));
    }

    // React to dynamic add/remove of workspace folders (VS Code's
    // "Add Folder to Workspace…" / "Remove Folder from Workspace").
    // Without this, folders joined post-activation get no Rslint
    // instance (their files never lint until window reload), and
    // removed folders leak their worker + LSP client until the
    // window closes. The logic itself lives in `workspace-folders.ts`
    // as a pure helper so it can be unit-tested without spinning up
    // the VS Code extension host.
    this.disposables.push(
      workspace.onDidChangeWorkspaceFolders(async (event) =>
        applyFolderChange(
          event,
          {
            has: (id) => this.rslintInstances.has(id),
            create: async (folder) => {
              const inst = this.createRslintInstance(
                folderKey(folder),
                folder,
                this.outputChannel!,
                this.lspOutputChannel!,
              );
              await inst.start();
              this.setupStateChangeMonitoring(inst, folderKey(folder));
            },
            remove: async (id) => this.removeRslintInstance(id),
          },
          {
            warn: (m) => {
              this.logger.warn(m);
            },
            error: (m, err) => {
              this.logger.error(m, err);
            },
          },
        ),
      ),
    );

    setupStatusBar(this.context);
    RegisterCommands(this.context, this.outputChannel, this.lspOutputChannel);
    this.logger.info('Rslint extension activated successfully');
  }

  public async deactivate(): Promise<void> {
    this.logger.info('Rslint extension deactivating...');

    const stopPromises = Array.from(this.rslintInstances.values()).map(
      async (instance) => instance.stop(),
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

  public createRslintInstance(
    id: string,
    workspaceFolder: WorkspaceFolder,
    outputChannel: OutputChannel,
    lspOutputChannel: OutputChannel,
  ): Rslint {
    if (this.rslintInstances.has(id)) {
      this.logger.warn(`Rslint instance with id '${id}' already exists`);
      return this.rslintInstances.get(id)!;
    }

    // TODO: single file mode
    const rslint = new Rslint(
      this,
      workspaceFolder,
      outputChannel,
      lspOutputChannel,
    );
    this.rslintInstances.set(id, rslint);
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

    // Release this instance's state-change listener (else it leaks per
    // add/remove cycle).
    this.stateChangeDisposables.get(id)?.dispose();
    this.stateChangeDisposables.delete(id);

    this.logger.debug(`Removed Rslint instance with id: ${id}`);
  }

  public getAllRslintInstances(): Map<string, Rslint> {
    return new Map(this.rslintInstances);
  }

  public dispose(): void {
    this.rslintInstances.forEach((instance) => {
      instance.dispose();
    });
    this.rslintInstances.clear();

    this.stateChangeDisposables.forEach((d) => d.dispose());
    this.stateChangeDisposables.clear();

    this.disposables.forEach((disposable) => {
      disposable.dispose();
    });
    this.disposables = [];

    // Shared output channels are owned here (not per Rslint instance) —
    // dispose them once on extension teardown (#13).
    this.outputChannel?.dispose();
    this.lspOutputChannel?.dispose();

    this.logger.dispose();
  }

  private setupStateChangeMonitoring(rslint: Rslint, instanceId: string): void {
    const stateChangeDisposable = rslint.onDidChangeState((event) => {
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

    this.stateChangeDisposables.set(instanceId, stateChangeDisposable);
  }
}
