import * as assert from 'node:assert';
import * as vscode from 'vscode';
import type { Logger } from '../../src/logger';
import { MigrationNotices } from '../../src/migrationNotice';
import { recordMessages, type RecordedMessage } from '../utils/messages';

// Renaming a key silently resets every user's dismissal, so pin them verbatim.
const UNINSTALL_DISMISSED_KEY = 'rslint.deprecation.uninstallPromptDismissed';
const INSTALL_DISMISSED_KEY = 'rslint.deprecation.installPromptDismissed';

// The notices only log failed actions, which the error toast already reports.
// rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
const silentLogger = { error() {} } as unknown as Logger;

class FakeMemento implements vscode.Memento {
  readonly values = new Map<string, unknown>();

  keys(): readonly string[] {
    return [...this.values.keys()];
  }

  get<T>(key: string, defaultValue?: T): T | undefined {
    return this.values.has(key) ? (this.values.get(key) as T) : defaultValue;
  }

  async update(key: string, value: unknown): Promise<void> {
    this.values.set(key, value);
  }
}

/**
 * Records the extension-management and reload commands instead of running
 * them in the test Extension Host. Other commands still run.
 */
function recordCommands(fail?: Error) {
  const calls: Array<{ command: string; args: unknown[] }> = [];
  const namespace = vscode.commands as {
    executeCommand: typeof vscode.commands.executeCommand;
  };
  const original = namespace.executeCommand;
  namespace.executeCommand = async <T>(
    command: string,
    ...args: unknown[]
  ): Promise<T> => {
    if (
      !command.startsWith('workbench.extensions.') &&
      command !== 'workbench.action.reloadWindow'
    ) {
      return original<T>(command, ...args);
    }
    calls.push({ command, args });
    if (fail) throw fail;
    return undefined as T;
  };
  return {
    calls,
    restore() {
      namespace.executeCommand = original;
    },
  };
}

suite('Rstack migration notices', () => {
  let memento: FakeMemento;
  let notices: MigrationNotices;
  let answer: (message: RecordedMessage) => string | undefined;
  let messages: ReturnType<typeof recordMessages>;
  let commands: ReturnType<typeof recordCommands>;

  function newNotices(): MigrationNotices {
    return new MigrationNotices(memento, 'rstack.rslint', silentLogger);
  }

  setup(() => {
    memento = new FakeMemento();
    notices = newNotices();
    answer = () => undefined;
    messages = recordMessages((message) => answer(message));
    commands = recordCommands();
  });

  teardown(() => {
    messages.restore();
    commands.restore();
  });

  test('opens the Rstack extension in the Extensions view', async () => {
    await vscode.commands.executeCommand('rslint.openRstackExtension');
    assert.deepStrictEqual(commands.calls, [
      { command: 'workbench.extensions.search', args: ['@id:rstack.rstack'] },
    ]);
  });

  test('case 1 offers to uninstall this extension, then a reload', async () => {
    answer = (shown) =>
      shown.level === 'warning' ? 'Uninstall Rslint' : undefined;
    await notices.showMigrationPrompt(true);

    assert.strictEqual(messages.messages[0].level, 'warning');
    assert.match(messages.messages[0].message, /has taken over Rslint/);
    assert.deepStrictEqual(messages.messages[0].items, [
      'Uninstall Rslint',
      "Don't show again",
    ]);
    assert.deepStrictEqual(commands.calls, [
      {
        command: 'workbench.extensions.uninstallExtension',
        args: ['rstack.rslint'],
      },
    ]);
    assert.strictEqual(messages.messages[1]?.level, 'information');
    assert.deepStrictEqual(messages.messages[1]?.items, ['Reload Window']);
  });

  test('case 1 is suppressed by its globalState flag', async () => {
    answer = () => "Don't show again";
    await notices.showMigrationPrompt(true);
    assert.strictEqual(memento.get(UNINSTALL_DISMISSED_KEY), true);
    assert.strictEqual(memento.get(INSTALL_DISMISSED_KEY), undefined);

    messages.messages.length = 0;
    await newNotices().showMigrationPrompt(true);
    assert.deepStrictEqual(messages.messages, []);
  });

  test('case 2 shows an Install Rstack error that the flags cannot dismiss', async () => {
    await memento.update(UNINSTALL_DISMISSED_KEY, true);
    await memento.update(INSTALL_DISMISSED_KEY, true);
    answer = (shown) =>
      shown.level === 'error' ? 'Install Rstack' : undefined;
    await notices.showUnsupportedCore('unsupported core');

    assert.deepStrictEqual(messages.messages[0], {
      level: 'error',
      message: 'unsupported core',
      items: ['Install Rstack'],
    });
    assert.deepStrictEqual(commands.calls, [
      {
        command: 'workbench.extensions.installExtension',
        args: ['rstack.rstack', { enable: true }],
      },
    ]);
    assert.strictEqual(messages.messages[1]?.level, 'information');
  });

  test('case 2 shows each unsupported-core message once per session', async () => {
    // Two workspace folders sharing one core report the same message.
    await notices.showUnsupportedCore('core 0.10.0');
    await notices.showUnsupportedCore('core 0.10.0');
    await notices.showUnsupportedCore('core 0.8.2');

    assert.deepStrictEqual(
      messages.messages.map((shown) => shown.message),
      ['core 0.10.0', 'core 0.8.2'],
    );
  });

  test('case 3 prompts once after the first ready runtime', async () => {
    answer = (shown) =>
      shown.level === 'warning' ? 'Install Rstack' : undefined;
    await notices.runtimeReady();
    await notices.runtimeReady();

    const warnings = messages.messages.filter(
      (shown) => shown.level === 'warning',
    );
    assert.strictEqual(warnings.length, 1);
    assert.match(warnings[0].message, /The Rslint extension is deprecated/);
    assert.deepStrictEqual(warnings[0].items, [
      'Install Rstack',
      "Don't show again",
    ]);
    assert.deepStrictEqual(commands.calls, [
      {
        command: 'workbench.extensions.installExtension',
        args: ['rstack.rstack', { enable: true }],
      },
    ]);
  });

  test('case 3 is suppressed by its globalState flag', async () => {
    answer = () => "Don't show again";
    await notices.runtimeReady();
    assert.strictEqual(memento.get(INSTALL_DISMISSED_KEY), true);

    messages.messages.length = 0;
    await newNotices().runtimeReady();
    assert.deepStrictEqual(messages.messages, []);
    assert.deepStrictEqual(commands.calls, []);
  });

  test('case 3 is suppressed after an unsupported core', async () => {
    await notices.showUnsupportedCore('unsupported core');
    await notices.runtimeReady();

    assert.deepStrictEqual(
      messages.messages.map((shown) => shown.level),
      ['error'],
    );
  });

  test('reports a failed action instead of throwing', async () => {
    commands.restore();
    commands = recordCommands(new Error('uninstall failed'));
    answer = (shown) =>
      shown.level === 'warning' ? 'Uninstall Rslint' : undefined;
    await notices.showMigrationPrompt(true);

    assert.deepStrictEqual(messages.messages.at(-1), {
      level: 'error',
      message: 'uninstall failed',
      items: [],
    });
  });
});
