import * as assert from 'assert';
import * as vscode from 'vscode';
import fs from 'node:fs';
import path from 'node:path';
import { waitForContentChange } from '../suite/fixall-helpers';
import { saveDocumentOnce } from '../utils/codeActionRegistry';
import { withCodeActionsOnSave } from '../utils/configuration';
import { waitForRslintDiagnostics } from '../utils/diagnostics';
import {
  closeAndDeleteTemporaryDocument,
  temporaryFilePath,
} from '../utils/documents';

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
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) throw new Error('Test workspace is unavailable');
    return workspaceFolder.uri.fsPath;
  }

  // LSP diagnostic messages are formatted as `[<ruleName>] <description>`
  // (see internal/lsp/service.go), so ruleName is matchable on `.message`.
  function messages(diags: vscode.Diagnostic[]): string {
    return diags.map((d) => d.message).join(' | ');
  }

  // Keep the original flaky scenario first: no preceding test may warm the
  // diagnostics or code-action-on-save path in this fresh extension host.
  test('plugin auto-fix participates in source.fixAll on save', async () => {
    const tmpFile = temporaryFilePath(
      path.join(workspaceRoot(), 'src'),
      '_fixall_plugin_',
    );
    fs.writeFileSync(tmpFile, '// placeholder\n', 'utf-8');

    let doc: vscode.TextDocument | undefined;
    let testError: unknown;
    try {
      const openedDocument = await vscode.workspace.openTextDocument(tmpFile);
      doc = openedDocument;
      const editor = await vscode.window.showTextDocument(openedDocument);
      await withCodeActionsOnSave(
        openedDocument,
        { 'source.fixAll': 'explicit' },
        async () => {
          await editor.edit((b) =>
            b.replace(
              new vscode.Range(
                openedDocument.positionAt(0),
                openedDocument.positionAt(openedDocument.getText().length),
              ),
              'const numbers = [1, 2, 3];\nconst ok = numbers.filter((n) => n > 0).length > 0;\nexport { ok };\n',
            ),
          );

          const diagnostics = await waitForRslintDiagnostics(
            openedDocument,
            (diags) =>
              diags.some((d) => d.message.includes('local/prefer-array-some')),
          );
          assert.ok(
            diagnostics.some((d) =>
              d.message.includes('local/prefer-array-some'),
            ),
            `prefer-array-some did not appear; cannot exercise fixAll. Got: ${messages(diagnostics)}`,
          );

          await saveDocumentOnce(
            openedDocument,
            'Document should complete the plugin code-action-on-save pipeline',
          );
          await waitForContentChange(
            openedDocument,
            (content) => !content.includes('.filter('),
            60000,
          );

          assert.ok(
            !openedDocument.getText().includes('.filter('),
            `Plugin fix (filter -> some) should apply via source.fixAll on save.\nContent: ${openedDocument.getText()}`,
          );
          assert.ok(
            openedDocument.getText().includes('.some('),
            `Expected '.some(' after fixAll.\nContent: ${openedDocument.getText()}`,
          );
        },
      );
    } catch (error) {
      testError = error;
    }

    const errors: unknown[] = [];
    if (testError) errors.push(testError);
    try {
      await closeAndDeleteTemporaryDocument(doc, tmpFile);
    } catch (error) {
      errors.push(error);
    }
    if (errors.length === 1) throw errors[0];
    if (errors.length > 1) {
      throw new AggregateError(
        errors,
        'Plugin on-save test and temporary-file cleanup failed',
      );
    }
  });

  test('mounted plugin rules report, merged with native rules', async () => {
    const filePath = path.join(workspaceRoot(), 'src', 'index.ts');
    const doc = await vscode.workspace.openTextDocument(filePath);
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForRslintDiagnostics(doc, (diags) =>
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
});
