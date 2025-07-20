import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint extension', function() {
  this.timeout(50000);
  test('diagnostics', async () => {
    const doc = await vscode.workspace.openTextDocument(
      path.resolve(require.resolve('@rslint/core'), '../..', 'fixtures/src/index.ts'),
    );

    await vscode.window.showTextDocument(doc);
    await new Promise((resolve) => {
      const disposable = vscode.languages.onDidChangeDiagnostics(() => {
        disposable.dispose();
        resolve(void 0);
      });
    });

    const diagnostics = vscode.languages.getDiagnostics(doc.uri);
    assert.ok(diagnostics.length > 0);
  });
});
