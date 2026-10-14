import * as assert from 'node:assert';
import path from 'node:path';
import * as vscode from 'vscode';
import { waitForRslintDiagnostics } from '../utils/diagnostics';
import {
  recordMessages,
  waitForMessages,
  type RecordedMessage,
} from '../utils/messages';

function workspaceFile(folderName: string, relativePath: string): string {
  const folder = vscode.workspace.workspaceFolders?.find(
    (candidate) => candidate.name === folderName,
  );
  assert.ok(folder, `workspace folder ${folderName} is unavailable`);
  return path.join(folder.uri.fsPath, relativePath);
}

function installPrompts(
  messages: readonly RecordedMessage[],
): RecordedMessage[] {
  return messages.filter(
    (message) =>
      message.level === 'warning' &&
      message.message.includes('The Rslint extension is deprecated'),
  );
}

suite('Rstack install prompt (case 3)', function () {
  this.timeout(120_000);
  let recorder: ReturnType<typeof recordMessages>;

  // runSuite only opens an untitled probe document, which has no workspace
  // folder, so no language server has started before this hook.
  suiteSetup(() => {
    recorder = recordMessages();
  });

  suiteTeardown(() => {
    recorder.restore();
  });

  test('prompts once after the first language server starts', async () => {
    const parent = await vscode.workspace.openTextDocument(
      workspaceFile('parent', 'src/index.ts'),
    );
    await waitForRslintDiagnostics(parent);
    await waitForMessages(
      recorder.messages,
      (messages) => installPrompts(messages).length > 0,
      'the install prompt',
    );

    // A second workspace folder starts its own language server.
    const sentinel = await vscode.workspace.openTextDocument(
      workspaceFile('sentinel', 'src/index.ts'),
    );
    await waitForRslintDiagnostics(sentinel);

    const prompts = installPrompts(recorder.messages);
    assert.strictEqual(prompts.length, 1);
    assert.deepStrictEqual(prompts[0].items, [
      'Install Rstack',
      "Don't show again",
    ]);
    assert.deepStrictEqual(
      recorder.messages.filter((message) => message.level === 'error'),
      [],
    );
  });
});
