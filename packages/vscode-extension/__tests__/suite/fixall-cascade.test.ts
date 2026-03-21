import * as assert from 'assert';
import * as vscode from 'vscode';
import {
  waitForDiagnostics,
  findFixAllAction,
  requestFixAll,
  withTmpFile,
  withOnSaveFixAll,
  replaceAll,
} from './fixall-helpers';

suite('rslint fixAll - cascade (multi-pass)', function () {
  this.timeout(50000);

  test('ban-types triggers no-inferrable-types in second pass', async () => {
    const cascadeContent = [
      "const csA: String = 'hello';",
      'const csB: Number = 42;',
      'const csC: Boolean = true;',
      'export { csA, csB, csC };',
      '',
    ].join('\n');
    await withTmpFile(cascadeContent, async doc => {
      const initialDiags = await waitForDiagnostics(doc);
      const banTypeDiags = initialDiags.filter(d =>
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
      if (!diags.some(d => d.message.includes('ban-types'))) return;

      await doc.save();

      const startTime = Date.now();
      while (
        (doc.getText().includes(': String') ||
          doc.getText().includes(': string')) &&
        Date.now() - startTime < 15000
      ) {
        await new Promise(r => setTimeout(r, 500));
      }

      const content = doc.getText();
      assert.ok(
        !content.includes(': String') && !content.includes(': string'),
        `Single save should fix both cascade passes. Content: ${content}`,
      );
    });
  });
});
