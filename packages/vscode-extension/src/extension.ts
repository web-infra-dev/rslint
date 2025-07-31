import { workspace, ExtensionContext, window, Uri, commands } from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from 'vscode-languageclient/node';

let client: LanguageClient;

async function createClient(
  context: ExtensionContext,
): Promise<LanguageClient> {
  const binPathConfig = workspace
    .getConfiguration()
    .get('rslint.binPath') as string;
  const binPath =
    binPathConfig && binPathConfig.trim() !== ''
      ? binPathConfig
      : Uri.joinPath(context.extensionUri, 'dist', 'rslint').fsPath;
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
      fileEvents: workspace.createFileSystemWatcher('**/.clientrc'),
    },
  };

  return new LanguageClient(
    'rslint',
    'Rslint Language Server',
    serverOptions,
    clientOptions,
  );
}

async function startClient(): Promise<void> {
  if (client) {
    await client.start();
  }
}

async function stopClient(): Promise<void> {
  if (client) {
    await client.stop();
  }
}

async function restartServer(context: ExtensionContext): Promise<void> {
  try {
    window.showInformationMessage('Restarting Rslint language server...');

    // Stop the current client
    await stopClient();

    // Create a new client
    client = await createClient(context);

    // Start the new client
    await startClient();

    window.showInformationMessage(
      'Rslint language server restarted successfully',
    );
  } catch (error) {
    window.showErrorMessage(
      `Failed to restart Rslint language server: ${error}`,
    );
  }
}

export async function activate(context: ExtensionContext): Promise<void> {
  // Create and start the client
  client = await createClient(context);
  await startClient();

  // Register the restart server command
  const restartCommand = commands.registerCommand(
    'rslint.restartServer',
    () => {
      void restartServer(context);
    },
  );

  context.subscriptions.push(
    restartCommand,
    client.onDidChangeState(event => {
      if (event.newState === 2) {
        window.showInformationMessage('Rslint language server started');
      }
    }),
  );
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}
