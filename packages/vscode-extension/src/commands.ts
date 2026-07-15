import * as vscode from 'vscode';
export function registerCommands(
  outputChannel: vscode.OutputChannel,
  traceOutputChannel: vscode.OutputChannel,
): vscode.Disposable[] {
  const disposables: vscode.Disposable[] = [];
  try {
    disposables.push(
      vscode.commands.registerCommand('rslint.showMenu', showCommands),
    );
    // vscode.commands.registerCommand('rslint.restart', async () => {
    //   await vscode.commands.executeCommand('rslint.restart');
    // }),
    disposables.push(
      vscode.commands.registerCommand('rslint.output.focus', () => {
        outputChannel.show();
      }),
    );
    disposables.push(
      vscode.commands.registerCommand('rslint.lsp-trace.focus', () => {
        traceOutputChannel.show();
      }),
    );
    return disposables;
  } catch (error) {
    for (const disposable of disposables.reverse()) disposable.dispose();
    throw error;
  }
}

async function showCommands(): Promise<void> {
  const commands: readonly {
    label: string;
    description: string;
    command: string;
  }[] = [
    // {
    //     label: "$(refresh) RestartRslint Server",
    //     description: "Restart the Rslint language server",
    //     command: "rslint.restart",
    // },
    {
      label: '$(output) Show Rslint Server Log',
      description: 'Show the Rslint server log',
      command: 'rslint.output.focus',
    },
    {
      label: '$(debug-console) Show Rslint LSP Messages',
      description: 'Show the LSP communication trace',
      command: 'rslint.lsp-trace.focus',
    },
  ];

  // eslint-disable-next-line @typescript-eslint/await-thenable
  const selected = await vscode.window.showQuickPick(commands, {
    placeHolder: 'Rslint Commands',
  });

  if (selected) {
    // eslint-disable-next-line
    await vscode.commands.executeCommand(selected.command);
  }
}
