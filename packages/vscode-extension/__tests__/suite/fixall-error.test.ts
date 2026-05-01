import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  findFixAllAction,
  prewarmOnSaveFixAll,
  requestFixAll,
  withTmpFile,
  withOnSaveFixAll,
  replaceAll,
} from './fixall-helpers';

suite('rslint fixAll - error flows', function () {
  this.timeout(120000);

  // Prime the on-save fixAll pipeline once before the first test that
  // exercises it. The helper is process-wide idempotent — if another
  // suite has already warmed it, this resolves immediately.
  suiteSetup(async function () {
    this.timeout(120000);
    await prewarmOnSaveFixAll();
  });

  test('fixAll on file with syntax errors does not crash', async () => {
    const brokenContent = 'const x: string = \nfunction (\nexport { \n';
    await withTmpFile(brokenContent, async (doc) => {
      await new Promise((r) => setTimeout(r, 3000));

      const codeActions = await requestFixAll(doc);
      const fixAllAction = findFixAllAction(codeActions);

      if (fixAllAction?.edit) {
        const edits = fixAllAction.edit.get(doc.uri);
        if (edits && edits.length > 0) {
          const applied = await vscode.workspace.applyEdit(fixAllAction.edit);
          assert.ok(applied, 'Edit from fixAll on broken file should apply');
        }
      }
    });
  });

  test('fixAll on empty file does not crash', async () => {
    await withTmpFile('', async (doc) => {
      await new Promise((r) => setTimeout(r, 2000));

      const codeActions = await requestFixAll(doc);
      const fixAllAction = findFixAllAction(codeActions);

      if (fixAllAction?.edit) {
        const edits = fixAllAction.edit?.get(doc.uri);
        assert.ok(
          !edits || edits.length === 0,
          'fixAll should not produce edits for empty file',
        );
      }
    });
  });

  test('on-save with syntax errors saves normally', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      await replaceAll(
        editor,
        "const pVal: string = 'x';\nconst pRes = (pVal as string).trim();\n",
      );
      const probeDiags = await waitForDiagnostics(doc);
      if (
        !probeDiags.some((d) =>
          d.message.includes('no-unnecessary-type-assertion'),
        )
      ) {
        return;
      }
      await doc.save();
      const probeStart = Date.now();
      while (
        doc.getText().includes('pVal as string') &&
        Date.now() - probeStart < 20000
      ) {
        await new Promise((r) => setTimeout(r, 500));
      }

      const brokenContent = 'const x = \nfunction {\nexport {\n';
      await replaceAll(editor, brokenContent);
      await doc.save();

      await new Promise((r) => setTimeout(r, 3000));

      assert.ok(
        doc.getText().length > 0,
        'Document should have content after save',
      );
    });
  });
});
