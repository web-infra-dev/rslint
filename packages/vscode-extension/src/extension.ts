import {
  workspace,
  ExtensionContext,
  window,
  Uri,
  languages,
  CodeActionProvider,
  CodeAction,
  CodeActionKind,
  Range,
  TextDocument,
  WorkspaceEdit,
  Position,
  Diagnostic,
  CodeActionContext,
  CancellationToken,
} from 'vscode';
import {
  Executable,
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from 'vscode-languageclient/node';

let client: LanguageClient;

// Simple code action provider for disable rule actions
class RslintCodeActionProvider implements CodeActionProvider {
  public provideCodeActions(
    document: TextDocument,
    range: Range,
    context: CodeActionContext,
    token: CancellationToken,
  ): CodeAction[] {
    console.log('RslintCodeActionProvider.provideCodeActions called');
    console.log('Context diagnostics:', context.diagnostics.length);
    console.log(
      'Diagnostics:',
      context.diagnostics.map(d => ({ source: d.source, message: d.message })),
    );

    const actions: CodeAction[] = [];

    // Get diagnostics in the current range from context
    // Check for RSLint diagnostics by looking for the [@typescript-eslint/...] pattern in the message
    const diagnostics = context.diagnostics.filter(
      d => d.source === 'rslint' || d.message.includes('[@typescript-eslint/'),
    );

    console.log('RSLint diagnostics:', diagnostics.length);

    if (diagnostics.length === 0) {
      return actions;
    }

    // For each diagnostic, provide disable actions
    for (const diagnostic of diagnostics) {
      // Extract rule name from diagnostic message [ruleName] format
      const ruleMatch = diagnostic.message.match(/\[([^\]]+)\]/);
      const ruleName = ruleMatch ? ruleMatch[1] : 'rule';

      // Create "Disable rule for this line" action
      const disableLineAction = new CodeAction(
        `Disable ${ruleName} for this line`,
        CodeActionKind.QuickFix,
      );
      disableLineAction.edit = new WorkspaceEdit();
      disableLineAction.diagnostics = [diagnostic];

      // Get the line of the diagnostic
      const line = diagnostic.range.start.line;
      const lineStart = new Position(line, 0);

      // Add eslint-disable-next-line comment
      disableLineAction.edit.insert(
        document.uri,
        lineStart,
        `// eslint-disable-next-line ${ruleName}\n`,
      );
      disableLineAction.isPreferred = false;
      actions.push(disableLineAction);

      // Create "Disable rule for entire file" action
      const disableFileAction = new CodeAction(
        `Disable ${ruleName} for entire file`,
        CodeActionKind.QuickFix,
      );
      disableFileAction.edit = new WorkspaceEdit();
      disableFileAction.diagnostics = [diagnostic];

      // Add eslint-disable comment at the beginning of the file
      const fileStart = new Position(0, 0);
      disableFileAction.edit.insert(
        document.uri,
        fileStart,
        `/* eslint-disable ${ruleName} */\n`,
      );
      disableFileAction.isPreferred = false;
      actions.push(disableFileAction);
    }

    return actions;
  }
}

export function activate(context: ExtensionContext) {
  console.log('RSLint extension activating...');

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
    middleware: {
      provideCodeActions: async (document, range, context, token, next) => {
        // Call the LSP server's code actions
        const actions = await next(document, range, context, token);
        console.log('LSP code actions:', actions?.length || 0);
        return actions;
      },
    },
  };

  client = new LanguageClient(
    'rslint',
    'Rslint Language Server',
    serverOptions,
    clientOptions,
  );

  console.log('Starting RSLint language client...');
  client.start();

  // Register code action provider
  const codeActionProvider = new RslintCodeActionProvider();
  const codeActionProviderDisposable = languages.registerCodeActionsProvider(
    [
      { scheme: 'file', language: 'typescript' },
      { scheme: 'file', language: 'typescriptreact' },
      { scheme: 'file', language: 'javascript' },
      { scheme: 'file', language: 'javascriptreact' },
    ],
    codeActionProvider,
    {
      providedCodeActionKinds: [CodeActionKind.QuickFix],
    },
  );

  context.subscriptions.push(
    codeActionProviderDisposable,
    client.onDidChangeState(event => {
      console.log(
        'RSLint client state changed:',
        event.oldState,
        '->',
        event.newState,
      );
      if (event.newState === 2) {
        console.log('RSLint language server is running');
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
