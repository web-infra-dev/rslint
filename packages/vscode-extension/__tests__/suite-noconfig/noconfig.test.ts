import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

suite('rslint no config fallback', function () {
  this.timeout(60000);

  async function openFixture(filename: string): Promise<vscode.TextDocument> {
    const workspaceRoot = vscode.workspace.workspaceFolders![0].uri.fsPath;
    const filePath = path.join(workspaceRoot, 'src', filename);
    return vscode.workspace.openTextDocument(filePath);
  }

  test('extension should start without config file and produce no diagnostics', async () => {
    const doc = await openFixture('index.ts');
    await vscode.window.showTextDocument(doc);

    // Wait a reasonable time for diagnostics to appear (they shouldn't)
    await new Promise(resolve => setTimeout(resolve, 5000));

    const diagnostics = vscode.languages.getDiagnostics(doc.uri);

    // Without any config, no rslint rules are enabled, so no diagnostics
    assert.strictEqual(
      diagnostics.filter(d => d.source === 'rslint').length,
      0,
      'Should have no rslint diagnostics when no config file exists',
    );
  });
});
