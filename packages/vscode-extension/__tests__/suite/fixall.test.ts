import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  waitForDiagnosticsToChange,
  waitForDiagnosticsCount,
  openFixture,
  findFixAllAction,
  requestFixAll,
  withTmpFile,
} from './fixall-helpers';

suite('rslint fixAll - code actions', function () {
  this.timeout(50000);

  // ======== Basic fixAll behavior (read-only, safe to use fixtures) ========

  test('returns fixes for auto-fixable file', async () => {
    const doc = await openFixture('fixall.ts');
    await vscode.window.showTextDocument(doc);

    await waitForDiagnostics(doc);

    const fixAllAction = findFixAllAction(await requestFixAll(doc));

    assert.ok(fixAllAction, 'Should have a fixAll code action');
    assert.ok(fixAllAction.edit, 'fixAll action should have an edit');

    const edits = fixAllAction.edit.get(doc.uri);
    assert.ok(
      edits && edits.length > 0,
      `fixAll should produce edits, got ${edits?.length ?? 0}`,
    );
  });

  test('no action for non-TS file', async () => {
    const doc = await vscode.workspace.openTextDocument(
      require('node:path').resolve(
        require.resolve('@rslint/core'),
        '../..',
        'fixtures/src/',
        'styles.css',
      ),
    );
    await vscode.window.showTextDocument(doc);

    await new Promise((r) => setTimeout(r, 2000));

    const fixAllAction = findFixAllAction(await requestFixAll(doc));

    if (fixAllAction) {
      const edits = fixAllAction.edit?.get(doc.uri);
      assert.ok(
        !edits || edits.length === 0,
        'fixAll should not produce edits for non-TS file',
      );
    }
  });

  test('no fixes for file with only non-fixable diagnostics', async () => {
    const doc = await openFixture('disable.ts');
    await vscode.window.showTextDocument(doc);

    const diagnostics = await waitForDiagnostics(doc);
    assert.ok(diagnostics.length > 0, 'Should have diagnostics');

    const fixAllAction = findFixAllAction(await requestFixAll(doc));

    if (fixAllAction) {
      const edits = fixAllAction.edit?.get(doc.uri);
      assert.ok(
        !edits || edits.length === 0,
        'fixAll should not produce edits for non-fixable diagnostics',
      );
    }
  });

  // ======== Tests that modify content (use tmp files) ========

  test('no action for clean file', async () => {
    const cleanContent = '// no lint errors\nexport {};\n';
    await withTmpFile(cleanContent, async (doc) => {
      await waitForDiagnosticsCount(doc, 0);

      const fixAllAction = findFixAllAction(await requestFixAll(doc));

      if (fixAllAction) {
        const edits = fixAllAction.edit?.get(doc.uri);
        assert.ok(
          !edits || edits.length === 0,
          'fixAll should not produce edits for clean file',
        );
      }
    });
  });

  test('fixes reduce diagnostics after apply', async () => {
    const fixableContent =
      "const frVal: string = 'hello';\nconst frRes = (frVal as string).toUpperCase();\n";
    await withTmpFile(fixableContent, async (doc) => {
      const initialDiags = await waitForDiagnostics(doc);
      assert.ok(initialDiags.length > 0, 'Should have initial diagnostics');

      const fixableDiags = initialDiags.filter((d) =>
        d.message.includes('no-unnecessary-type-assertion'),
      );
      if (fixableDiags.length === 0) return;

      const fixAllAction = findFixAllAction(await requestFixAll(doc));

      if (fixAllAction?.edit) {
        const applied = await vscode.workspace.applyEdit(fixAllAction.edit);
        assert.ok(applied, 'fixAll edit should apply successfully');

        const updatedDiags = await waitForDiagnosticsToChange(
          doc,
          initialDiags.length,
        );

        assert.ok(
          updatedDiags.length < initialDiags.length,
          `Diagnostics should decrease after fixAll. Before: ${initialDiags.length}, After: ${updatedDiags.length}`,
        );
      }
    });
  });

  test('mixed fixable and non-fixable - only fixes fixable', async () => {
    const mixedContent = [
      "const mfVal: string = 'hello';",
      'const mfRes = (mfVal as string).toUpperCase();',
      'const mfUnsafe: any = {};',
      'mfUnsafe.foo;',
      '',
    ].join('\n');
    await withTmpFile(mixedContent, async (doc) => {
      const initialDiags = await waitForDiagnostics(doc);
      assert.ok(initialDiags.length > 0, 'Should have diagnostics');

      const fixableBefore = initialDiags.filter((d) =>
        d.message.includes('no-unnecessary-type-assertion'),
      );
      const nonFixableBefore = initialDiags.filter((d) =>
        d.message.includes('no-unsafe'),
      );
      if (fixableBefore.length === 0 || nonFixableBefore.length === 0) return;

      const fixAllAction = findFixAllAction(await requestFixAll(doc));

      if (fixAllAction?.edit) {
        await vscode.workspace.applyEdit(fixAllAction.edit);

        const updatedDiags = await waitForDiagnosticsToChange(
          doc,
          initialDiags.length,
        );

        const fixableAfter = updatedDiags.filter((d) =>
          d.message.includes('no-unnecessary-type-assertion'),
        );
        assert.ok(
          fixableAfter.length < fixableBefore.length,
          `Fixable diagnostics should decrease. Before: ${fixableBefore.length}, After: ${fixableAfter.length}`,
        );

        const nonFixableAfter = updatedDiags.filter((d) =>
          d.message.includes('no-unsafe'),
        );
        assert.ok(
          nonFixableAfter.length > 0,
          `Non-fixable diagnostics should remain after fixAll, got ${nonFixableAfter.length}`,
        );
      }
    });
  });

  test('works after rapid edit without debounce', async () => {
    const { replaceAll } = await import('./fixall-helpers');
    await withTmpFile('// initial\nexport {};\n', async (doc, editor) => {
      await new Promise((r) => setTimeout(r, 2000));

      await replaceAll(
        editor,
        "const rapidVal: string = 'test';\nconst rapidResult = (rapidVal as string).trim();\n",
      );

      const fixAllAction = findFixAllAction(await requestFixAll(doc));

      assert.ok(fixAllAction, 'fixAll should work even before debounce fires');
      if (fixAllAction?.edit) {
        const edits = fixAllAction.edit.get(doc.uri);
        assert.ok(
          edits && edits.length > 0,
          'fixAll should produce edits for newly edited content',
        );
      }
    });
  });

  test('second fixAll after first has fewer fixes', async () => {
    const fixableContent =
      "const sfVal: string = 'x';\nconst sfRes = (sfVal as string).trim();\n";
    await withTmpFile(fixableContent, async (doc) => {
      const initialDiags = await waitForDiagnostics(doc);
      const fixableCount = initialDiags.filter((d) =>
        d.message.includes('no-unnecessary-type-assertion'),
      ).length;
      if (fixableCount === 0) return;

      const fixAll1 = findFixAllAction(await requestFixAll(doc));
      assert.ok(fixAll1?.edit, 'First fixAll should have edits');
      await vscode.workspace.applyEdit(fixAll1!.edit!);

      await waitForDiagnosticsToChange(doc, initialDiags.length);

      const fixAll2 = findFixAllAction(await requestFixAll(doc));

      if (fixAll2?.edit) {
        const edits2 = fixAll2.edit.get(doc.uri);
        assert.ok(
          !edits2 || edits2.length === 0,
          `Second fixAll should have no edits, got ${edits2?.length ?? 0}`,
        );
      }
    });
  });
});
