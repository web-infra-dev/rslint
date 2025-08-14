import * as vscode from 'vscode';
export function setupStatusBar(context: vscode.ExtensionContext) {
  const statusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );
  statusBar.text = '$(beaker) rslint';
  statusBar.tooltip = 'Rslint Language Server';
  statusBar.command = 'rslint.showMenu';
  statusBar.backgroundColor = new vscode.ThemeColor(
    'statusBarItem.warningBackground',
  );
  statusBar.show();

  context.subscriptions.push(statusBar);
}
