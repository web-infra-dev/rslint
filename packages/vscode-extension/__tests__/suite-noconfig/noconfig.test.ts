import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';
import fs from 'node:fs';
import { waitForRslintDiagnostics as waitForDiagnostics } from '../utils/diagnostics';

suite('rslint no-config fallback', function () {
  this.timeout(120000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function replaceDocument(
    editor: vscode.TextEditor,
    content: string,
  ): Promise<void> {
    const document = editor.document;
    const end = document.positionAt(document.getText().length);
    const applied = await editor.edit((edit) => {
      edit.replace(new vscode.Range(new vscode.Position(0, 0), end), content);
    });
    assert.ok(applied, 'Failed to update the no-config fixture document');
  }

  async function triggerDiagnosticRefresh(
    editor: vscode.TextEditor,
  ): Promise<void> {
    const inserted = await editor.edit((edit) => {
      edit.insert(new vscode.Position(0, 0), ' ');
    });
    const deleted = await editor.edit((edit) => {
      edit.delete(new vscode.Range(0, 0, 0, 1));
    });
    assert.ok(inserted && deleted, 'Failed to trigger a diagnostic refresh');
  }

  test('legacy JSON is ignored across the module-config lifecycle', async () => {
    const workspace = getWorkspaceRoot();
    const jsonPath = path.join(workspace, 'rslint.json');
    const modulePath = path.join(workspace, 'rslint.config.js');
    const document = await vscode.workspace.openTextDocument(
      path.join(workspace, 'src', 'index.ts'),
    );
    const editor = await vscode.window.showTextDocument(document);
    const original = document.getText();

    const removeConfigs = (): void => {
      fs.rmSync(jsonPath, { force: true });
      fs.rmSync(modulePath, { force: true });
    };
    removeConfigs();

    try {
      // A malformed legacy file is a stronger sentinel than a valid one: any
      // accidental runtime read would surface a parse failure or diagnostics.
      fs.writeFileSync(jsonPath, '{ intentionally malformed legacy config');
      await replaceDocument(editor, `${original}\nconst broken = ;\n`);
      let diagnostics = await waitForDiagnostics(
        document,
        (items) => items.length > 0,
      );
      assert.ok(
        diagnostics.every(
          (diagnostic) => !diagnostic.message.includes('no-explicit-any'),
        ),
        'Legacy JSON unexpectedly configured a rule while syntax-only fallback was active',
      );

      await replaceDocument(editor, original);
      diagnostics = await waitForDiagnostics(
        document,
        (items) => items.length === 0,
      );
      assert.strictEqual(
        diagnostics.length,
        0,
        'No-config fallback should publish an empty result for valid source',
      );

      const configured = waitForDiagnostics(document, (items) =>
        items.some((diagnostic) =>
          diagnostic.message.includes('no-explicit-any'),
        ),
      );
      fs.writeFileSync(
        modulePath,
        `export default [{
  plugins: ['@typescript-eslint'],
  rules: { '@typescript-eslint/no-explicit-any': 'error' },
}];\n`,
      );
      await triggerDiagnosticRefresh(editor);
      diagnostics = await configured;
      assert.ok(
        diagnostics.some((diagnostic) =>
          diagnostic.message.includes('no-explicit-any'),
        ),
        'Creating a module config should enable its configured rule',
      );

      const cleared = waitForDiagnostics(
        document,
        (items) => items.length === 0,
      );
      fs.rmSync(modulePath);
      await triggerDiagnosticRefresh(editor);
      diagnostics = await cleared;
      assert.strictEqual(
        diagnostics.length,
        0,
        'Deleting the module config must not fall back to legacy JSON',
      );
    } finally {
      await replaceDocument(editor, original);
      removeConfigs();
    }
  });
});
