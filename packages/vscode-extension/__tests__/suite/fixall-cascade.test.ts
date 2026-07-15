import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  waitForContentChange,
  findFixAllAction,
  requestFixAll,
  withTmpFile,
  withOnSaveFixAll,
  replaceAll,
  saveDocumentOnce,
} from './fixall-helpers';

suite('rslint fixAll - cascade (multi-pass)', function () {
  this.timeout(120000);

  test('no-wrapper-object-types triggers no-inferrable-types in second pass', async () => {
    const cascadeContent = [
      "const csA: String = 'hello';",
      'const csB: Number = 42;',
      'const csC: Boolean = true;',
      'export { csA, csB, csC };',
      '',
    ].join('\n');
    await withTmpFile(cascadeContent, async (doc) => {
      const initialDiags = await waitForDiagnostics(doc);
      const wrapperDiags = initialDiags.filter((d) =>
        d.message.includes('no-wrapper-object-types'),
      );
      assert.ok(
        wrapperDiags.length > 0,
        `Expected no-wrapper-object-types diagnostics. Got: ${initialDiags
          .map((d) => d.message)
          .join(' | ')}`,
      );

      const fixAllAction = findFixAllAction(await requestFixAll(doc));
      assert.ok(fixAllAction?.edit, 'Cascade fixAll should provide an edit');

      assert.ok(
        await vscode.workspace.applyEdit(fixAllAction.edit),
        'Cascade fixAll edit should apply',
      );

      const fixedContent = doc.getText();
      assert.ok(
        !fixedContent.includes(': String') &&
          !fixedContent.includes(': Number') &&
          !fixedContent.includes(': Boolean'),
        `no-wrapper-object-types should be fixed. Content: ${fixedContent}`,
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
      assert.ok(
        diags.some((d) => d.message.includes('no-wrapper-object-types')),
        `Expected no-wrapper-object-types before on-save cascade. Got: ${diags
          .map((d) => d.message)
          .join(' | ')}`,
      );

      await saveDocumentOnce(doc, 'Cascade document should save');

      // Event-driven wait: resolves the moment the on-save fixAll edit
      // lands on the document, instead of polling on a 500ms interval.
      // 60s budget gives Windows runners headroom under load.
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
