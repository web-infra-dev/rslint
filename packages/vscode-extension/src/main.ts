import { ExtensionContext } from 'vscode';
import { Extension } from './Extension';

let extension: Extension | undefined;

export async function activate(context: ExtensionContext): Promise<void> {
  extension = new Extension(context);

  try {
    await extension.activate();
  } catch (activationError) {
    let closeError: unknown;
    try {
      await extension.close();
    } catch (error) {
      closeError = error;
    }
    extension = undefined;
    if (closeError !== undefined) {
      throw new AggregateError(
        [activationError, closeError],
        'Rslint activation and partial-start cleanup both failed',
      );
    }
    throw activationError;
  }
}

export async function deactivate(): Promise<void> {
  const activeExtension = extension;
  extension = undefined;
  await activeExtension?.deactivate();
}
