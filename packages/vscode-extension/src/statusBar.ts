import * as vscode from 'vscode';
import { RslintBinPath } from './utils';
export function setupStatusBar(context: vscode.ExtensionContext) {
  const statusBar = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Right,
    100,
  );
  const binPathConfig = vscode.workspace
    .getConfiguration()
    .get<RslintBinPath>('rslint.binPath')!;
  statusBar.text = `$(wrench) Rslint`;
  statusBar.tooltip = `Rslint Language Server (${binPathConfig})`;
  statusBar.command = 'rslint.showMenu';
  statusBar.show();

  context.subscriptions.push(statusBar);
}
