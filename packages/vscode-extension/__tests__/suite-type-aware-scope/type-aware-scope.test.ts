import * as assert from 'assert';
import * as vscode from 'vscode';
import path from 'node:path';

// Tests that type-aware rules (e.g. require-await) only run on files covered
// by parserOptions.project, matching CLI behavior.
//
// Fixture: rslint.config.js has project: ['./packages/core/tsconfig.json']
// - packages/core/src/index.ts: IN tsconfig, require-await SHOULD fire
// - packages/cli/src/preview.ts: NOT in tsconfig, require-await should NOT fire
suite('rslint type-aware rule scope', function () {
  this.timeout(60000);

  function getWorkspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  async function waitForDiagnostics(
    doc: vscode.TextDocument,
    predicate?: (diags: vscode.Diagnostic[]) => boolean,
  ): Promise<vscode.Diagnostic[]> {
    for (let i = 0; i < 20; i++) {
      const diagnostics = vscode.languages.getDiagnostics(doc.uri);
      if (predicate ? predicate(diagnostics) : diagnostics.length > 0) {
        return diagnostics;
      }
      await new Promise((resolve) => {
        const disposable = vscode.languages.onDidChangeDiagnostics((e) => {
          for (const uri of e.uris) {
            if (uri.toString() === doc.uri.toString()) {
              disposable.dispose();
              resolve(void 0);
              return;
            }
          }
        });
        setTimeout(() => {
          disposable.dispose();
          resolve(void 0);
        }, 1500);
      });
    }
    return vscode.languages.getDiagnostics(doc.uri);
  }

  test('file IN parserOptions.project tsconfig should get type-aware rules', async () => {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages/core/src/index.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('require-await')),
    );

    // require-await should fire (type-aware, file IS in tsconfig)
    assert.ok(
      diagnostics.some((d) => d.message.includes('require-await')),
      `Expected require-await for file in tsconfig. Got: ${diagnostics.map((d) => d.message).join(', ')}`,
    );

    // no-console should also fire (non-type-aware, always runs)
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-console')),
      'Expected no-console for file in tsconfig',
    );
  });

  test('file NOT in parserOptions.project tsconfig should NOT get type-aware rules', async () => {
    const filePath = path.join(
      getWorkspaceRoot(),
      'packages/cli/src/preview.ts',
    );
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    // Wait for no-console (non-type-aware) to appear — proves rslint IS linting the file
    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('no-console')),
    );

    // no-console SHOULD fire (non-type-aware, always runs)
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-console')),
      `Expected no-console for file outside tsconfig. Got: ${diagnostics.map((d) => d.message).join(', ')}`,
    );

    // require-await should NOT fire (type-aware, file is NOT in configured tsconfig)
    assert.ok(
      !diagnostics.some((d) => d.message.includes('require-await')),
      'require-await should NOT fire for file outside parserOptions.project tsconfig',
    );
  });
});
