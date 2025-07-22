import * as path from 'path';
import { workspace, ExtensionContext, window, Uri } from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
  
  const binPathConfig = workspace.getConfiguration().get('rslint.binPath') as string;
  const binPath = binPathConfig && binPathConfig.trim() !== '' ? binPathConfig : Uri.joinPath(context.extensionUri,'out','rslint').fsPath;
  const run: Executable  = {
    command: binPath,
    args: ["--lsp"]
  }
  const serverOptions: ServerOptions = {
    run,
    debug: run
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: 'file', language: 'typescript' },
      { scheme: 'file', language: 'typescriptreact' },
      { scheme: 'file', language: 'javascript' },
      { scheme: 'file', language: 'javascriptreact' }
    ],
    synchronize: {
      fileEvents: workspace.createFileSystemWatcher('**/.clientrc')
    }
  };

  client = new LanguageClient(
    'rslint',
    'Rslint Language Server',
    serverOptions,
    clientOptions
  );

  client.start();

  context.subscriptions.push(
    client.onDidChangeState((event) => {
      if (event.newState === 2) {
        window.showInformationMessage('Rslint language server started');
      }
    })
  );
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}