import * as assert from 'node:assert';
import fs from 'node:fs/promises';
import path from 'node:path';
import * as vscode from 'vscode';
import { recordMessages, waitForMessages } from '../utils/messages';

suite('unsupported @rslint/core (case 2)', function () {
  this.timeout(120_000);
  let recorder: ReturnType<typeof recordMessages>;
  let workspaceRoot: string;

  // runSuite only opens an untitled probe document, which has no workspace
  // folder, so no core has been resolved before this hook.
  suiteSetup(async () => {
    recorder = recordMessages();
    const folder = vscode.workspace.workspaceFolders?.[0];
    assert.ok(folder, 'test requires a workspace folder');
    workspaceRoot = folder.uri.fsPath;
    // Only package.json exists: if the guard did not run before module
    // loading, resolution would fail with a missing ./config-loader export.
    const fakeCore = path.join(workspaceRoot, 'fake-core');
    await fs.mkdir(fakeCore, { recursive: true });
    await fs.writeFile(
      path.join(fakeCore, 'package.json'),
      JSON.stringify({ name: '@rslint/core', version: '0.10.0' }),
    );
    await vscode.workspace
      .getConfiguration('rslint')
      .update('corePath', './fake-core', vscode.ConfigurationTarget.Workspace);
  });

  suiteTeardown(() => {
    recorder.restore();
  });

  test('shows the Install Rstack error and starts no server', async () => {
    const isUnsupportedCoreError = (message: {
      level: string;
      message: string;
    }) =>
      message.level === 'error' &&
      message.message.includes(
        'only supports @rslint/core ^0.9.0 (found 0.10.0)',
      );

    const document = await vscode.workspace.openTextDocument(
      path.join(workspaceRoot, 'src', 'index.ts'),
    );
    await vscode.window.showTextDocument(document);
    await waitForMessages(
      recorder.messages,
      (messages) => messages.some(isUnsupportedCoreError),
      'the unsupported-core error',
    );

    // Reopening the document resolves the same core again without a repeat.
    await vscode.commands.executeCommand('workbench.action.closeAllEditors');
    await vscode.window.showTextDocument(
      await vscode.workspace.openTextDocument(
        path.join(workspaceRoot, 'src', 'index.ts'),
      ),
    );
    await new Promise((resolve) => setTimeout(resolve, 2_000));

    const errors = recorder.messages.filter(isUnsupportedCoreError);
    assert.strictEqual(errors.length, 1);
    assert.deepStrictEqual(errors[0].items, ['Install Rstack']);
    // Neither the generic resolver warning nor the case 3 prompt appears.
    assert.deepStrictEqual(
      recorder.messages.filter((message) => message.level === 'warning'),
      [],
    );
    assert.deepStrictEqual(
      vscode.languages
        .getDiagnostics(document.uri)
        .filter((diagnostic) => diagnostic.source === 'rslint'),
      [],
    );
  });
});
