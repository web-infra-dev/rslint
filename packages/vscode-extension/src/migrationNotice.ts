import {
  commands,
  extensions,
  MarkdownString,
  StatusBarAlignment,
  ThemeColor,
  window,
  workspace,
  type ExtensionContext,
  type StatusBarItem,
} from 'vscode';
import { Logger } from './logger';

const RSTACK_EXTENSION_ID = 'rstack.rstack';
const OPEN_EXTENSION_COMMAND = 'rslint.openRstackExtension';
const MIGRATION_NOTES_URL =
  'https://github.com/rstackjs/rstack-editor/blob/main/packages/vscode/README.md#coming-from-the-standalone-extensions';

/**
 * The Rstack extension bundles the same Rslint integration as this extension.
 * When it is enabled with Rslint switched on, this extension must stand down
 * so only one copy of Rslint runs.
 */
export function rstackEditorTakesOver(): boolean {
  return (
    extensions.getExtension(RSTACK_EXTENSION_ID) !== undefined &&
    workspace.getConfiguration('rstack.rslint').get('enable', true)
  );
}

type NoticeState = 'migrate' | 'off' | 'reload';

const NOTICES: Record<
  NoticeState,
  { text: string; tooltip: MarkdownString; log: string }
> = {
  migrate: {
    text: '$(sparkle-filled) Rslint → Rstack',
    tooltip: new MarkdownString(
      [
        '**The standalone Rslint extension is retired.**',
        '',
        `New editor features land in the unified **Rstack** extension (\`${RSTACK_EXTENSION_ID}\`), which covers testing, linting and formatting. Click to open it in the Extensions view.`,
        '',
        `Settings move from \`rslint.*\` to \`rstack.rslint.*\` and are not migrated automatically — see the [migration notes](${MIGRATION_NOTES_URL}).`,
      ].join('\n'),
    ),
    log: `The standalone Rslint extension is retired; new features land in ${RSTACK_EXTENSION_ID}. Migration notes: ${MIGRATION_NOTES_URL}`,
  },
  off: {
    text: '$(sparkle-filled) Rslint: off',
    tooltip: new MarkdownString(
      [
        `**Rstack (\`${RSTACK_EXTENSION_ID}\`) is running Rslint, so this extension is inactive.**`,
        '',
        'Click to open this extension in the Extensions view and uninstall it.',
      ].join('\n'),
    ),
    log: `${RSTACK_EXTENSION_ID} is active, so the standalone Rslint extension stands down. Uninstall it to remove this notice.`,
  },
  reload: {
    text: '$(sparkle-filled) Rslint: reload window',
    tooltip: new MarkdownString(
      [
        `**Rstack (\`${RSTACK_EXTENSION_ID}\`) changed after this window started.**`,
        '',
        'Click to reload the window so exactly one copy of Rslint runs.',
      ].join('\n'),
    ),
    log: `${RSTACK_EXTENSION_ID} changed after activation. Reload the window so exactly one copy of Rslint runs.`,
  },
};

/**
 * Shows a non-modal reminder that the Rstack extension supersedes this one.
 * A running language server cannot be created or torn down safely when the
 * Rstack extension changes, so a state mismatch asks the user to reload.
 */
export function createMigrationNotice(
  context: ExtensionContext,
  standingDown: boolean,
): void {
  Logger.setDefaultLogLevel(context);
  const logger = new Logger('Rslint migration').useDefaultLogLevel();
  context.subscriptions.push(logger);

  context.subscriptions.push(
    commands.registerCommand(
      OPEN_EXTENSION_COMMAND,
      (extensionId: string = RSTACK_EXTENSION_ID) =>
        commands.executeCommand(
          'workbench.extensions.search',
          `@id:${extensionId}`,
        ),
    ),
  );

  const item = window.createStatusBarItem(
    'rslint.migrationNotice',
    StatusBarAlignment.Right,
  );
  item.name = 'Rslint: Migrate to Rstack';
  item.backgroundColor = new ThemeColor('statusBarItem.warningBackground');
  context.subscriptions.push(item);

  const noticeCommands: Record<NoticeState, StatusBarItem['command']> = {
    migrate: OPEN_EXTENSION_COMMAND,
    off: {
      command: OPEN_EXTENSION_COMMAND,
      title: 'Uninstall the standalone Rslint extension',
      arguments: [context.extension.id],
    },
    reload: 'workbench.action.reloadWindow',
  };

  let state: NoticeState | undefined;
  const apply = (takesOver: boolean): void => {
    const next: NoticeState =
      takesOver !== standingDown ? 'reload' : standingDown ? 'off' : 'migrate';
    if (next === state) return;

    state = next;
    const notice = NOTICES[next];
    item.text = notice.text;
    item.tooltip = notice.tooltip;
    item.command = noticeCommands[next];
    item.show();
    logger.warn(notice.log);
  };

  apply(standingDown);
  context.subscriptions.push(
    extensions.onDidChange(() => {
      apply(rstackEditorTakesOver());
    }),
    workspace.onDidChangeConfiguration((event) => {
      if (event.affectsConfiguration('rstack.rslint.enable')) {
        apply(rstackEditorTakesOver());
      }
    }),
  );
}
