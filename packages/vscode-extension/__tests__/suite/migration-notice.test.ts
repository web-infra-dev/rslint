import * as assert from 'node:assert';
import * as vscode from 'vscode';
import { rstackEditorTakesOver } from '../../src/migrationNotice';

suite('Rstack migration notice', () => {
  test('opens the Rstack extension without throwing', async () => {
    await vscode.commands.executeCommand('rslint.openRstackExtension');
  });

  test('keeps Rslint active when the Rstack extension is absent', () => {
    // The isolated Extension Host has no Rstack extension installed.
    assert.strictEqual(rstackEditorTakesOver(), false);
  });
});
