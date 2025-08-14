import * as vscode from 'vscode';
export function RegisterCommands(
  context: vscode.ExtensionContext,
  outputChannel: vscode.OutputChannel,
  traceOutputChannel: vscode.OutputChannel,
) {
  context.subscriptions.push(
    vscode.commands.registerCommand('rslint.showMenu', showCommands),
  );
  // context.subscriptions.push(vscode.commands.registerCommand('rslint.restart', async () => {
  //     await vscode.commands.executeCommand('rslint.restart');
  // }));
  context.subscriptions.push(
    vscode.commands.registerCommand('rslint.output.focus', () => {
      outputChannel.show();
    }),
  );
  context.subscriptions.push(
    vscode.commands.registerCommand('rslint.lsp-trace.focus', () => {
      traceOutputChannel.show();
    }),
  );
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
