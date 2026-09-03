import * as assert from 'node:assert';
import * as vscode from 'vscode';
import { rstackEditorTakesOver } from '../../src/migrationNotice';

suite('Rstack migration notice', () => {
  test('opens the Rstack extension without throwing', async () => {
    await vscode.commands.executeCommand('rslint.openRstackExtension');
  });

  test('keeps Rslint active when the Rstack extension is absent', () => {
    assert.strictEqual(
      vscode.extensions.getExtension('rstack.rstack'),
      undefined,
      'The isolated Extension Host must not install the Rstack extension',
    );
    assert.strictEqual(rstackEditorTakesOver(), false);
  });
});
