import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  waitForContentChange,
  findFixAllAction,
  prewarmOnSaveFixAll,
  requestFixAll,
  withTmpFile,
  withOnSaveFixAll,
  replaceAll,
} from './fixall-helpers';

suite('rslint fixAll - cascade (multi-pass)', function () {
  this.timeout(120000);

  // Prime the on-save fixAll pipeline once so the first real test below
  // doesn't pay VS Code's codeActionsOnSave + LSP cold-start cost (~30s on
  // Windows under load). See fixall-helpers.ts:prewarmOnSaveFixAll.
  suiteSetup(async function () {
    this.timeout(120000);
    await prewarmOnSaveFixAll();
  });

  test('ban-types triggers no-inferrable-types in second pass', async () => {
    const cascadeContent = [
      "const csA: String = 'hello';",
      'const csB: Number = 42;',
      'const csC: Boolean = true;',
      'export { csA, csB, csC };',
      '',
    ].join('\n');
    await withTmpFile(cascadeContent, async (doc) => {
      const initialDiags = await waitForDiagnostics(doc);
      const banTypeDiags = initialDiags.filter((d) =>
        d.message.includes('ban-types'),
      );
      if (banTypeDiags.length === 0) return;

      const fixAllAction = findFixAllAction(await requestFixAll(doc));
      if (!fixAllAction?.edit) return;

      await vscode.workspace.applyEdit(fixAllAction.edit);

      const fixedContent = doc.getText();
      assert.ok(
        !fixedContent.includes(': String') &&
          !fixedContent.includes(': Number') &&
          !fixedContent.includes(': Boolean'),
        `ban-types should be fixed. Content: ${fixedContent}`,
      );
      assert.ok(
        !fixedContent.includes(': string') &&
          !fixedContent.includes(': number') &&
          !fixedContent.includes(': boolean'),
        `no-inferrable-types should also be fixed (cascade). Content: ${fixedContent}`,
      );
    });
  });

  test('cascade on-save - single save fixes both passes', async () => {
    await withOnSaveFixAll(async (doc, editor) => {
      const cascadeContent = [
        "const osA: String = 'world';",
        'const osB: Number = 99;',
        'export { osA, osB };',
        '',
      ].join('\n');
      await replaceAll(editor, cascadeContent);

      const diags = await waitForDiagnostics(doc);
      if (!diags.some((d) => d.message.includes('ban-types'))) return;

      await doc.save();

      // Event-driven wait: resolves the moment the on-save fixAll edit
      // lands on the document, instead of polling on a 500ms interval.
      // 60s budget gives Windows runners headroom even after pre-warm.
      // The helper rejects with a descriptive timeout error including the
      // last seen document content; let that propagate verbatim so the
      // original stack survives.
      await waitForContentChange(
        doc,
        (content) =>
          !content.includes(': String') && !content.includes(': string'),
        60000,
      );
    });
  });
});
