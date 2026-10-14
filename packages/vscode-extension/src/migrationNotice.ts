import {
  commands,
  extensions,
  MarkdownString,
  StatusBarAlignment,
  ThemeColor,
  window,
  type ExtensionContext,
  type Memento,
  type StatusBarItem,
} from 'vscode';
import type { Logger } from './logger';

const RSTACK_EXTENSION_ID = 'rstack.rstack';
const OPEN_EXTENSION_COMMAND = 'rslint.openRstackExtension';
const RSTACK_REPO_URL = 'https://github.com/rstackjs/rstack-editor';
const MIGRATION_NOTES_URL =
  'https://github.com/rstackjs/rstack-editor/blob/main/packages/vscode/README.md#coming-from-the-standalone-extensions';

/**
 * The Rstack extension bundles the same Rslint integration as this extension.
 * Its presence alone decides that this extension stands down.
 * `extensions.getExtension` only sees enabled extensions, so a disabled Rstack
 * extension does not take over.
 */
export function rstackEditorTakesOver(): boolean {
  return extensions.getExtension(RSTACK_EXTENSION_ID) !== undefined;
}

/**
 * Migration prompts for the final standalone release. Prompts never block
 * activation, and dismissal flags live in `globalState`, never in settings.
 *
 * 1. Rstack is enabled: this extension stands down and asks to be uninstalled.
 * 2. A resolved core is outside the supported range: its runtime does not
 *    start, and an error offers Install Rstack. It cannot be dismissed.
 * 3. Otherwise: the first runtime that becomes ready offers Install Rstack
 *    once, unless an unsupported core was reported before it.
 */
export class MigrationNotices {
  // Deduplicate core errors for this session; reloading the Extension Host
  // resets the set.
  private readonly shownUnsupportedCoreMessages = new Set<string>();
  private unsupportedCoreSeen = false;
  private installPromptDone = false;

  public constructor(
    private readonly globalState: Memento,
    private readonly extensionId: string,
    private readonly logger: Logger,
  ) {}

  /** Case 3: called when a language server becomes ready. */
  public async runtimeReady(): Promise<void> {
    if (this.installPromptDone) return;
    this.installPromptDone = true;
    if (!this.unsupportedCoreSeen) await this.showMigrationPrompt(false);
  }

  /** Case 2: a resolved core is outside the supported range. */
  public async showUnsupportedCore(message: string): Promise<void> {
    this.unsupportedCoreSeen = true;
    if (this.shownUnsupportedCoreMessages.has(message)) return;
    this.shownUnsupportedCoreMessages.add(message);
    await this.runMigrationAction('Failed to install Rstack', async () => {
      const action = await window.showErrorMessage(message, 'Install Rstack');
      if (action === 'Install Rstack') {
        await this.installRstack();
      }
    });
  }

  /** Case 1 when `standingDown`, otherwise case 3. */
  public async showMigrationPrompt(standingDown: boolean): Promise<void> {
    const flag = standingDown
      ? 'rslint.deprecation.uninstallPromptDismissed'
      : 'rslint.deprecation.installPromptDismissed';
    if (this.globalState.get<boolean>(flag)) return;

    await this.runMigrationAction(
      'Failed to migrate the standalone Rslint extension',
      async () => {
        const action = await window.showWarningMessage(
          standingDown
            ? `[Rstack](${RSTACK_REPO_URL}) has taken over Rslint. Uninstall the Rslint extension.`
            : `Rslint has moved into the [Rstack](${RSTACK_REPO_URL}) extension. The Rslint extension is deprecated and will not be updated. Please install Rstack.`,
          standingDown ? 'Uninstall Rslint' : 'Install Rstack',
          "Don't show again",
        );
        if (action === "Don't show again") {
          await this.globalState.update(flag, true);
        } else if (action === 'Uninstall Rslint') {
          await commands.executeCommand(
            'workbench.extensions.uninstallExtension',
            this.extensionId,
          );
          await offerReload(
            'The standalone Rslint extension was uninstalled. Reload the window to finish.',
          );
        } else if (action === 'Install Rstack') {
          await this.installRstack();
        }
      },
    );
  }

  private async installRstack(): Promise<void> {
    await commands.executeCommand(
      'workbench.extensions.installExtension',
      RSTACK_EXTENSION_ID,
      { enable: true },
    );
    await offerReload(
      `[Rstack](${RSTACK_REPO_URL}) was installed. Reload the window so exactly one copy of Rslint runs.`,
    );
  }

  private async runMigrationAction(
    label: string,
    action: () => Promise<void>,
  ): Promise<void> {
    try {
      await action();
    } catch (error) {
      this.logger.error(label, error);
      void window.showErrorMessage(
        error instanceof Error ? error.message : String(error),
      );
    }
  }
}

async function offerReload(message: string): Promise<void> {
  const action = await window.showInformationMessage(message, 'Reload Window');
  if (action === 'Reload Window') {
    await commands.executeCommand('workbench.action.reloadWindow');
  }
}

type NoticeState = 'migrate' | 'off' | 'reload';

/**
 * Shows a non-modal reminder that the Rstack extension supersedes this one.
 * A running language server cannot be created or torn down safely when the
 * Rstack extension changes, so a state mismatch asks the user to reload.
 */
export function createMigrationNotice(
  context: ExtensionContext,
  standingDown: boolean,
  logger: Logger,
): void {
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

  const notices: Record<
    NoticeState,
    {
      text: string;
      tooltip: MarkdownString;
      command: StatusBarItem['command'];
      log: string;
    }
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
      command: OPEN_EXTENSION_COMMAND,
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
      command: {
        command: OPEN_EXTENSION_COMMAND,
        title: 'Uninstall the standalone Rslint extension',
        arguments: [context.extension.id],
      },
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
      command: 'workbench.action.reloadWindow',
      log: `${RSTACK_EXTENSION_ID} changed after activation. Reload the window so exactly one copy of Rslint runs.`,
    },
  };

  const initial: NoticeState = standingDown ? 'off' : 'migrate';
  let state: NoticeState | undefined;
  const apply = (takesOver: boolean): void => {
    const next: NoticeState = takesOver === standingDown ? initial : 'reload';
    if (next === state) return;

    state = next;
    const notice = notices[next];
    item.text = notice.text;
    item.tooltip = notice.tooltip;
    item.command = notice.command;
    item.show();
    logger.warn(notice.log);
  };

  apply(standingDown);
  context.subscriptions.push(
    extensions.onDidChange(() => {
      apply(rstackEditorTakesOver());
    }),
  );
}
