import * as assert from 'assert';
import * as vscode from 'vscode';
import fs from 'node:fs';
import path from 'node:path';

// End-to-end VS Code coverage for the object-form `plugins` reverse-dispatch
// path: the LSP server lints natively but dispatches rules mounted via a
// config's object-form `plugins` to the extension-side worker pool
// (PluginLintPool), then merges + publishes. The Go merge/dispatch units are
// covered in internal/lsp/eslint_plugin_test.go; this exercises the full loop.
//
// Fixture: rslint.config.mjs mounts ./local-plugin.mjs under object-form
// `plugins` with rules local/no-null + local/prefer-array-some, plus native
// no-console.
suite('rslint object-form plugins integration', function () {
  this.timeout(120000);

  function workspaceRoot(): string {
    return vscode.workspace.workspaceFolders![0].uri.fsPath;
  }

  // LSP diagnostic messages are formatted as `[<ruleName>] <description>`
  // (see internal/lsp/service.go), so ruleName is matchable on `.message`.
  function messages(diags: vscode.Diagnostic[]): string {
    return diags.map((d) => d.message).join(' | ');
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

  test('mounted plugin rules report, merged with native rules', async () => {
    const filePath = path.join(workspaceRoot(), 'src', 'index.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc, (diags) =>
      diags.some((d) => d.message.includes('local/no-null')),
    );
    const msgs = messages(diagnostics);

    // Both plugin rules must come back from the worker...
    assert.ok(
      diagnostics.some((d) => d.message.includes('local/no-null')),
      `Expected local/no-null. Got: ${msgs}`,
    );
    assert.ok(
      diagnostics.some((d) => d.message.includes('local/prefer-array-some')),
      `Expected local/prefer-array-some. Got: ${msgs}`,
    );
    // ...alongside the natively-linted rule, proving the merge.
    assert.ok(
      diagnostics.some((d) => d.message.includes('no-console')),
      `Expected native no-console merged with plugin diagnostics. Got: ${msgs}`,
    );
  });

  test('plugin auto-fix participates in source.fixAll on save', async () => {
    const tmpFile = path.join(
      workspaceRoot(),
      'src',
      `_fixall_plugin_${Date.now()}.ts`,
    );
    fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

    const editorConfig = vscode.workspace.getConfiguration('editor');
    const previous = editorConfig.get('codeActionsOnSave');
    try {
      await editorConfig.update(
        'codeActionsOnSave',
        { 'source.fixAll': 'explicit' },
        vscode.ConfigurationTarget.Workspace,
      );

      const doc = await vscode.workspace.openTextDocument(tmpFile);
      const editor = await vscode.window.showTextDocument(doc);
      await editor.edit((b) =>
        b.replace(
          new vscode.Range(
            doc.positionAt(0),
            doc.positionAt(doc.getText().length),
          ),
          'const numbers = [1, 2, 3];\nconst ok = numbers.filter((n) => n > 0).length > 0;\nexport { ok };\n',
        ),
      );

      const diagnostics = await waitForDiagnostics(doc, (diags) =>
        diags.some((d) => d.message.includes('local/prefer-array-some')),
      );
      assert.ok(
        diagnostics.some((d) => d.message.includes('local/prefer-array-some')),
        `prefer-array-some did not appear; cannot exercise fixAll. Got: ${messages(diagnostics)}`,
      );

      assert.ok(
        await doc.save(),
        'Document should complete the plugin code-action-on-save pipeline',
      );

      const start = Date.now();
      while (doc.getText().includes('.filter(') && Date.now() - start < 20000) {
        await new Promise((r) => setTimeout(r, 500));
      }

      assert.ok(
        !doc.getText().includes('.filter('),
        `Plugin fix (filter -> some) should apply via source.fixAll on save.\nContent: ${doc.getText()}`,
      );
      assert.ok(
        doc.getText().includes('.some('),
        `Expected '.some(' after fixAll.\nContent: ${doc.getText()}`,
      );
    } finally {
      await editorConfig.update(
        'codeActionsOnSave',
        previous,
        vscode.ConfigurationTarget.Workspace,
      );
      if (vscode.window.activeTextEditor?.document.uri.fsPath === tmpFile) {
        await vscode.commands.executeCommand(
          'workbench.action.closeActiveEditor',
        );
      }
      if (fs.existsSync(tmpFile)) fs.unlinkSync(tmpFile);
    }
  });
});
