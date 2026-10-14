import { ExtensionContext } from 'vscode';
import { Extension } from './Extension';
import { Logger } from './logger';
import {
  createMigrationNotice,
  MigrationNotices,
  rstackEditorTakesOver,
} from './migrationNotice';

let extension: Extension | undefined;

export async function activate(context: ExtensionContext): Promise<void> {
  Logger.setDefaultLogLevel(context);
  const logger = new Logger('Rslint (extension)').useDefaultLogLevel();
  context.subscriptions.push(logger);

  const notices = new MigrationNotices(
    context.globalState,
    context.extension.id,
    logger,
  );
  const standingDown = rstackEditorTakesOver();
  createMigrationNotice(context, standingDown, logger);
  if (standingDown) {
    void notices.showMigrationPrompt(true);
    return;
  }

  extension = new Extension(logger, notices);

  try {
    extension.activate();
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
        { cause: activationError },
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
