import { workspace, ExtensionContext, window, Uri } from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from 'vscode-languageclient/node';
import { logger, LogLevel } from './logger';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
  const isDevelopment = context.extensionMode === 1; // Development mode
  logger.setLogLevel(isDevelopment ? LogLevel.DEBUG : LogLevel.INFO);

  logger.info('Rslint extension activating...');

  const binPathConfig = workspace
    .getConfiguration()
    .get('rslint.binPath') as string;
  const binPath =
    binPathConfig && binPathConfig.trim() !== ''
      ? binPathConfig
      : Uri.joinPath(context.extensionUri, 'dist', 'rslint').fsPath;

  logger.debug('Rslint binary path:', binPath);
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

  client = new LanguageClient(
    'rslint',
    'Rslint Language Server',
    serverOptions,
    clientOptions,
  );

  client
    .start()
    .then(() => {
      logger.info('Rslint language client started successfully');
    })
    .catch((err: unknown) => {
      logger.error('Failed to start Rslint language client', err);
    });

  context.subscriptions.push(
    client.onDidChangeState(event => {
      logger.debug(
        'Rslint client state changed:',
        event.oldState,
        '->',
        event.newState,
      );
      if (event.newState === 2) {
        logger.info('Rslint language server started');
      }
    }),
  );

  context.subscriptions.push({
    dispose: () => logger.dispose(),
  });
}

export function deactivate(): Thenable<void> | undefined {
  logger.info('Rslint extension deactivating...');

  if (!client) {
    return undefined;
  }

  return client
    .stop()
    .then(() => {
      logger.info('Rslint language client stopped');
    })
    .catch((err: unknown) => {
      logger.error('Error stopping Rslint language client', err);
    });
}
