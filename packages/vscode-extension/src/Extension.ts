import {
  window,
  type Disposable,
  type ExtensionContext,
  type OutputChannel,
} from 'vscode';
import { Logger } from './logger';
import { setupStatusBar } from './statusBar';
import { registerCommands } from './commands';
import { RuntimeManager } from './RuntimeManager';

/** Extension-wide owner for shared UI, channels and workspace runtimes. */
export class Extension {
  private readonly logger: Logger;
  private readonly globalDisposables: Disposable[] = [];
  private outputChannel: OutputChannel | undefined;
  private lspOutputChannel: OutputChannel | undefined;
  private runtimeManager: RuntimeManager | undefined;
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

    const runtimeManager = new RuntimeManager({
      outputChannel,
      traceOutputChannel: lspOutputChannel,
      logger: this.logger,
    });
    this.runtimeManager = runtimeManager;

    // Install every global facility before awaiting a root. A slow or broken
    // config must not create an event-listener gap or hide commands/channels.
    this.globalDisposables.push(setupStatusBar());
    this.globalDisposables.push(
      ...registerCommands(outputChannel, lspOutputChannel),
    );
    await runtimeManager.start();
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
    try {
      await this.runtimeManager?.close();
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
    this.runtimeManager = undefined;

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
