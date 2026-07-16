import * as vscode from 'vscode';
export function setupStatusBar(): vscode.StatusBarItem {
  const statusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );
  try {
    statusBar.text = `$(wrench) Rslint`;
    statusBar.tooltip = 'Rslint Language Server';
    statusBar.command = 'rslint.showMenu';
    statusBar.show();
    return statusBar;
  } catch (error) {
    statusBar.dispose();
    throw error;
  }
}
