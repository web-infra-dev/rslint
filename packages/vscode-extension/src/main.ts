import { ExtensionContext } from 'vscode';
import { Extension } from './Extension';

let extension: Extension;

export async function activate(context: ExtensionContext): Promise<void> {
  extension = new Extension(context);

  try {
    await extension.activate();
    context.subscriptions.push(extension);
  } catch (error) {
    extension?.dispose();
    throw error;
  }
}

export async function deactivate(): Promise<void> {
  if (extension) {
    await extension.deactivate();
  }
}
